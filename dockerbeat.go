package main

import (
	"time"

	"errors"
	"github.com/elastic/libbeat/beat"
	"github.com/elastic/libbeat/cfgfile"
	"github.com/elastic/libbeat/common"
	"github.com/elastic/libbeat/logp"
	"github.com/elastic/libbeat/publisher"
	"github.com/fsouza/go-dockerclient"
	"strconv"
	"strings"
)

type Dockerbeat struct {
	isAlive        bool
	period         time.Duration
	socket         string
	TbConfig       ConfigSettings
	dockerClient   *docker.Client
	events         publisher.Client
	eventGenerator EventGenerator
}

func (d *Dockerbeat) Config(b *beat.Beat) error {

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
	d.dockerClient, _ = docker.NewClient(d.socket)
	d.eventGenerator = EventGenerator{map[string]NetworkData{}, map[string]BlkioStats{}}
	return nil
}

func (d *Dockerbeat) Run(b *beat.Beat) error {

	d.isAlive = true

	var err error

	//main loop
	for d.isAlive {
		time.Sleep(d.period)

		// check prerequisites
		env, err := d.dockerClient.Version()

		if err != nil {
			logp.Err("Docker server unreachable: %s", err)
			continue
		}

		valid, _ := d.validVersion(env.Get("version"))

		if !valid {
			logp.Err("Docker server is too old (version 1.5.x and earlier is required)")
		}

		// collect and emit metrics
		d.RunOneTime(b)
	}

	return err
}

func (d *Dockerbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (d *Dockerbeat) Stop() {
	d.isAlive = false
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
			d.eventGenerator.getNetworkEvent(&container, stats),
			d.eventGenerator.getBlkioEvent(&container, stats),
		}

		d.events.PublishEvents(events)
	}()

	return nil
}

func (d *Dockerbeat) validVersion(version string) (bool, error) {
	expectedMinimalMajorVersion := 1
	expectedMinimalMinorVersion := 5

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

	if actualMajorVersion > expectedMinimalMajorVersion ||
	(actualMajorVersion == expectedMinimalMajorVersion && actualMinorVersion >= expectedMinimalMinorVersion) {
		output = true
	} else {
		output = false
	}
	return output, nil
}
