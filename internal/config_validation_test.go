package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfigStructure(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				HTTP:    HTTP{Port: 9090},
				MongoDb: MongoDB{URI: "mongodb://localhost:27017"},
				Metrics: []Metric{{
					Name:             "test_metric",
					Db:               "testdb",
					Collection:       "testcol",
					Find:             "{}",
					MetricsAttribute: "count",
				}},
			},
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			config: Config{
				HTTP:    HTTP{Port: 0},
				MongoDb: MongoDB{URI: "mongodb://localhost:27017"},
				Metrics: []Metric{{Name: "test", Db: "db", Collection: "col", Find: "{}", MetricsAttribute: "count"}},
			},
			wantErr: true,
			errMsg:  "invalid HTTP port: 0",
		},
		{
			name: "invalid port - too high",
			config: Config{
				HTTP:    HTTP{Port: 70000},
				MongoDb: MongoDB{URI: "mongodb://localhost:27017"},
				Metrics: []Metric{{Name: "test", Db: "db", Collection: "col", Find: "{}", MetricsAttribute: "count"}},
			},
			wantErr: true,
			errMsg:  "invalid HTTP port: 70000",
		},
		{
			name: "empty MongoDB URI",
			config: Config{
				HTTP:    HTTP{Port: 9090},
				MongoDb: MongoDB{URI: ""},
				Metrics: []Metric{{Name: "test", Db: "db", Collection: "col", Find: "{}", MetricsAttribute: "count"}},
			},
			wantErr: true,
			errMsg:  "MongoDB URI cannot be empty",
		},
		{
			name: "no metrics",
			config: Config{
				HTTP:    HTTP{Port: 9090},
				MongoDb: MongoDB{URI: "mongodb://localhost:27017"},
				Metrics: []Metric{},
			},
			wantErr: true,
			errMsg:  "at least one metric must be configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfigStructure(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMetric(t *testing.T) {
	tests := []struct {
		name    string
		metric  Metric
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid metric with find",
			metric: Metric{
				Name:             "valid_metric",
				Db:               "testdb",
				Collection:       "testcol",
				Find:             "{}",
				MetricsAttribute: "count",
			},
			wantErr: false,
		},
		{
			name: "valid metric with aggregate",
			metric: Metric{
				Name:             "valid_metric",
				Db:               "testdb",
				Collection:       "testcol",
				Aggregate:        "[{\"$group\": {\"_id\": null, \"count\": {\"$sum\": 1}}}]",
				MetricsAttribute: "count",
			},
			wantErr: false,
		},
		{
			name: "empty metric name",
			metric: Metric{
				Name:             "",
				Db:               "testdb",
				Collection:       "testcol",
				Find:             "{}",
				MetricsAttribute: "count",
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "invalid prometheus name",
			metric: Metric{
				Name:             "123invalid",
				Db:               "testdb",
				Collection:       "testcol",
				Find:             "{}",
				MetricsAttribute: "count",
			},
			wantErr: true,
			errMsg:  "invalid Prometheus metric name",
		},
		{
			name: "both find and aggregate",
			metric: Metric{
				Name:             "test_metric",
				Db:               "testdb",
				Collection:       "testcol",
				Find:             "{}",
				Aggregate:        "[]",
				MetricsAttribute: "count",
			},
			wantErr: true,
			errMsg:  "cannot specify both 'find' and 'aggregate'",
		},
		{
			name: "neither find nor aggregate",
			metric: Metric{
				Name:             "test_metric",
				Db:               "testdb",
				Collection:       "testcol",
				MetricsAttribute: "count",
			},
			wantErr: true,
			errMsg:  "either 'find' or 'aggregate' query must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMetric(tt.metric, 0)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
