package inner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEmptyConfigDoesNotReturnError(t *testing.T) {
	var emptyConfig []byte

	c, err := ReadConfig(emptyConfig)

	assert.Equal(t, err, nil, "No error expected")
	assert.Equal(t, c, Config{}, "Empty Config struct expected")
}

func TestParseNonYamlStringDoesReturnError(t *testing.T) {
	invalidConfig := []byte("This is not an yaml config string")

	_, err := ReadConfig(invalidConfig)

	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `This is...` into inner.Config")
}

func TestParseMinimalConfig(t *testing.T) {
	yaml := ""
	yaml += "version: 1.0\n"
	yaml += "http:\n"
	yaml += "  port: 9090\n"
	minimalConfig := []byte(yaml)

	c, err := ReadConfig(minimalConfig)

	assert.Equal(t, err, nil, "No error expected")
	assert.Equal(t, c.HTTP.Port, 9090)
}

func TestParseInvalidFieldReturnsError(t *testing.T) {
	yaml := ""
	yaml += "version: 1.0\n"
	yaml += "http:\n"
	yaml += "  invalid: 123\n"
	minimalConfig := []byte(yaml)

	_, err := ReadConfig(minimalConfig)

	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 3: field invalid not found in type inner.http")
}

func TestParseInvalidValueReturnsError(t *testing.T) {
	yaml := ""
	yaml += "version: 1.0\n"
	yaml += "http:\n"
	yaml += "  port: abc\n"
	minimalConfig := []byte(yaml)

	_, err := ReadConfig(minimalConfig)

	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 3: cannot unmarshal !!str `abc` into int")
}
