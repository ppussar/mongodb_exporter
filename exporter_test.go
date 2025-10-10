package main

import (
	"testing"
	"time"

	"github.com/ppussar/mongodb_exporter/internal"
	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  internal.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: internal.Config{
				HTTP: internal.HTTP{Port: 9090},
				MongoDb: internal.MongoDB{URI: "mongodb://localhost:27017"},
				Metrics: []internal.Metric{{
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
			name: "invalid port",
			config: internal.Config{
				HTTP: internal.HTTP{Port: -1},
				MongoDb: internal.MongoDB{URI: "mongodb://localhost:27017"},
				Metrics: []internal.Metric{{
					Name:             "test_metric",
					Db:               "testdb",
					Collection:       "testcol",
					Find:             "{}",
					MetricsAttribute: "count",
				}},
			},
			wantErr: true,
			errMsg:  "invalid port: -1",
		},
		{
			name: "empty MongoDB URI",
			config: internal.Config{
				HTTP: internal.HTTP{Port: 9090},
				MongoDb: internal.MongoDB{URI: ""},
				Metrics: []internal.Metric{{
					Name:             "test_metric",
					Db:               "testdb",
					Collection:       "testcol",
					Find:             "{}",
					MetricsAttribute: "count",
				}},
			},
			wantErr: true,
			errMsg:  "mongodb URI is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewExporter(t *testing.T) {
	config := internal.Config{
		HTTP: internal.HTTP{Port: 9090},
		MongoDb: internal.MongoDB{URI: "mongodb://localhost:27017"},
		Metrics: []internal.Metric{{
			Name:             "test_metric",
			Db:               "testdb",
			Collection:       "testcol",
			Find:             "{}",
			MetricsAttribute: "count",
		}},
	}

	exporter := NewExporter(config)

	assert.NotNil(t, exporter)
	assert.NotNil(t, exporter.ctx)
	assert.NotNil(t, exporter.cancel)
	assert.Equal(t, config, exporter.config)
	assert.Empty(t, exporter.collectors)
}

func TestExporterShutdown(t *testing.T) {
	config := internal.Config{
		HTTP: internal.HTTP{Port: 0}, // Use port 0 for testing
		MongoDb: internal.MongoDB{URI: "mongodb://localhost:27017"},
		Metrics: []internal.Metric{{
			Name:             "test_metric",
			Db:               "testdb",
			Collection:       "testcol",
			Find:             "{}",
			MetricsAttribute: "count",
		}},
	}

	exporter := NewExporter(config)
	
	// Test context cancellation
	select {
	case <-exporter.ctx.Done():
		t.Fatal("Context should not be cancelled initially")
	default:
	}

	// Test shutdown - just test context cancellation since HTTP server is nil
	exporter.cancel()

	// Context should be cancelled after cancel
	select {
	case <-exporter.ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should be cancelled after cancel")
	}
}
