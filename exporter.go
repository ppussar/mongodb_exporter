package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/ppussar/mongodb_exporter/internal/logger"
	"github.com/ppussar/mongodb_exporter/internal/wrapper"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ppussar/mongodb_exporter/internal"
)

var log = logger.GetInstance()

// An Exporter queries a mongodb to gather metrics and provide those on a prometheus http endpoint
type Exporter struct {
	srv        *internal.HttpServer
	config     internal.Config
	collectors []*internal.Collector
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	
	config, err := internal.ReadConfigFile(os.Args[1])
	if err != nil {
		log.Error(fmt.Sprintf("Failed to read config: %v", err))
		os.Exit(1)
	}

	exporter := NewExporter(config)
	handleSignals(exporter)
	
	if err := exporter.start(); err != nil {
		log.Error(fmt.Sprintf("Failed to start exporter: %v", err))
		os.Exit(1)
	}
}

func handleSignals(exporter *Exporter) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		log.Info("Received shutdown signal")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := exporter.shutdown(ctx); err != nil {
			log.Error(fmt.Sprintf("Shutdown error: %v", err))
		}
		os.Exit(0)
	}()
}

func printUsage() {
	fmt.Printf("Usage: \n\t%s configuration.yaml\n", os.Args[0])
}

// NewExporter creates a new Exporter defined by the given config
func NewExporter(config internal.Config) *Exporter {
	ctx, cancel := context.WithCancel(context.Background())
	return &Exporter{
		config:     config,
		srv:        internal.NewHttpServer(config),
		collectors: make([]*internal.Collector, 0),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (e *Exporter) start() error {
	go e.connect()

	wg := &sync.WaitGroup{}
	log.Info("Started")
	wg.Add(1)
	e.srv.Start(wg)
	wg.Wait()
	return nil
}

func (e *Exporter) connect() {
	errorC := make(chan error, 10) // Buffered to prevent blocking

	for {
		select {
		case <-e.ctx.Done():
			return
		default:
		}

		con, err := internal.NewConnection(e.config.MongoDb.URI)
		safeURI := internal.MaskURI(e.config.MongoDb.URI)
		if err != nil {
			internal.ConnectionStatus.WithLabelValues(safeURI).Set(0)
			log.Info(fmt.Sprintf("Error during connection creation: %v; Retry in 2s...", err))
			select {
			case <-time.After(2 * time.Second):
			case <-e.ctx.Done():
				return
			}
			continue
		}

		if con != nil {
			internal.ConnectionStatus.WithLabelValues(safeURI).Set(1)
			e.mu.Lock()
			if len(e.collectors) == 0 {
				e.registerCollectors(e.config.Metrics, con, errorC)
			} else {
				e.updateCollectorConnection(con)
			}
			e.mu.Unlock()
		}

		// Fix 6: drain collector errors but only reconnect when a
		// connection-level error occurs, not on every query error.
		for {
			select {
			case err := <-errorC:
				log.Error(fmt.Sprintf("Collector error: %v", err))
				// Only reconnect if it looks like a connection problem
				if isConnectionError(err) {
					internal.ConnectionStatus.WithLabelValues(safeURI).Set(0)
					goto reconnect
				}
			case <-e.ctx.Done():
				return
			}
		}
	reconnect:
	}
}

func (e *Exporter) shutdown(ctx context.Context) error {
	e.cancel()
	return e.srv.Shutdown(ctx)
}

func (e *Exporter) registerCollectors(configs []internal.Metric, con wrapper.IConnection, errorC chan error) {
	for _, c := range configs {
		collector := internal.NewCollector(c, con, errorC)
		e.collectors = append(e.collectors, collector)
		log.Info("Register new collector: " + collector.String())
		prometheus.MustRegister(collector)
	}
}

func (e *Exporter) updateCollectorConnection(con wrapper.IConnection) {
	for _, curCollector := range e.collectors {
		log.Info("Update connection in collector: " + curCollector.String())
		curCollector.UpdateConnection(con)
	}
}

// isConnectionError returns true for errors that indicate a lost MongoDB connection.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no mongodb connection") ||
		strings.Contains(msg, "connection") ||
		strings.Contains(msg, "socket") ||
		strings.Contains(msg, "eof") ||
		strings.Contains(msg, "broken pipe")
}
