package inner

import (
	"gopkg.in/yaml.v2"
)

// ReadConfig Parses given config content
func ReadConfig(data []byte) (Config, error) {
	c := Config{}
	err := yaml.Unmarshal(data, &c)
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
	Port int
	Path string
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
