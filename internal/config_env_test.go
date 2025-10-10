package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvOverrides(t *testing.T) {
	// Save original env vars
	originalVars := map[string]string{
		"HTTP_PORT":        os.Getenv("HTTP_PORT"),
		"HTTP_PROMETHEUS":  os.Getenv("HTTP_PROMETHEUS"),
		"MONGODB_URI":      os.Getenv("MONGODB_URI"),
	}
	
	// Clean up after test
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, value)
			}
		}
	}()

	// Set test environment variables
	_ = os.Setenv("HTTP_PORT", "8080")
	_ = os.Setenv("HTTP_PROMETHEUS", "/custom-metrics")
	_ = os.Setenv("MONGODB_URI", "mongodb://env-host:27017")

	yaml := `
version: 1.0
http:
  port: 9090
  prometheus: /prometheus
mongodb:
  uri: mongodb://localhost:27017
metrics:
  - name: test_metric
    db: testdb
    collection: testcol
    find: '{}'
    metricsAttribute: count
`

	config, err := ReadConfig([]byte(yaml))
	
	assert.NoError(t, err)
	assert.Equal(t, 8080, config.HTTP.Port, "HTTP_PORT should override config")
	assert.Equal(t, "/custom-metrics", config.HTTP.Prometheus, "HTTP_PROMETHEUS should override config")
	assert.Equal(t, "mongodb://env-host:27017", config.MongoDb.URI, "MONGODB_URI should override config")
}

func TestEnvOverridesPartial(t *testing.T) {
	// Only set one env var
	_ = os.Setenv("HTTP_PORT", "3000")
	defer func() { _ = os.Unsetenv("HTTP_PORT") }()

	yaml := `
version: 1.0
http:
  port: 9090
  prometheus: /prometheus
mongodb:
  uri: mongodb://localhost:27017
metrics:
  - name: test_metric
    db: testdb
    collection: testcol
    find: '{}'
    metricsAttribute: count
`

	config, err := ReadConfig([]byte(yaml))
	
	assert.NoError(t, err)
	assert.Equal(t, 3000, config.HTTP.Port, "HTTP_PORT should override config")
	assert.Equal(t, "/prometheus", config.HTTP.Prometheus, "Should keep original value")
	assert.Equal(t, "mongodb://localhost:27017", config.MongoDb.URI, "Should keep original value")
}

func TestEnvOverridesInvalidPort(t *testing.T) {
	// Set invalid port
	_ = os.Setenv("HTTP_PORT", "invalid")
	defer func() { _ = os.Unsetenv("HTTP_PORT") }()

	yaml := `
version: 1.0
http:
  port: 9090
mongodb:
  uri: mongodb://localhost:27017
metrics:
  - name: test_metric
    db: testdb
    collection: testcol
    find: '{}'
    metricsAttribute: count
`

	config, err := ReadConfig([]byte(yaml))
	
	assert.NoError(t, err)
	assert.Equal(t, 9090, config.HTTP.Port, "Should keep original value for invalid env var")
}
