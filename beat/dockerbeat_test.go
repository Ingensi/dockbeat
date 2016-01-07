package beat

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// helper method
func getEmptyDockerbeat() Dockerbeat {
	return Dockerbeat{
		make(chan struct{}),
		time.Duration(10),
		"/fake/path/to/socket.sock",
		ConfigSettings{
			DockerConfig{nil, nil},
		},
		nil,
		nil,
		EventGenerator{
			map[string]map[string]NetworkData{},
			map[string]BlkioStats{},
		},
		SoftwareVersion{1, 5},
	}
}

func TestDockerbeatValidVersionTooOld(t *testing.T) {
	// GIVEN
	var versions = []string{"1.3.0", "1.4.2", "1.4.9"}
	var beat = getEmptyDockerbeat()

	for _, version := range versions {
		// WHEN
		var valid, err = beat.validVersion(version)

		// THEN
		assert.False(t, valid)
		assert.Nil(t, err)
	}
}

func TestDockerbeatValidVersionMalformed(t *testing.T) {
	// GIVEN
	var versions = []string{"1.xD", "malformed", "1.5-testMalformed"}
	var beat = getEmptyDockerbeat()

	for _, version := range versions {
		// WHEN
		var valid, err = beat.validVersion(version)

		// THEN
		assert.False(t, valid)
		assert.NotNil(t, err)
	}
}

func TestDockerbeatValidVersion(t *testing.T) {
	// GIVEN
	var versions = []string{"1.5.0", "1.5.3", "1.6.12", "1.8.2"}
	var beat = getEmptyDockerbeat()

	for _, version := range versions {
		// WHEN
		var valid, err = beat.validVersion(version)

		// THEN
		assert.True(t, valid)
		assert.Nil(t, err)
	}
}
