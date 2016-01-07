package beat

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEventGeneratorGetNetworksEvent(t *testing.T) {
	// GIVEN
	// a container
	labels := make(map[string]string)
	labels["label1"] = "value1"
	labels["label2"] = "value2"
	containerId := "container_id"
	var container = docker.APIContainers{
		containerId,
		"container_image",
		"container command",
		9876543210,
		"Up",
		[]docker.APIPort{docker.APIPort{1234, 4567, "portType", "123.456.879.1"}},
		123,
		456,
		[]string{"/name1", "name1/fake"},
		labels,
	}

	oldTimestamp := time.Now()
	period := time.Second
	newTimestamp := oldTimestamp.Add(period)

	// network stats
	networkStatsMap := map[string]docker.NetworkStats{}
	// /!\ values order: RxDropped, RxBytes, RxErrors, TxPackets, TxDropped, RxPackets, TxErrors, TxBytes
	networkStatsMap["eth0"] = docker.NetworkStats{20, 10, 30, 80, 60, 40, 70, 50}
	networkStatsMap["em1"] = docker.NetworkStats{200, 100, 300, 800, 600, 400, 700, 500}

	// saved network status
	savedNetworkData := map[string]map[string]NetworkData{}
	savedNetworkData[containerId] = map[string]NetworkData{}
	savedNetworkData[containerId]["eth0"] = NetworkData{oldTimestamp, 5, 10, 15, 20, 25, 30, 35, 40}
	savedNetworkData[containerId]["em1"] = NetworkData{oldTimestamp, 10, 20, 30, 40, 50, 60, 70, 80}

	// main stats object
	var stats = new(docker.Stats)
	stats.Read = newTimestamp
	stats.Networks = networkStatsMap
	var eventGenerator = EventGenerator{savedNetworkData, nil, CalculatorFactoryImpl{}}

	// expected events
	expectedEvents := []common.MapStr{}
	expectedEvents = append(expectedEvents,
		common.MapStr{
			"@timestamp":    common.Time(newTimestamp),
			"type":          "net",
			"containerID":   container.ID,
			"containerName": "name1",
			"net": common.MapStr{
				"name":         "em1",
				"rxBytes_ps":   90,
				"rxDropped_ps": 180,
				"rxErrors_ps":  270,
				"rxPackets_ps": 360,
				"txBytes_ps":   450,
				"txDropped_ps": 540,
				"txErrors_ps":  630,
				"txPackets_ps": 720,
			}},
		common.MapStr{
			"@timestamp":    common.Time(newTimestamp),
			"type":          "net",
			"containerID":   container.ID,
			"containerName": "name1",
			"net": common.MapStr{
				"name":         "eth0",
				"rxBytes_ps":   5,
				"rxDropped_ps": 10,
				"rxErrors_ps":  15,
				"rxPackets_ps": 20,
				"txBytes_ps":   25,
				"txDropped_ps": 30,
				"txErrors_ps":  35,
				"txPackets_ps": 40,
			}})

	// WHEN
	events := eventGenerator.getNetworksEvent(&container, stats, period)

	// THEN
	// check returned events
	assert.Equal(t, len(expectedEvents), 2)
	for i, _ := range expectedEvents {
		checked := false
		for j, _ := range events {
			if expectedEvents[i].String() == events[j].String() {
				checked = true
				break
			}
		}
		if !checked {
			assert.Fail(t, "unable to find network in events: %s", expectedEvents[i].String())
		}
	}

	// check that new stats saved
	assert.Equal(t, eventGenerator.networkStats[container.ID]["eth0"], NetworkData{newTimestamp, 10, 20, 30, 40, 50, 60, 70, 80})
	assert.Equal(t, eventGenerator.networkStats[container.ID]["em1"], NetworkData{newTimestamp, 100, 200, 300, 400, 500, 600, 700, 800})
}

func TestEventGeneratorGetContainerEvent(t *testing.T) {
	// GIVEN
	labels := make(map[string]string)
	labels["label1"] = "value1"
	labels["label2"] = "value2"
	var container = docker.APIContainers{
		"container_id",
		"container_image",
		"container command",
		9876543210,
		"Up",
		[]docker.APIPort{docker.APIPort{1234, 4567, "portType", "123.456.879.1"}},
		123,
		456,
		[]string{"/name1", "name1/fake"},
		labels,
	}

	timestamp := time.Now()
	var stats = new(docker.Stats)
	stats.Read = timestamp
	var eventGenerator = EventGenerator{nil, nil, CalculatorFactoryImpl{}}

	// expected output
	expectedEvent := common.MapStr{
		"@timestamp":    common.Time(timestamp),
		"type":          "container",
		"containerID":   container.ID,
		"containerName": "name1",
		"container": common.MapStr{
			"id":      container.ID,
			"command": container.Command,
			"created": time.Unix(container.Created, 0),
			"image":   container.Image,
			"labels":  container.Labels,
			"names":   container.Names,
			"ports": []map[string]interface{}{common.MapStr{
				"ip":          container.Ports[0].IP,
				"privatePort": container.Ports[0].PrivatePort,
				"publicPort":  container.Ports[0].PublicPort,
				"type":        container.Ports[0].Type,
			}},
			"sizeRootFs": container.SizeRootFs,
			"sizeRw":     container.SizeRw,
			"status":     container.Status,
		},
	}

	// WHEN
	event := eventGenerator.getContainerEvent(&container, stats)

	// THEN
	assert.Equal(t, expectedEvent, event)
}

func TestBuildStats(t *testing.T) {
	//GIVEN
	var eventGenerator = EventGenerator{nil, nil, CalculatorFactoryImpl{}}
	var data []docker.BlkioStatsEntry
	data = make([]docker.BlkioStatsEntry, 20, 20)
	data[0] = docker.BlkioStatsEntry{0, 0, "Read", 1000}
	data[1] = docker.BlkioStatsEntry{0, 0, "Write", 2000}
	data[2] = docker.BlkioStatsEntry{0, 0, "Total", 3000}
	data[3] = docker.BlkioStatsEntry{0, 1, "Read", 10}
	data[4] = docker.BlkioStatsEntry{0, 1, "Write", 20}
	data[5] = docker.BlkioStatsEntry{0, 1, "Total", 30}

	//WHEN
	value := eventGenerator.buildStats(data)

	//THEN
	assert.Equal(t, uint64(1010), value.reads)
	assert.Equal(t, uint64(2020), value.writes)
	assert.Equal(t, uint64(3030), value.totals)
}

func TestExtractContainerNameAlone(t *testing.T) {
	// GIVEN
	var eventGenerator = EventGenerator{nil, nil, CalculatorFactoryImpl{}}
	expectedName := "containerName"

	// WHEN
	name := eventGenerator.extractContainerName([]string{"/" + expectedName})

	// THEN
	assert.Equal(t, expectedName, name)
}

func TestExtractContainerNameMultiple(t *testing.T) {
	// GIVEN
	var eventGenerator = EventGenerator{nil, nil, CalculatorFactoryImpl{}}
	expectedName := "containerName"

	// WHEN
	name := eventGenerator.extractContainerName([]string{"/name1/fake", "/" + expectedName, "/name3/fake"})

	// THEN
	assert.Equal(t, expectedName, name)
}
