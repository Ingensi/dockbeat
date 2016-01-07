package beat

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"
	"github.com/fsouza/go-dockerclient"
)

type SoftwareVersion struct {
	major int
	minor int
}

type Dockerbeat struct {
	done                 chan struct{}
	period               time.Duration
	socket               string
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

	// Requires Docker 1.5 minimum
	d.minimalDockerVersion = SoftwareVersion{1, 5}

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
	//init the socket
	if d.TbConfig.Input.Socket != nil {
		d.socket = *d.TbConfig.Input.Socket
	} else {
		d.socket = "unix:///var/run/docker.sock" // default docker socket location
	}

	logp.Debug("dockerbeat", "Init dockerbeat")
	logp.Debug("dockerbeat", "Follow docker socket %q\n", d.socket)
	logp.Debug("dockerbeat", "Period %v\n", d.period)

	return nil
}

func (d *Dockerbeat) Setup(b *beat.Beat) error {
	//populate Dockerbeat
	d.events = b.Events
	d.done = make(chan struct{})
	d.dockerClient, _ = docker.NewClient(d.socket)
	d.eventGenerator = EventGenerator{map[string]map[string]NetworkData{}, map[string]BlkioStats{}}

	return d.checkPrerequisites()
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
		if d.checkPrerequisites() != nil {
			logp.Err("Unable to collect metrics: %s", err)
			continue
		}

		timerStart := time.Now()
		d.RunOneTime(b)
		timerEnd := time.Now()

		duration := timerEnd.Sub(timerStart)
		if duration.Nanoseconds() > d.period.Nanoseconds() {
			logp.Warn("Ignoring tick(s) due to processing taking longer than one period")
		}
	}

	return err
}

func (d *Dockerbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (d *Dockerbeat) Stop() {
	close(d.done)
}

func (d *Dockerbeat) RunOneTime(b *beat.Beat) error {
	containers, err := d.dockerClient.ListContainers(docker.ListContainersOptions{})

	if err == nil {
		//export stats for each container
		for _, container := range containers {
			d.exportContainerStats(container)
		}
	} else {
		logp.Err("Cannot get container list: %d", err)
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
	statsOptions := docker.StatsOptions{container.ID, statsC, false, done, -1}
	// goroutine to listen to the stats
	go func() {
		errC <- d.dockerClient.Stats(statsOptions)
		close(errC)
	}()
	// goroutine to get the stats & publish it
	go func() {
		stats := <-statsC

		events := []common.MapStr{
			d.eventGenerator.getContainerEvent(&container, stats),
			d.eventGenerator.getCpuEvent(&container, stats),
			d.eventGenerator.getMemoryEvent(&container, stats),
			d.eventGenerator.getBlkioEvent(&container, stats),
		}

		events = append(events, d.eventGenerator.getNetworksEvent(&container, stats, d.period)...)

		d.events.PublishEvents(events)
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
