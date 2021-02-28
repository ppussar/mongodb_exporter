package internal

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// ReadConfigFile Initializes a Config instance from a given file path
func ReadConfigFile(configFile string) (Config, error) {
	dat, err := ioutil.ReadFile(configFile)
	if err != nil {
		return Config{}, err
	}
	return ReadConfig(dat)
}

// ReadConfig Parses given config content
func ReadConfig(data []byte) (Config, error) {
	c := Config{}
	err := yaml.UnmarshalStrict(data, &c)
	return c, err
}

// Config Root config struct
type Config struct {
	Version string
	HTTP    http
	MongoDb mongoDb
	Metrics []Metric
}
type http struct {
	Port       int
	Prometheus string
	Health     string
	Liveliness string
}
type mongoDb struct {
	URI string
}

// Metric Collector configuration
type Metric struct {
	Name             string
	Help             string
	Db               string
	Collection       string
	Tags             map[string]string
	Find             string
	Aggregate        string
	MetricsAttribute string            `yaml:"metricsAttribute"`
	TagAttributes    map[string]string `yaml:"tagAttributes"`
}
