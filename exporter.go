package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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
		return errors.New("Missing config")
	}

	dat, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		return err
	}
	config, err := inner.ReadConfig(dat)
	if err != nil {
		return err
	}

	con, err := inner.NewConnection(config.MongoDb.URI)
	if err != nil {
		return err
	}
	registerCollectors(config.Metrics, con)
	serveMetricsEndpoint(config.HTTP.Port, config.HTTP.Path)

	return nil
}

func printUsage() {
	fmt.Printf("Usage: \n\t%s configuration.yaml\n", os.Args[0])
}

func serveMetricsEndpoint(port int, path string) {
	http.Handle(path, promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func registerCollectors(configs []inner.Metric, con inner.Connection) {
	for _, c := range configs {
		collector := inner.NewCollector(c, con)
		prometheus.MustRegister(collector)
	}
}
