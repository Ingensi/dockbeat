package beater

import (
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/publisher"
	"github.com/ingensi/dockbeat/calculator"
	"github.com/ingensi/dockbeat/config"
	"github.com/ingensi/dockbeat/event"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// DOCKBEAT TYPE IS CURRENTLY NOT REALLY TESTABLE BECAUSE OF THE DOCKER GO CLIENT WICH DOES NOT DEFINE INTERFACE
// (NOT MOCKABLE)

// SETUP TESTS

func TestDockbeatSetupMethod(t *testing.T) {
	// GIVEN
	// a dockbeat instance
	var dockbeat = getEmptyDockbeat()
	events := publisher.ChanClient{}
	fakeBeat := beat.Beat{Events: events}

	// WHEN
	dockbeat.Setup(&fakeBeat)

	// THEN
	// events set as dockbeat.events
	assert.Equal(t, events, dockbeat.events)
	// dockbeat.done initialized
	assert.NotNil(t, dockbeat.done)
	// dockerClient initialized with given socket
	assert.NotNil(t, dockbeat.dockerClient)
	assert.Equal(t, dockbeat.socketConfig.socket, dockbeat.dockerClient.Endpoint())
	// eventGenerator initialized
	assert.NotNil(t, dockbeat.eventGenerator)
	assert.NotNil(t, dockbeat.eventGenerator.Socket)
	assert.NotNil(t, dockbeat.eventGenerator.BlkioStats)
	assert.NotNil(t, dockbeat.eventGenerator.NetworkStats)
	assert.NotNil(t, dockbeat.eventGenerator.CalculatorFactory)
	assert.Equal(t, dockbeat.period, dockbeat.eventGenerator.Period)
}

// CLOSE TESTS

func TestDockebeatCloseMethod(t *testing.T) {
	// GIVEN
	var dockbeat = getEmptyDockbeat()

	// WHEN
	dockbeat.Stop()

	// THEN
	_, ok := <-dockbeat.done
	if ok {
		assert.Fail(t, "dockbeat.done not closed")
	}
}

// VALID VERSION TESTS

func TestDockbeatValidVersionTooOld(t *testing.T) {
	// GIVEN
	var versions = []string{"1.3.0", "1.4.2", "1.4.9"}
	var beat = getEmptyDockbeat()

	for _, version := range versions {
		// WHEN
		var valid, err = beat.validVersion(version)

		// THEN
		assert.False(t, valid)
		assert.Nil(t, err)
	}
}

func TestDockbeatValidVersionMalformed(t *testing.T) {
	// GIVEN
	var versions = []string{"1.xD", "malformed", "1.5-testMalformed"}
	var beat = getEmptyDockbeat()

	for _, version := range versions {
		// WHEN
		var valid, err = beat.validVersion(version)

		// THEN
		assert.False(t, valid)
		assert.NotNil(t, err)
	}
}

func TestDockbeatValidVersion(t *testing.T) {
	// GIVEN
	var versions = []string{"1.5.0", "1.5.3", "1.6.12", "1.8.2"}
	var beat = getEmptyDockbeat()

	for _, version := range versions {
		// WHEN
		var valid, err = beat.validVersion(version)

		// THEN
		assert.True(t, valid)
		assert.Nil(t, err)
	}
}

// Docker client getter function
func TestDockerClientGetterWithUnixPath(t *testing.T) {
	// GIVEN
	var beat = getEmptyDockbeat()
	socket := "unix:///some/socket/path.sock"
	beat.socketConfig.socket = socket

	// WHEN
	var _, err = beat.getDockerClient()

	// THEN
	assert.Nil(t, err)
	// TODO check if client is initialized according to the given socket
}

func TestDockerClientGetterWithTCPPath(t *testing.T) {
	// GIVEN
	var beat = getEmptyDockbeat()
	socket := "tcp://someHostname:9876"
	beat.socketConfig.socket = socket

	// WHEN
	var _, err = beat.getDockerClient()

	// THEN
	assert.Nil(t, err)
	// TODO check if client is initialized according to the given socket
}

// TODO write test for TLS docker client instantiation

// helper method
func getEmptyDockbeat() Dockbeat {
	return Dockbeat{
		done:   make(chan struct{}),
		period: time.Duration(10),
		socketConfig: SocketConfig{
			socket:    "/fake/path/to/socket.sock",
			enableTls: false,
			caPath:    "",
			certPath:  "",
			keyPath:   "",
		},
		statsConfig: StatsConfig{
			Container: true,
			Cpu:       true,
			Net:       true,
			Blkio:     true,
			Memory:    true,
		},
		beatConfig: &config.Config{
			Dockbeat: config.DockbeatConfig{
				Period: nil,
				Socket: nil,
				Tls: config.TlsConfig{
					Enable:   nil,
					CaPath:   nil,
					CertPath: nil,
					KeyPath:  nil,
				},
				Stats: config.StatsConfig{
					Container: nil,
					Cpu:       nil,
					Net:       nil,
					Blkio:     nil,
					Memory:    nil,
				},
			},
		},
		dockerClient: nil,
		events:       nil,
		eventGenerator: &event.EventGenerator{
			NetworkStats:      event.EGNetworkStats{M: map[string]map[string]calculator.NetworkData{}},
			BlkioStats:        event.EGBlkioStats{M: map[string]calculator.BlkioData{}},
			CalculatorFactory: calculator.CalculatorFactoryImpl{},
			Period:            time.Second,
		},
		minimalDockerVersion: SoftwareVersion{major: 1, minor: 5},
	}
}
