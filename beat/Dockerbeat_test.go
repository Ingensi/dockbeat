package beat

import (
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/publisher"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// DOCKERBEAT TYPE IS CURRENTLY NOT REALLY TESTABLE BECAUSE OF THE DOCKER GO CLIENT WICH DOES NOT DEFINE INTERFACE
// (NOT MOCKABLE)

// SETUP TESTS

func TestDockerbeatSetupMethod(t *testing.T) {
	// GIVEN
	// a dockerbeat instance
	var dockerbeat = getEmptyDockerbeat()
	events := publisher.ChanClient{}
	fakeBeat := beat.Beat{Events: events}

	// WHEN
	dockerbeat.Setup(&fakeBeat)

	// THEN
	// events set as dockerbeat.events
	assert.Equal(t, events, dockerbeat.events)
	// dockerbeat.done initialized
	assert.NotNil(t, dockerbeat.done)
	// dockerClient initialized with given socket
	assert.NotNil(t, dockerbeat.dockerClient)
	assert.Equal(t, dockerbeat.socket, dockerbeat.dockerClient.Endpoint())
	// eventGenerator initialized
	assert.NotNil(t, dockerbeat.eventGenerator)
	assert.NotNil(t, dockerbeat.eventGenerator.blkioStats)
	assert.NotNil(t, dockerbeat.eventGenerator.networkStats)
	assert.NotNil(t, dockerbeat.eventGenerator.calculatorFactory)
	assert.Equal(t, dockerbeat.period, dockerbeat.eventGenerator.period)

}

// CLOSE TESTS

func TestDockebeatCloseMethod(t *testing.T) {
	// GIVEN
	var dockerbeat = getEmptyDockerbeat()

	// WHEN
	dockerbeat.Stop()

	// THEN
	_, ok := <-dockerbeat.done
	if ok {
		assert.Fail(t, "dockerbeat.done not closed")
	}
}

// VALID VERSION TESTS

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

// helper method
func getEmptyDockerbeat() Dockerbeat {
	return Dockerbeat{
		done:   make(chan struct{}),
		period: time.Duration(10),
		socket: "/fake/path/to/socket.sock",
		TbConfig: ConfigSettings{
			Input: DockerConfig{Period: nil, Socket: nil},
		},
		dockerClient: nil,
		events:       nil,
		eventGenerator: EventGenerator{
			networkStats:      map[string]map[string]NetworkData{},
			blkioStats:        map[string]BlkioData{},
			calculatorFactory: CalculatorFactoryImpl{},
			period:            time.Second,
		},
		minimalDockerVersion: SoftwareVersion{major: 1, minor: 5},
	}
}
