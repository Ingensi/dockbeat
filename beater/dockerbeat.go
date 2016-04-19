package beater

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"fmt"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/fsouza/go-dockerclient"

	"github.com/ingensi/dockerbeat/calculator"
	"github.com/ingensi/dockerbeat/config"
	"github.com/ingensi/dockerbeat/event"
)

// const for event logs
const (
	ERROR = "error"
	WARN  = "warning"
	INFO  = "info"
	DEBUG = "debug"
	TRACE = "trace"
)

type SoftwareVersion struct {
	major int
	minor int
}

type SocketConfig struct {
	socket    string
	enableTls bool
	caPath    string
	certPath  string
	keyPath   string
}

type Dockerbeat struct {
	done                 chan struct{}
	period               time.Duration
	socketConfig         SocketConfig
	beatConfig           *config.Config
	dockerClient         *docker.Client
	events               publisher.Client
	eventGenerator       event.EventGenerator
	minimalDockerVersion SoftwareVersion
}

// Creates beater
func New() *Dockerbeat {
	return &Dockerbeat{}
}

/// *** Beater interface methods ***///

func (bt *Dockerbeat) Config(b *beat.Beat) error {

	// Requires Docker 1.9 minimum
	bt.minimalDockerVersion = SoftwareVersion{major: 1, minor: 9}

	err := cfgfile.Read(&bt.beatConfig, "")
	if err != nil {
		logp.Err("Error reading configuration file: %v", err)
		return err
	}

	//init the period
	if bt.beatConfig.Dockerbeat.Period != nil {
		bt.period = time.Duration(*bt.beatConfig.Dockerbeat.Period) * time.Second
	} else {
		bt.period = 1 * time.Second
	}
	//init the socketConfig
	bt.socketConfig = SocketConfig{
		socket:    "",
		enableTls: false,
		caPath:    "",
		certPath:  "",
		keyPath:   "",
	}

	if bt.beatConfig.Dockerbeat.Socket != nil {
		bt.socketConfig.socket = *bt.beatConfig.Dockerbeat.Socket
	} else {
		bt.socketConfig.socket = "unix:///var/run/docker.sock" // default docker socket location
	}
	if bt.beatConfig.Dockerbeat.Tls.Enable != nil {
		bt.socketConfig.enableTls = *bt.beatConfig.Dockerbeat.Tls.Enable
	} else {
		bt.socketConfig.enableTls = false
	}
	if bt.socketConfig.enableTls {
		if bt.beatConfig.Dockerbeat.Tls.CaPath != nil {
			bt.socketConfig.caPath = *bt.beatConfig.Dockerbeat.Tls.CaPath
		}
		if bt.beatConfig.Dockerbeat.Tls.CertPath != nil {
			bt.socketConfig.certPath = *bt.beatConfig.Dockerbeat.Tls.CertPath
		}
		if bt.beatConfig.Dockerbeat.Tls.KeyPath != nil {
			bt.socketConfig.keyPath = *bt.beatConfig.Dockerbeat.Tls.KeyPath
		}
	}

	logp.Info("dockerbeat", "Init dockerbeat")
	logp.Info("dockerbeat", "Follow docker socket %q\n", bt.socketConfig.socket)
	if bt.socketConfig.enableTls {
		logp.Info("dockerbeat", "TLS enabled\n")
	} else {
		logp.Info("dockerbeat", "TLS disabled\n")
	}
	logp.Info("dockerbeat", "Period %v\n", bt.period)

	return nil
}

func (bt *Dockerbeat) getDockerClient() (*docker.Client, error) {
	var client *docker.Client
	var err error

	if bt.socketConfig.enableTls {
		client, err = docker.NewTLSClient(
			bt.socketConfig.socket,
			bt.socketConfig.certPath,
			bt.socketConfig.keyPath,
			bt.socketConfig.caPath,
		)
	} else {
		client, err = docker.NewClient(bt.socketConfig.socket)
	}
	return client, err
}

func (bt *Dockerbeat) Setup(b *beat.Beat) error {
	var clientErr error
	var err error
	//populate Dockerbeat
	bt.events = b.Events
	bt.done = make(chan struct{})
	bt.dockerClient, clientErr = bt.getDockerClient()
	bt.eventGenerator = event.EventGenerator{
		Socket:            &bt.socketConfig.socket,
		NetworkStats:      event.EGNetworkStats{M: map[string]map[string]calculator.NetworkData{}},
		BlkioStats:        event.EGBlkioStats{M: map[string]calculator.BlkioData{}},
		CalculatorFactory: calculator.CalculatorFactoryImpl{},
		Period:            bt.period,
	}

	if clientErr != nil {
		err = errors.New(fmt.Sprintf("Unable to create docker client, please check your docker socket/TLS settings: %v", clientErr))
	}
	return err
}

func (bt *Dockerbeat) Run(b *beat.Beat) error {
	logp.Info("dockerbeat is running! Hit CTRL-C to stop it.")
	var err error

	ticker := time.NewTicker(bt.period)
	defer ticker.Stop()

	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		// check prerequisites
		err = bt.checkPrerequisites()
		if err != nil {
			logp.Err("Unable to collect metrics: %v", err)
			bt.publishLogEvent(ERROR, fmt.Sprintf("Unable to collect metrics: %v", err))
			continue
		}

		timerStart := time.Now()
		bt.RunOneTime(b)
		timerEnd := time.Now()

		duration := timerEnd.Sub(timerStart)
		if duration.Nanoseconds() > bt.period.Nanoseconds() {
			logp.Warn("Ignoring tick(s) due to processing taking longer than one period")
			bt.publishLogEvent(WARN, "Ignoring tick(s) due to processing taking longer than one period")
		}
	}
}

func (d *Dockerbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (d *Dockerbeat) Stop() {
	close(d.done)
	logp.Info("Stopping dockerbeat")
}

func (d *Dockerbeat) RunOneTime(b *beat.Beat) error {
	containers, err := d.dockerClient.ListContainers(docker.ListContainersOptions{})

	if err == nil {
		//export stats for each container
		for _, container := range containers {
			d.exportContainerStats(container)
		}
	} else {
		logp.Err("Cannot get container list: %v", err)
		d.publishLogEvent(ERROR, fmt.Sprintf("Cannot get container list: %v", err))
	}

	d.eventGenerator.CleanOldStats(containers)

	return nil
}

func (d *Dockerbeat) exportContainerStats(container docker.APIContainers) error {
	// statsOptions creation
	statsC := make(chan *docker.Stats)
	done := make(chan bool)
	errC := make(chan error, 1)
	// the stream bool is set to false to only listen the first stats
	statsOptions := docker.StatsOptions{
		ID:      container.ID,
		Stats:   statsC,
		Stream:  false,
		Done:    done,
		Timeout: -1,
	}
	// goroutine to listen to the stats
	go func() {
		errC <- d.dockerClient.Stats(statsOptions)
		close(errC)
	}()
	// goroutine to get the stats & publish it
	go func() {
		stats := <-statsC
		err := <-errC

		if err == nil && stats != nil {
			events := []common.MapStr{
				d.eventGenerator.GetContainerEvent(&container, stats),
				d.eventGenerator.GetCpuEvent(&container, stats),
				d.eventGenerator.GetMemoryEvent(&container, stats),
				d.eventGenerator.GetBlkioEvent(&container, stats),
			}

			events = append(events, d.eventGenerator.GetNetworksEvent(&container, stats)...)

			d.events.PublishEvents(events)
		} else if err == nil && stats == nil {
			logp.Warn("Container was existing at listing but not when getting statistics: %v", container.ID)
			d.publishLogEvent(WARN, fmt.Sprintf("Container was existing at listing but not when getting statistics: %v", container.ID))
		} else {
			logp.Err("An error occurred while getting docker stats: %v", err)
			d.publishLogEvent(ERROR, fmt.Sprintf("An error occurred while getting docker stats: %v", err))
		}
	}()

	return nil
}

func (d *Dockerbeat) checkPrerequisites() error {
	var output error = nil

	env, err := d.dockerClient.Version()

	if err == nil {
		version := env.Get("Version")
		valid, _ := d.validVersion(version)

		if !valid {
			output = errors.New("Docker server is too old (version " +
				strconv.Itoa(d.minimalDockerVersion.major) + "." + strconv.Itoa(d.minimalDockerVersion.minor) + ".x" +
				" and earlier is required)")
		}

	} else {
		output = errors.New("Docker server unreachable: " + err.Error())
	}

	return output
}

func (d *Dockerbeat) validVersion(version string) (bool, error) {

	splitsStr := strings.Split(version, ".")

	if cap(splitsStr) < 2 {
		return false, errors.New("Malformed version")
	}

	actualMajorVersion, err := strconv.Atoi(splitsStr[0])
	if err != nil {
		return false, err
	}
	actualMinorVersion, err := strconv.Atoi(splitsStr[1])
	if err != nil {
		return false, err
	}
	var output bool

	if actualMajorVersion > d.minimalDockerVersion.major ||
		(actualMajorVersion == d.minimalDockerVersion.major && actualMinorVersion >= d.minimalDockerVersion.minor) {
		output = true
	} else {
		output = false
	}
	return output, nil
}

func (d *Dockerbeat) publishLogEvent(level string, message string) {
	event := d.eventGenerator.GetLogEvent(level, message)
	d.events.PublishEvent(event)
}
