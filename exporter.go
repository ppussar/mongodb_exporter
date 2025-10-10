package main

import (
	"context"
	"fmt"
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
	
	if err := validateConfig(config); err != nil {
		log.Error(fmt.Sprintf("Invalid config: %v", err))
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

func validateConfig(config internal.Config) error {
	if config.HTTP.Port <= 0 || config.HTTP.Port > 65535 {
		return fmt.Errorf("invalid port: %d", config.HTTP.Port)
	}
	if config.MongoDb.URI == "" {
		return fmt.Errorf("mongodb URI is required")
	}
	for i, metric := range config.Metrics {
		if metric.Name == "" {
			return fmt.Errorf("metric[%d]: name is required", i)
		}
		if metric.Db == "" {
			return fmt.Errorf("metric[%d]: db is required", i)
		}
		if metric.Collection == "" {
			return fmt.Errorf("metric[%d]: collection is required", i)
		}
		if metric.Find == "" && metric.Aggregate == "" {
			return fmt.Errorf("metric[%d]: either find or aggregate is required", i)
		}
		if metric.MetricsAttribute == "" {
			return fmt.Errorf("metric[%d]: metricsAttribute is required", i)
		}
	}
	return nil
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
		if err != nil {
			log.Info(fmt.Sprintf("Error during connection creation: %v; Retry in 2s...", err))
			select {
			case <-time.After(2 * time.Second):
			case <-e.ctx.Done():
				return
			}
			continue
		}
		
		if con != nil {
			e.mu.Lock()
			if len(e.collectors) == 0 {
				e.registerCollectors(e.config.Metrics, con, errorC)
			} else {
				e.updateCollectorConnection(con)
			}
			e.mu.Unlock()
		}
		
		select {
		case err := <-errorC:
			log.Error(fmt.Sprintf("Collector error: %v", err))
		case <-e.ctx.Done():
			return
		}
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
