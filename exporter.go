package main

import (
	"errors"
	"fmt"
	"github.com/ppussar/mongodb_exporter/inner/wrapper"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/ppussar/mongodb_exporter/inner"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
		return errors.New("missing config")
	}

	dat, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		return err
	}
	config, err := inner.ReadConfig(dat)
	if err != nil {
		return err
	}

	for {
		con, err := inner.NewConnection(config.MongoDb.URI)
		if err != nil {
			fmt.Printf("Waiting %s", err)
			time.Sleep(2 * time.Second)
			continue
		}
		registerCollectors(config.Metrics, con)
		fmt.Println("Started")
		serveMetricsEndpoint(config.HTTP.Port, config.HTTP.Path)
	}
}

func printUsage() {
	fmt.Printf("Usage: \n\t%s configuration.yaml\n", os.Args[0])
}

func serveMetricsEndpoint(port int, path string) {
	http.Handle(path, promhttp.Handler())
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func registerCollectors(configs []inner.Metric, con wrapper.IConnection) {
	for _, c := range configs {
		collector := inner.NewCollector(c, con)
		prometheus.MustRegister(collector)
	}
}
