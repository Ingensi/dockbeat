package beat

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
	TbConfig             ConfigSettings
	dockerClient         *docker.Client
	events               publisher.Client
	eventGenerator       EventGenerator
	minimalDockerVersion SoftwareVersion
}

func New() *Dockerbeat {
	return &Dockerbeat{}
}

func (d *Dockerbeat) Config(b *beat.Beat) error {

	// Requires Docker 1.9 minimum
	d.minimalDockerVersion = SoftwareVersion{major: 1, minor: 9}

	err := cfgfile.Read(&d.TbConfig, "")
	if err != nil {
		logp.Err("Error reading configuration file: %v", err)
		return err
	}

	//init the period
	if d.TbConfig.Input.Period != nil {
		d.period = time.Duration(*d.TbConfig.Input.Period) * time.Second
	} else {
		d.period = 1 * time.Second
	}
	//init the socketConfig
	d.socketConfig = SocketConfig{
		socket:    "",
		enableTls: false,
		caPath:    "",
		certPath:  "",
		keyPath:   "",
	}

	if d.TbConfig.Input.Socket != nil {
		d.socketConfig.socket = *d.TbConfig.Input.Socket
	} else {
		d.socketConfig.socket = "unix:///var/run/docker.sock" // default docker socket location
	}
	if d.TbConfig.Input.Tls.Enable != nil {
		d.socketConfig.enableTls = *d.TbConfig.Input.Tls.Enable
	} else {
		d.socketConfig.enableTls = false
	}
	if d.socketConfig.enableTls {
		if d.TbConfig.Input.Tls.CaPath != nil {
			d.socketConfig.caPath = *d.TbConfig.Input.Tls.CaPath
		}
		if d.TbConfig.Input.Tls.CertPath != nil {
			d.socketConfig.certPath = *d.TbConfig.Input.Tls.CertPath
		}
		if d.TbConfig.Input.Tls.KeyPath != nil {
			d.socketConfig.keyPath = *d.TbConfig.Input.Tls.KeyPath
		}
	}

	logp.Info("dockerbeat", "Init dockerbeat")
	logp.Info("dockerbeat", "Follow docker socket %q\n", d.socketConfig.socket)
	if d.socketConfig.enableTls {
		logp.Info("dockerbeat", "TLS enabled\n")
	} else {
		logp.Info("dockerbeat", "TLS disabled\n")
	}
	logp.Info("dockerbeat", "Period %v\n", d.period)

	return nil
}

func (d *Dockerbeat) getDockerClient() (*docker.Client, error) {
	var client *docker.Client
	var err error

	if d.socketConfig.enableTls {
		client, err = docker.NewTLSClient(
			d.socketConfig.socket,
			d.socketConfig.certPath,
			d.socketConfig.keyPath,
			d.socketConfig.caPath,
		)
	} else {
		client, err = docker.NewClient(d.socketConfig.socket)
	}
	return client, err
}

func (d *Dockerbeat) Setup(b *beat.Beat) error {
	var clientErr error
	var err error
	//populate Dockerbeat
	d.events = b.Events
	d.done = make(chan struct{})
	d.dockerClient, clientErr = d.getDockerClient()
	d.eventGenerator = EventGenerator{
		socket:            &d.socketConfig.socket,
		networkStats:      EGNetworkStats{m: map[string]map[string]NetworkData{}},
		blkioStats:        EGBlkioStats{m: map[string]BlkioData{}},
		calculatorFactory: CalculatorFactoryImpl{},
		period:            d.period,
	}

	if clientErr != nil {
		err = errors.New(fmt.Sprintf("Unable to create docker client, please check your docker socket/TLS settings: %v", clientErr))
	}
	return err
}

func (d *Dockerbeat) Run(b *beat.Beat) error {
	var err error

	ticker := time.NewTicker(d.period)
	defer ticker.Stop()

	for {
		select {
		case <-d.done:
			return nil
		case <-ticker.C:
		}

		// check prerequisites
		err = d.checkPrerequisites()
		if err != nil {
			logp.Err("Unable to collect metrics: %v", err)
			d.publishLogEvent(ERROR, fmt.Sprintf("Unable to collect metrics: %v", err))
			continue
		}

		timerStart := time.Now()
		d.RunOneTime(b)
		timerEnd := time.Now()

		duration := timerEnd.Sub(timerStart)
		if duration.Nanoseconds() > d.period.Nanoseconds() {
			logp.Warn("Ignoring tick(s) due to processing taking longer than one period")
			d.publishLogEvent(WARN, "Ignoring tick(s) due to processing taking longer than one period")
		}
	}

	return err
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

	d.eventGenerator.cleanOldStats(containers)

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
				d.eventGenerator.getContainerEvent(&container, stats),
				d.eventGenerator.getCpuEvent(&container, stats),
				d.eventGenerator.getMemoryEvent(&container, stats),
				d.eventGenerator.getBlkioEvent(&container, stats),
			}

			events = append(events, d.eventGenerator.getNetworksEvent(&container, stats)...)

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
	event := d.eventGenerator.getLogEvent(level, message)
	d.events.PublishEvent(event)
}
