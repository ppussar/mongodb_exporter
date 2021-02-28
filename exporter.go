package main

import (
	"context"
	"fmt"
	"github.com/ppussar/mongodb_exporter/internal/logger"
	"github.com/ppussar/mongodb_exporter/internal/wrapper"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"sync"
	"time"

	"github.com/ppussar/mongodb_exporter/internal"
)

var log = logger.GetInstance()

// An Exporter queries a mongodb to gather metrics and provide those on a prometheus http endpoint
type Exporter struct {
	srv        *internal.HttpServer
	config     internal.Config
	collectors []*internal.Collector
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	config, err := internal.ReadConfigFile(os.Args[1])
	if err != nil {
		log.Fatal(err.Error())
	}
	NewExporter(config).start()
}

func printUsage() {
	fmt.Printf("Usage: \n\t%s configuration.yaml\n", os.Args[0])
}

// NewExporter creates a new Exporter defined by the given config
func NewExporter(config internal.Config) *Exporter {
	return &Exporter{
		config:     config,
		srv:        internal.NewHttpServer(config),
		collectors: make([]*internal.Collector, 0),
	}
}

func (e *Exporter) start() {
	go e.connect()

	wg := &sync.WaitGroup{}
	log.Info("Started")
	wg.Add(1)
	e.srv.Start(wg)
	wg.Wait()
}

func (e *Exporter) connect() {
	errorC := make(chan error, 1)

	for {
		con, err := internal.NewConnection(e.config.MongoDb.URI)
		if err != nil {
			log.Info(fmt.Sprintf("Error during connection creation: %v; Retry in 2s...", err))
			time.Sleep(2 * time.Second)
			continue
		}
		if con != nil {
			if len(e.collectors) == 0 {
				e.registerCollectors(e.config.Metrics, con, errorC)
			} else {
				e.updateCollectorConnection(con)
			}
		}
		<-errorC
	}
}

func (e *Exporter) shutdown(ctx context.Context) error {
	return e.srv.Shutdown(ctx)
}

func (e *Exporter) registerCollectors(configs []internal.Metric, con wrapper.IConnection, errorC chan error) {
	//e.collectors = make([]*internal.Collector, len(configs))
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
		curCollector.Mongo = con
	}
}
