package event

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/fsouza/go-dockerclient"
	"github.com/ingensi/dockbeat/calculator"
	"strings"
	"sync"
	"time"
)

type EGNetworkStats struct {
	sync.RWMutex
	M map[string]map[string]calculator.NetworkData
}

type EGBlkioStats struct {
	sync.RWMutex
	M map[string]calculator.BlkioData
}

type Label struct {
	key   string
	value string
}

type EventGenerator struct {
	Socket            *string
	NetworkStats      EGNetworkStats
	BlkioStats        EGBlkioStats
	CalculatorFactory calculator.CalculatorFactory
	Period            time.Duration
}

func (d *EventGenerator) GetContainerEvent(container *docker.APIContainers, stats *docker.Stats) common.MapStr {
	logp.Debug("generator", "Generate container event %v", container.ID)
	event := common.MapStr{
		"@timestamp":      common.Time(stats.Read),
		"type":            "container",
		"containerID":     container.ID,
		"containerName":   d.extractContainerName(container.Names),
		"containerLabels": d.buildLabelArray(container.Labels),
		"dockerSocket":    d.Socket,
		"container": common.MapStr{
			"id":         container.ID,
			"command":    container.Command,
			"created":    common.Time(time.Unix(container.Created, 0)),
			"image":      container.Image,
			"names":      container.Names,
			"ports":      d.convertContainerPorts(&container.Ports),
			"sizeRootFs": container.SizeRootFs,
			"sizeRw":     container.SizeRw,
			"status":     container.Status,
		},
	}
	return event
}

func (d *EventGenerator) GetCpuEvent(container *docker.APIContainers, stats *docker.Stats) common.MapStr {
	logp.Debug("generator", "Generate cpu event %v", container.ID)
	calculator := d.CalculatorFactory.NewCPUCalculator(
		calculator.CPUData{
			PerCpuUsage:       stats.PreCPUStats.CPUUsage.PercpuUsage,
			TotalUsage:        stats.PreCPUStats.CPUUsage.TotalUsage,
			UsageInKernelmode: stats.PreCPUStats.CPUUsage.UsageInKernelmode,
			UsageInUsermode:   stats.PreCPUStats.CPUUsage.UsageInUsermode,
		},
		calculator.CPUData{
			PerCpuUsage:       stats.CPUStats.CPUUsage.PercpuUsage,
			TotalUsage:        stats.CPUStats.CPUUsage.TotalUsage,
			UsageInKernelmode: stats.CPUStats.CPUUsage.UsageInKernelmode,
			UsageInUsermode:   stats.CPUStats.CPUUsage.UsageInUsermode,
		},
	)

	event := common.MapStr{
		"@timestamp":      common.Time(stats.Read),
		"type":            "cpu",
		"containerID":     container.ID,
		"containerName":   d.extractContainerName(container.Names),
		"containerLabels": d.buildLabelArray(container.Labels),
		"dockerSocket":    d.Socket,
		"cpu": common.MapStr{
			"percpuUsage":       calculator.PerCpuUsage(),
			"totalUsage":        calculator.TotalUsage(),
			"usageInKernelmode": calculator.UsageInKernelmode(),
			"usageInUsermode":   calculator.UsageInUsermode(),
		},
	}

	return event
}

func (d *EventGenerator) GetNetworksEvent(container *docker.APIContainers, stats *docker.Stats) []common.MapStr {
	logp.Debug("generator", "Generate network events %v", container.ID)
	events := []common.MapStr{}

	for netName, netStats := range stats.Networks {
		events = append(events, d.GetNetworkEvent(container, stats.Read, netName, &netStats))
	}

	// purge old saved data
	d.NetworkStats.Lock()
	for container, networkDataMap := range d.NetworkStats.M {
		useless := true
		for networkName, networkData := range networkDataMap {
			// if data older than two ticks, then delete it
			if d.expiredSavedData(networkData.Time) {
				delete(networkDataMap, networkName)
			} else {
				useless = false
			}
		}

		// if all network data are useless, then delete container entry
		if useless {
			delete(d.NetworkStats.M, container)
		}
	}
	d.NetworkStats.Unlock()

	return events
}

func (d *EventGenerator) GetNetworkEvent(container *docker.APIContainers, time time.Time, network string, networkStats *docker.NetworkStats) common.MapStr {
	logp.Debug("generator", "Generate network event %v", container.ID)
	newNetworkData := calculator.NetworkData{
		Time:      time,
		RxBytes:   networkStats.RxBytes,
		RxDropped: networkStats.RxDropped,
		RxErrors:  networkStats.RxErrors,
		RxPackets: networkStats.RxPackets,
		TxBytes:   networkStats.TxBytes,
		TxDropped: networkStats.TxDropped,
		TxErrors:  networkStats.TxErrors,
		TxPackets: networkStats.TxPackets,
	}

	var event common.MapStr

	d.NetworkStats.RLock()
	oldNetworkData, ok := d.NetworkStats.M[container.ID][network]
	d.NetworkStats.RUnlock()

	if ok {
		calculator := d.CalculatorFactory.NewNetworkCalculator(oldNetworkData, newNetworkData)
		event = common.MapStr{
			"@timestamp":      common.Time(time),
			"type":            "net",
			"containerID":     container.ID,
			"containerName":   d.extractContainerName(container.Names),
			"containerLabels": d.buildLabelArray(container.Labels),
			"dockerSocket":    d.Socket,
			"net": common.MapStr{
				"name":         network,
				"rxBytes_ps":   calculator.GetRxBytesPerSecond(),
				"rxDropped_ps": calculator.GetRxDroppedPerSecond(),
				"rxErrors_ps":  calculator.GetRxErrorsPerSecond(),
				"rxPackets_ps": calculator.GetRxPacketsPerSecond(),
				"txBytes_ps":   calculator.GetTxBytesPerSecond(),
				"txDropped_ps": calculator.GetTxDroppedPerSecond(),
				"txErrors_ps":  calculator.GetTxErrorsPerSecond(),
				"txPackets_ps": calculator.GetTxPacketsPerSecond(),
			},
		}
	} else {
		event = common.MapStr{
			"@timestamp":      common.Time(time),
			"type":            "net",
			"containerID":     container.ID,
			"containerName":   d.extractContainerName(container.Names),
			"containerLabels": d.buildLabelArray(container.Labels),
			"dockerSocket":    d.Socket,
			"net": common.MapStr{
				"name":         network,
				"rxBytes_ps":   0,
				"rxDropped_ps": 0,
				"rxErrors_ps":  0,
				"rxPackets_ps": 0,
				"txBytes_ps":   0,
				"txDropped_ps": 0,
				"txErrors_ps":  0,
				"txPackets_ps": 0,
			},
		}
	}

	// save status
	d.NetworkStats.Lock()
	if _, exists := d.NetworkStats.M[container.ID]; !exists {
		d.NetworkStats.M[container.ID] = map[string]calculator.NetworkData{}
	}
	d.NetworkStats.M[container.ID][network] = newNetworkData
	d.NetworkStats.Unlock()
	return event
}

func (d *EventGenerator) GetMemoryEvent(container *docker.APIContainers, stats *docker.Stats) common.MapStr {
	logp.Debug("generator", "Generate memory event %v", container.ID)
	event := common.MapStr{
		"@timestamp":      common.Time(stats.Read),
		"type":            "memory",
		"containerID":     container.ID,
		"containerName":   d.extractContainerName(container.Names),
		"containerLabels": d.buildLabelArray(container.Labels),
		"dockerSocket":    d.Socket,
		"memory": common.MapStr{
			"failcnt":    stats.MemoryStats.Failcnt,
			"limit":      stats.MemoryStats.Limit,
			"maxUsage":   stats.MemoryStats.MaxUsage,
			"totalRss":   stats.MemoryStats.Stats.TotalRss,
			"totalRss_p": float64(stats.MemoryStats.Stats.TotalRss) / float64(stats.MemoryStats.Limit),
			"usage":      stats.MemoryStats.Usage,
			"usage_p":    float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit),
		},
	}

	return event
}

func (d *EventGenerator) GetBlkioEvent(container *docker.APIContainers, stats *docker.Stats) common.MapStr {
	logp.Debug("generator", "Generate blkio event %v", container.ID)
	blkioStats := d.buildStats(stats.Read, stats.BlkioStats.IOServicedRecursive)

	var event common.MapStr

	d.BlkioStats.RLock()
	oldBlkioStats, ok := d.BlkioStats.M[container.ID]
	d.BlkioStats.RUnlock()

	if ok {
		calculator := d.CalculatorFactory.NewBlkioCalculator(oldBlkioStats, blkioStats)
		event = common.MapStr{
			"@timestamp":      common.Time(stats.Read),
			"type":            "blkio",
			"containerID":     container.ID,
			"containerName":   d.extractContainerName(container.Names),
			"containerLabels": d.buildLabelArray(container.Labels),
			"dockerSocket":    d.Socket,
			"blkio": common.MapStr{
				"read_ps":  calculator.GetReadPs(),
				"write_ps": calculator.GetWritePs(),
				"total_ps": calculator.GetTotalPs(),
			},
		}
	} else {
		event = common.MapStr{
			"@timestamp":      common.Time(stats.Read),
			"type":            "blkio",
			"containerID":     container.ID,
			"containerName":   d.extractContainerName(container.Names),
			"containerLabels": d.buildLabelArray(container.Labels),
			"dockerSocket":    d.Socket,
			"blkio": common.MapStr{
				"read_ps":  float64(0),
				"write_ps": float64(0),
				"total_ps": float64(0),
			},
		}
	}

	d.BlkioStats.Lock()
	d.BlkioStats.M[container.ID] = blkioStats

	// purge old saved data
	for containerId, blkioStat := range d.BlkioStats.M {
		// if data older than two ticks, then delete it
		if d.expiredSavedData(blkioStat.Time) {
			delete(d.BlkioStats.M, containerId)
		}
	}
	d.BlkioStats.Unlock()
	return event
}

func (d *EventGenerator) GetLogEvent(level string, message string) common.MapStr {
	logp.Debug("generator", "Generate log event with message: %v", message)
	event := common.MapStr{
		"@timestamp":   common.Time(time.Now()),
		"type":         "log",
		"dockerSocket": d.Socket,
		"log": common.MapStr{
			"level":   level,
			"message": message,
		},
	}
	return event
}

func (d *EventGenerator) convertContainerPorts(ports *[]docker.APIPort) []map[string]interface{} {
	var outputPorts = []map[string]interface{}{}
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

func (d *EventGenerator) CleanOldStats(containers []docker.APIContainers) {
	found := false
	d.NetworkStats.Lock()
	for containerStatKey := range d.NetworkStats.M {
		for _, container := range containers {
			if container.ID == containerStatKey {
				found = true
				continue
			}
		}
		if !found {
			delete(d.NetworkStats.M, containerStatKey)
		}
	}
	d.NetworkStats.Unlock()
}

func (d *EventGenerator) buildStats(time time.Time, entry []docker.BlkioStatsEntry) calculator.BlkioData {
	var stats = calculator.BlkioData{Time: time, Reads: 0, Writes: 0, Totals: 0}
	for _, s := range entry {
		if s.Op == "Read" {
			stats.Reads += s.Value
		} else if s.Op == "Write" {
			stats.Writes += s.Value
		} else if s.Op == "Total" {
			stats.Totals += s.Value
		}
	}
	return stats
}

func (d *EventGenerator) extractContainerName(names []string) string {
	output := names[0]

	if cap(names) > 1 {
		for _, name := range names {
			if strings.Count(output, "/") > strings.Count(name, "/") {
				output = name
			}
		}
	}
	return strings.Trim(output, "/")
}

func (d *EventGenerator) expiredSavedData(date time.Time) bool {
	return !date.Add(2 * d.Period).After(time.Now())
}

func (d *EventGenerator) buildLabelArray(labels map[string]string) []common.MapStr {

	output_labels := make([]common.MapStr, len(labels))

	i := 0
	for k, v := range labels {
		label := strings.Replace(k, ".", "_", -1)
		output_labels[i] = common.MapStr{
			"key":   label,
			"value": v,
		}
		i++
	}

	return output_labels
}
