package internal

import (
	"fmt"
	"regexp"
	"strings"
	
	"gopkg.in/yaml.v2"
	"os"
)

var prometheusNameRegex = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)

// ReadConfigFile Initializes a Config instance from a given file path
func ReadConfigFile(configFile string) (Config, error) {
	dat, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}
	return ReadConfig(dat)
}

// ReadConfig Parses given config content
func ReadConfig(data []byte) (Config, error) {
	c := Config{}
	if err := yaml.UnmarshalStrict(data, &c); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}
	
	if err := validateConfigStructure(c); err != nil {
		return Config{}, fmt.Errorf("config validation failed: %w", err)
	}
	
	return c, nil
}

func validateConfigStructure(c Config) error {
	// Validate HTTP config
	if c.HTTP.Port <= 0 || c.HTTP.Port > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", c.HTTP.Port)
	}
	
	// Validate MongoDB config
	if strings.TrimSpace(c.MongoDb.URI) == "" {
		return fmt.Errorf("MongoDB URI cannot be empty")
	}
	
	// Validate metrics
	if len(c.Metrics) == 0 {
		return fmt.Errorf("at least one metric must be configured")
	}
	
	for i, metric := range c.Metrics {
		if err := validateMetric(metric, i); err != nil {
			return err
		}
	}
	
	return nil
}

func validateMetric(m Metric, index int) error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("metric[%d]: name cannot be empty", index)
	}
	
	if !prometheusNameRegex.MatchString(m.Name) {
		return fmt.Errorf("metric[%d]: invalid Prometheus metric name '%s'", index, m.Name)
	}
	
	if strings.TrimSpace(m.Db) == "" {
		return fmt.Errorf("metric[%d]: database name cannot be empty", index)
	}
	
	if strings.TrimSpace(m.Collection) == "" {
		return fmt.Errorf("metric[%d]: collection name cannot be empty", index)
	}
	
	if strings.TrimSpace(m.Find) == "" && strings.TrimSpace(m.Aggregate) == "" {
		return fmt.Errorf("metric[%d]: either 'find' or 'aggregate' query must be specified", index)
	}
	
	if strings.TrimSpace(m.Find) != "" && strings.TrimSpace(m.Aggregate) != "" {
		return fmt.Errorf("metric[%d]: cannot specify both 'find' and 'aggregate' queries", index)
	}
	
	if strings.TrimSpace(m.MetricsAttribute) == "" {
		return fmt.Errorf("metric[%d]: metricsAttribute cannot be empty", index)
	}
	
	return nil
}

// Config Root config struct
type Config struct {
	Version string  `yaml:"version"`
	HTTP    HTTP    `yaml:"http"`
	MongoDb MongoDB `yaml:"mongodb"`
	Metrics []Metric `yaml:"metrics"`
}

type HTTP struct {
	Port       int    `yaml:"port"`
	Prometheus string `yaml:"prometheus"`
	Health     string `yaml:"health"`
	Liveliness string `yaml:"liveliness"`
}

type MongoDB struct {
	URI string `yaml:"uri"`
}

// Metric Collector configuration
type Metric struct {
	Name             string            `yaml:"name"`
	Help             string            `yaml:"help"`
	Db               string            `yaml:"db"`
	Collection       string            `yaml:"collection"`
	Tags             map[string]string `yaml:"tags"`
	Find             string            `yaml:"find"`
	Aggregate        string            `yaml:"aggregate"`
	MetricsAttribute string            `yaml:"metricsAttribute"`
	TagAttributes    map[string]string `yaml:"tagAttributes"`
}
