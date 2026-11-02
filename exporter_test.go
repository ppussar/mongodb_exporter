package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ppussar/mongodb_exporter/internal"
	"github.com/ppussar/mongodb_exporter/internal/mocks"
	"github.com/ppussar/mongodb_exporter/internal/wrapper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResolveConfigPath(t *testing.T) {
	tests := []struct {
		name string
		args []string
		env  map[string]string
		want string
	}{
		{
			name: "cli argument takes precedence",
			args: []string{"/from/arg.yaml"},
			env:  map[string]string{"CONFIG": "/from/env.yaml"},
			want: "/from/arg.yaml",
		},
		{
			name: "falls back to CONFIG env var when no argument",
			args: []string{},
			env:  map[string]string{"CONFIG": "/from/env.yaml"},
			want: "/from/env.yaml",
		},
		{
			name: "empty when neither argument nor env is set",
			args: []string{},
			env:  map[string]string{},
			want: "",
		},
		{
			name: "empty argument falls back to env",
			args: []string{""},
			env:  map[string]string{"CONFIG": "/from/env.yaml"},
			want: "/from/env.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string { return tt.env[key] }
			assert.Equal(t, tt.want, resolveConfigPath(tt.args, getenv))
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			yaml: `
http:
  port: 9090
mongodb:
  uri: mongodb://localhost:27017
metrics:
  - name: test_metric
    db: testdb
    collection: testcol
    find: "{}"
    metricsAttribute: count
`,
			wantErr: false,
		},
		{
			name: "invalid port",
			yaml: `
http:
  port: -1
mongodb:
  uri: mongodb://localhost:27017
metrics:
  - name: test_metric
    db: testdb
    collection: testcol
    find: "{}"
    metricsAttribute: count
`,
			wantErr: true,
			errMsg:  "invalid HTTP port",
		},
		{
			name: "empty MongoDB URI",
			yaml: `
http:
  port: 9090
mongodb:
  uri: ""
metrics:
  - name: test_metric
    db: testdb
    collection: testcol
    find: "{}"
    metricsAttribute: count
`,
			wantErr: true,
			errMsg:  "MongoDB URI cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := internal.ReadConfig([]byte(tt.yaml))
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
		HTTP:    internal.HTTP{Port: 9090},
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
		HTTP:    internal.HTTP{Port: 0}, // Use port 0 for testing
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

func TestConnectDisconnectsPreviousConnectionOnReconnect(t *testing.T) {
	metric := internal.Metric{
		Name:             "reconnect_test_metric",
		Db:               "testdb",
		Collection:       "testcol",
		Find:             "{}",
		MetricsAttribute: "count",
	}
	exporter := NewExporter(internal.Config{
		MongoDb: internal.MongoDB{URI: "mongodb://localhost:27017"},
		Metrics: []internal.Metric{metric},
	})
	defer exporter.cancel()

	connections := make(chan wrapper.IConnection, 2)
	disconnected := make(chan struct{})
	first := &mocks.IConnection{}
	first.EXPECT().Find(mock.Anything, metric.Db, metric.Collection, metric.Find).
		Return(nil, errors.New("connection reset by peer")).Once()
	first.EXPECT().Disconnect(mock.Anything).
		Run(func(ctx context.Context) {
			// the reconnect must have created the new connection already
			assert.Empty(t, connections)
			close(disconnected)
		}).
		Return(nil).Once()
	second := &mocks.IConnection{}
	second.EXPECT().Disconnect(mock.Anything).Return(nil).Maybe()

	connections <- first
	connections <- second
	exporter.newConn = func(string) (wrapper.IConnection, error) {
		return <-connections, nil
	}

	done := make(chan struct{})
	go func() {
		exporter.connect()
		close(done)
	}()

	require.Eventually(t, func() bool {
		exporter.mu.RLock()
		defer exporter.mu.RUnlock()
		return len(exporter.collectors) == 1
	}, 5*time.Second, 10*time.Millisecond)
	exporter.mu.RLock()
	collector := exporter.collectors[0]
	exporter.mu.RUnlock()
	defer prometheus.Unregister(collector)

	// a connection error on the first connection triggers the reconnect
	collector.Collect(make(chan prometheus.Metric, 1))

	select {
	case <-disconnected:
	case <-time.After(5 * time.Second):
		t.Fatal("previous connection was not disconnected after reconnect")
	}
	exporter.cancel()
	<-done

	first.AssertExpectations(t)
	second.AssertNotCalled(t, "Disconnect", mock.Anything)
}
