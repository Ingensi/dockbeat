package main

import (
	"time"

	"github.com/elastic/libbeat/beat"
	"github.com/elastic/libbeat/cfgfile"
	"github.com/elastic/libbeat/logp"
	"github.com/elastic/libbeat/publisher"
	"github.com/fsouza/go-dockerclient"
	"github.com/elastic/libbeat/common"
)

type Dockerbeat struct {
	isAlive      bool
	period       time.Duration
	socket       string
	TbConfig     ConfigSettings
	dockerClient *docker.Client
	networkStats map[string]NetworkData
	events       publisher.Client
}

func (d *Dockerbeat) Config(b *beat.Beat) error {

	err := cfgfile.Read(&d.TbConfig, "")
	if err != nil {
		logp.Err("Error reading configuration file: %v", err)
		return err
	}

	if d.TbConfig.Input.Period != nil {
		d.period = time.Duration(*d.TbConfig.Input.Period) * time.Second
	} else {
		d.period = 1 * time.Second
	}
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
	d.events = b.Events
	d.dockerClient, _ = docker.NewClient(d.socket)
	d.networkStats = make(map[string]NetworkData)
	return nil
}

func (d *Dockerbeat) Run(b *beat.Beat) error {

	d.isAlive = true

	var err error

	for d.isAlive {
		time.Sleep(d.period)
		containers, err := d.dockerClient.ListContainers(docker.ListContainersOptions{})

		if err == nil {
			for _, container := range containers {
				d.exportContainerStats(container)
			}
		} else {
			logp.Err("Cannot get container list: %d", err)
		}

		d.cleanOldStats(containers)
	}

	return err
}

func (d *Dockerbeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (d *Dockerbeat) Stop() {
	d.isAlive = false
}

func (d *Dockerbeat) exportContainerStats(container docker.APIContainers) error {
	statsC := make(chan *docker.Stats)
	done := make(chan bool)
	errC := make(chan error, 1)
	statsOptions := docker.StatsOptions{container.ID, statsC, false, done, -1}
	go func() {
		errC <- d.dockerClient.Stats(statsOptions)
		close(errC)
	}()

	go func() {
		stats := <-statsC

		events := []common.MapStr{
			d.getContainerEvent(&container, stats),
			d.getCpuEvent(&container, stats),
			d.getMemoryEvent(&container, stats),
			d.getNetworkEvent(&container, stats),
		}

		d.events.PublishEvents(events)
	}()

	return nil
}

func (d *Dockerbeat) getContainerEvent(container *docker.APIContainers, stats *docker.Stats) common.MapStr {
	event := common.MapStr{
		"timestamp":      common.Time(stats.Read),
		"type":           "container",
		"containerID":    container.ID,
		"containerNames": container.Names,
		"container":      common.MapStr{
			"id":         container.ID,
			"command":    container.Command,
			"created":    time.Unix(container.Created, 0),
			"image":      container.Image,
			"labels":     container.Labels,
			"names":      container.Names,
			"ports":      d.convertContainerPorts(&container.Ports),
			"sizeRootFs": container.SizeRootFs,
			"sizeRw":     container.SizeRw,
			"status":     container.Status,
		},
	}
	return event
}

func (d *Dockerbeat) getCpuEvent(container *docker.APIContainers, stats *docker.Stats) common.MapStr {

	calculator := CPUCalculator{
		CPUData{stats.PreCPUStats.CPUUsage.PercpuUsage, stats.PreCPUStats.CPUUsage.TotalUsage, stats.PreCPUStats.CPUUsage.UsageInKernelmode, stats.PreCPUStats.CPUUsage.UsageInUsermode},
		CPUData{stats.CPUStats.CPUUsage.PercpuUsage, stats.CPUStats.CPUUsage.TotalUsage, stats.CPUStats.CPUUsage.UsageInKernelmode, stats.CPUStats.CPUUsage.UsageInUsermode},
	}

	event := common.MapStr{
		"timestamp":     common.Time(stats.Read),
		"type":          "cpu",
		"containerID":    container.ID,
		"containerNames": container.Names,
		"cpu": common.MapStr{
			"percpuUsage":       calculator.perCpuUsage(),
			"totalUsage":        calculator.totalUsage(),
			"usageInKernelmode": calculator.usageInKernelmode(),
			"usageInUsermode":   calculator.usageInUsermode(),
		},
	}

	return event
}

func (d *Dockerbeat) getNetworkEvent(container *docker.APIContainers, stats *docker.Stats) common.MapStr {
	newNetworkData := NetworkData{
		stats.Read,
		stats.Network.RxBytes,
		stats.Network.RxDropped,
		stats.Network.RxErrors,
		stats.Network.RxPackets,
		stats.Network.TxBytes,
		stats.Network.TxDropped,
		stats.Network.TxErrors,
		stats.Network.TxPackets,
	}

	var event common.MapStr

	oldNetworkData, ok := d.networkStats[container.ID]

	if ok {
		calculator := NetworkCalculator{oldNetworkData, newNetworkData}
		event = common.MapStr{
			"timestamp":      common.Time(stats.Read),
			"type":           "net",
			"containerID":    container.ID,
			"containerNames": container.Names,
			"net":            common.MapStr{
				"rxBytes_ps":   calculator.getRxBytesPerSecond(),
				"rxDropped_ps": calculator.getRxDroppedPerSecond(),
				"rxErrors_ps":  calculator.getRxErrorsPerSecond(),
				"rxPackets_ps": calculator.getRxPacketsPerSecond(),
				"txBytes_ps":   calculator.getTxBytesPerSecond(),
				"txDropped_ps": calculator.getTxDroppedPerSecond(),
				"txErrors_ps":  calculator.getTxErrorsPerSecond(),
				"txPackets_ps": calculator.getTxPacketsPerSecond(),
			},
		}
	} else {
		event = common.MapStr{
			"timestamp":      common.Time(stats.Read),
			"type":           "net",
			"containerID":    container.ID,
			"containerNames": container.Names,
			"net":            common.MapStr{
				"rxBytes":   0,
				"rxDropped": 0,
				"rxErrors":  0,
				"rxPackets": 0,
				"txBytes":   0,
				"txDropped": 0,
				"txErrors":  0,
				"txPackets": 0,
			},
		}
	}

	d.networkStats[container.ID] = newNetworkData;
	return event

}

func (d *Dockerbeat) getMemoryEvent(container *docker.APIContainers, stats *docker.Stats) common.MapStr {
	event := common.MapStr{
		"timestamp":      common.Time(stats.Read),
		"type":           "memory",
		"containerID":    container.ID,
		"containerNames": container.Names,
		"memory":         common.MapStr{
			"failcnt":  stats.MemoryStats.Failcnt,
			"limit":    stats.MemoryStats.Limit,
			"maxUsage": stats.MemoryStats.MaxUsage,
			"usage":    stats.MemoryStats.Usage,
			"usage_p": (float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit)) * 100,
		},
	}

	return event
}

func (d *Dockerbeat) convertContainerPorts(ports *[]docker.APIPort) []map[string]interface{} {
	var outputPorts []map[string]interface{}
	for _, port := range *ports {
		outputPort := common.MapStr{
			"ip":          port.IP,
			"privatePort": port.PrivatePort,
			"publicPort":  port.PublicPort,
			"type":        port.Type,
		}
		outputPorts = append(outputPorts, outputPort)
	}

	return outputPorts
}

func (d *Dockerbeat) cleanOldStats(containers []docker.APIContainers) {
	found := false
	for containerStatKey, _ := range d.networkStats {
		for _, container := range containers {
			if container.ID == containerStatKey {
				found = true
				continue
			}
		}
		if !found {
			delete(d.networkStats, containerStatKey)
		}
	}
}