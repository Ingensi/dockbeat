package cfgfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	Output ElasticsearchConfig
}

type ElasticsearchConfig struct {
	Elasticsearch Connection
}

type Connection struct {
	Port int
	Host string
}

func TestRead(t *testing.T) {
	absPath, err := filepath.Abs("../tests/files/")

	assert.NotNil(t, absPath)
	assert.Nil(t, err)

	config := &TestConfig{}

	err = Read(config, absPath+"/config.yml")
	assert.Nil(t, err)

	// Access config
	assert.Equal(t, "localhost", config.Output.Elasticsearch.Host)

	// Chat that it is integer
	assert.Equal(t, 9200, config.Output.Elasticsearch.Port)
}

func TestExpandEnv(t *testing.T) {
	var tests = []struct {
		in  string
		out string
	}{
		// Environment variables can be specified as ${env} only.
		{"${y}", "y"},
		{"$y", "$y"},

		// Environment variables are case-sensitive.
		{"${Y}", ""},

		// Defaults can be specified.
		{"x${Z:D}", "xD"},
		{"x${Z:A B C D}", "xA B C D"}, // Spaces are allowed in the default.
		{"x${Z:}", "x"},

		// Un-matched braces are swallowed by the Go os.Expand function.
		{"x${Y ${Z:Z}", "xZ"},

		// Special environment variables are not replaced.
		{"$*", "$*"},
		{"${*}", ""},
		{"$@", "$@"},
		{"${@}", ""},
		{"$1", "$1"},
		{"${1}", ""},

		{"", ""},
		{"$$", "$$"},
	}

	for _, test := range tests {
		os.Setenv("y", "y")
		output := expandEnv([]byte(test.in))
		assert.Equal(t, test.out, string(output), "Input: %s", test.in)
	}
}
