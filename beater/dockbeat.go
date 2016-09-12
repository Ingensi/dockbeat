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

	"github.com/ingensi/dockbeat/calculator"
	"github.com/ingensi/dockbeat/config"
	"github.com/ingensi/dockbeat/event"
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

type StatsConfig struct {
	Container bool
	Net       bool
	Memory    bool
	Blkio     bool
	Cpu       bool
}

type Dockbeat struct {
	done                 chan struct{}
	period               time.Duration
	socketConfig         SocketConfig
	statsConfig          StatsConfig
	beatConfig           *config.Config
	dockerClient         *docker.Client
	events               publisher.Client
	eventGenerator       *event.EventGenerator
	minimalDockerVersion SoftwareVersion
}

// Creates beater
func New() *Dockbeat {
	return &Dockbeat{}
}

/// *** Beater interface methods ***///

func (bt *Dockbeat) Config(b *beat.Beat) error {

	// Requires Docker 1.9 minimum
	bt.minimalDockerVersion = SoftwareVersion{major: 1, minor: 9}

	err := cfgfile.Read(&bt.beatConfig, "")
	if err != nil {
		logp.Err("dockbeat", "Error reading configuration file: %v", err)
		return err
	}

	//init the period
	if bt.beatConfig.Dockbeat.Period != nil {
		bt.period = time.Duration(*bt.beatConfig.Dockbeat.Period) * time.Second
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

	if bt.beatConfig.Dockbeat.Socket != nil {
		bt.socketConfig.socket = *bt.beatConfig.Dockbeat.Socket
	} else {
		bt.socketConfig.socket = "unix:///var/run/docker.sock" // default docker socket location
	}
	if bt.beatConfig.Dockbeat.Tls.Enable != nil {
		bt.socketConfig.enableTls = *bt.beatConfig.Dockbeat.Tls.Enable
	} else {
		bt.socketConfig.enableTls = false
	}
	if bt.socketConfig.enableTls {
		if bt.beatConfig.Dockbeat.Tls.CaPath != nil {
			bt.socketConfig.caPath = *bt.beatConfig.Dockbeat.Tls.CaPath
		}
		if bt.beatConfig.Dockbeat.Tls.CertPath != nil {
			bt.socketConfig.certPath = *bt.beatConfig.Dockbeat.Tls.CertPath
		}
		if bt.beatConfig.Dockbeat.Tls.KeyPath != nil {
			bt.socketConfig.keyPath = *bt.beatConfig.Dockbeat.Tls.KeyPath
		}
	}

	// init the stats statsConfig
	bt.statsConfig = StatsConfig{
		Container: true,
		Net:       true,
		Memory:    true,
		Blkio:     true,
		Cpu:       true,
	}

	if bt.beatConfig.Dockbeat.Stats.Container != nil && !*bt.beatConfig.Dockbeat.Stats.Container {
		bt.statsConfig.Container = false
	}
	if bt.beatConfig.Dockbeat.Stats.Net != nil && !*bt.beatConfig.Dockbeat.Stats.Net {
		bt.statsConfig.Net = false
	}
	if bt.beatConfig.Dockbeat.Stats.Memory != nil && !*bt.beatConfig.Dockbeat.Stats.Memory {
		bt.statsConfig.Memory = false
	}
	if bt.beatConfig.Dockbeat.Stats.Blkio != nil && !*bt.beatConfig.Dockbeat.Stats.Blkio {
		bt.statsConfig.Blkio = false
	}
	if bt.beatConfig.Dockbeat.Stats.Cpu != nil && !*bt.beatConfig.Dockbeat.Stats.Cpu {
		bt.statsConfig.Cpu = false
	}

	logp.Info("dockbeat", "Init dockbeat")
	logp.Info("dockbeat", "Follow docker socket %v\n", bt.socketConfig.socket)
	if bt.socketConfig.enableTls {
		logp.Info("dockbeat", "TLS enabled\n")
	} else {
		logp.Info("dockbeat", "TLS disabled\n")
	}
	logp.Info("dockbeat", "Period %v\n", bt.period)

	return nil
}

func (bt *Dockbeat) getDockerClient() (*docker.Client, error) {
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

func (bt *Dockbeat) Setup(b *beat.Beat) error {
	var clientErr error
	var err error
	//populate Dockbeat
	bt.events = b.Events
	bt.done = make(chan struct{})
	bt.dockerClient, clientErr = bt.getDockerClient()
	bt.eventGenerator = &event.EventGenerator{
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

func (bt *Dockbeat) Run(b *beat.Beat) error {
	logp.Info("dockbeat", "dockbeat is running! Hit CTRL-C to stop it.")
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
			logp.Err("dockbeat", "Unable to collect metrics: %v", err)
			bt.publishLogEvent(ERROR, fmt.Sprintf("Unable to collect metrics: %v", err))
			continue
		}

		timerStart := time.Now()
		bt.RunOneTime(b)
		timerEnd := time.Now()

		duration := timerEnd.Sub(timerStart)
		if duration.Nanoseconds() > bt.period.Nanoseconds() {
			logp.Warn("dockbeat", "Ignoring tick(s) due to processing taking longer than one period")
			bt.publishLogEvent(WARN, "Ignoring tick(s) due to processing taking longer than one period")
		}
	}
}

func (d *Dockbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (d *Dockbeat) Stop() {
	close(d.done)
	logp.Info("dockbeat", "Stopping dockbeat")
}

func (d *Dockbeat) RunOneTime(b *beat.Beat) error {
	logp.Debug("dockbeat", "Tick!, getting list of containers")
	containers, err := d.dockerClient.ListContainers(docker.ListContainersOptions{})

	if err == nil {
		logp.Debug("dockbeat", "got %v containers", len(containers))
		//export stats for each container
		for _, container := range containers {
			d.exportContainerStats(container)
		}
	} else {
		logp.Err("dockbeat", "Cannot get container list: %v", err)
		d.publishLogEvent(ERROR, fmt.Sprintf("Cannot get container list: %v", err))
	}

	d.eventGenerator.CleanOldStats(containers)

	return nil
}

func (d *Dockbeat) exportContainerStats(container docker.APIContainers) error {
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
			events := []common.MapStr{}

			// export events if it is enabled in the configuration

			if d.statsConfig.Container {
				logp.Debug("dockbeat", "generating container event for %v", container.ID)
				events = append(events, d.eventGenerator.GetContainerEvent(&container, stats))
				logp.Debug("dockbeat", "container event append to event list (container %v)", container.ID)
			}

			if d.statsConfig.Cpu {
				logp.Debug("dockbeat", "generating cpu event for %v", container.ID)
				events = append(events, d.eventGenerator.GetCpuEvent(&container, stats))
				logp.Debug("dockbeat", "container cpu append to event list (container %v)", container.ID)

			}

			if d.statsConfig.Memory {
				logp.Debug("dockbeat", "generating memory event for %v", container.ID)
				events = append(events, d.eventGenerator.GetMemoryEvent(&container, stats))
				logp.Debug("dockbeat", "container memory append to event list (container %v)", container.ID)

			}

			if d.statsConfig.Blkio {
				logp.Debug("dockbeat", "generating blkio event for %v", container.ID)
				events = append(events, d.eventGenerator.GetBlkioEvent(&container, stats))
				logp.Debug("dockbeat", "container blkio append to event list (container %v)", container.ID)

			}

			if d.statsConfig.Net {
				logp.Debug("dockbeat", "generating net event for %v", container.ID)
				events = append(events, d.eventGenerator.GetNetworksEvent(&container, stats)...)
				logp.Debug("dockbeat", "container net append to event list (container %v)", container.ID)

			}

			logp.Info("dockbeat", "Publishing %v events", len(events))
			d.events.PublishEvents(events)
		} else if err == nil && stats == nil {
			logp.Warn("dockbeat", "Container was existing at listing but not when getting statistics: %v", container.ID)
			d.publishLogEvent(WARN, fmt.Sprintf("Container was existing at listing but not when getting statistics: %v", container.ID))
		} else {
			logp.Err("dockbeat", "An error occurred while getting docker stats: %v", err)
			d.publishLogEvent(ERROR, fmt.Sprintf("An error occurred while getting docker stats: %v", err))
		}
	}()

	return nil
}

func (d *Dockbeat) checkPrerequisites() error {
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

func (d *Dockbeat) validVersion(version string) (bool, error) {

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

func (d *Dockbeat) publishLogEvent(level string, message string) {
	event := d.eventGenerator.GetLogEvent(level, message)
	d.events.PublishEvent(event)
}
