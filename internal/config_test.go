package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEmptyConfigDoesNotReturnError(t *testing.T) {
	var emptyConfig []byte

	_, err := ReadConfig(emptyConfig)

	assert.Error(t, err, "Error expected due to validation")
	assert.Contains(t, err.Error(), "config validation failed")
}

func TestParseNonYamlStringDoesReturnError(t *testing.T) {
	invalidConfig := []byte("This is not an yaml config string")

	_, err := ReadConfig(invalidConfig)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config")
}

func TestParseMinimalConfig(t *testing.T) {
	yaml := ""
	yaml += "version: 1.0\n"
	yaml += "http:\n"
	yaml += "  port: 9090\n"
	yaml += "mongodb:\n"
	yaml += "  uri: mongodb://localhost:27017\n"
	yaml += "metrics:\n"
	yaml += "  - name: test_metric\n"
	yaml += "    db: testdb\n"
	yaml += "    collection: testcol\n"
	yaml += "    find: '{}'\n"
	yaml += "    metricsAttribute: count\n"
	minimalConfig := []byte(yaml)

	c, err := ReadConfig(minimalConfig)

	assert.NoError(t, err, "No error expected for valid config")
	assert.Equal(t, c.HTTP.Port, 9090)
}

func TestParseInvalidFieldReturnsError(t *testing.T) {
	yaml := ""
	yaml += "version: 1.0\n"
	yaml += "http:\n"
	yaml += "  invalid: 123\n"
	minimalConfig := []byte(yaml)

	_, err := ReadConfig(minimalConfig)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config")
}

func TestParseInvalidValueReturnsError(t *testing.T) {
	yaml := ""
	yaml += "version: 1.0\n"
	yaml += "http:\n"
	yaml += "  port: abc\n"
	minimalConfig := []byte(yaml)

	_, err := ReadConfig(minimalConfig)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config")
}
