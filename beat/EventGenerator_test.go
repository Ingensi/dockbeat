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
	// old and current timestamps
	oldTimestamp := time.Now()
	period := time.Second
	newTimestamp := oldTimestamp.Add(period)

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

	// network stats from Docker API
	networkStatsMap := map[string]docker.NetworkStats{}
	networkStatsMap["eth0"] = docker.NetworkStats{
		RxBytes: 10,
		RxDropped: 20,
		RxErrors: 30,
		RxPackets: 40,
		TxBytes: 50,
		TxDropped: 60,
		TxErrors: 70,
		TxPackets: 80,
	}
	networkStatsMap["em1"] = docker.NetworkStats{
		RxBytes: 90,
		RxDropped: 100,
		RxErrors: 110,
		RxPackets: 120,
		TxBytes: 130,
		TxDropped: 140,
		TxErrors: 150,
		TxPackets: 160,
	}

	// main stats object
	var stats = new(docker.Stats)
	stats.Read = newTimestamp
	stats.Networks = networkStatsMap

	// saved network status
	oldNetworkData := map[string]map[string]NetworkData{}
	oldNetworkData[containerId] = map[string]NetworkData{}
	oldNetworkData[containerId]["eth0"] = NetworkData{
		time: oldTimestamp,
		rxBytes   : 1,
		rxDropped : 2,
		rxErrors  : 3,
		rxPackets : 4,
		txBytes   : 5,
		txDropped : 6,
		txErrors  : 7,
		txPackets : 8,
	}
	oldNetworkData[containerId]["em1"] = NetworkData{
		time: oldTimestamp,
		rxBytes   : 9,
		rxDropped : 10,
		rxErrors  : 11,
		rxPackets : 12,
		txBytes   : 13,
		txDropped : 14,
		txErrors  : 15,
		txPackets : 16,
	}

	// mocking calculators
	// first - generate expected calls (NetworkStats to NetworkData conversion)
	newNetworkData := map[string]NetworkData{}
	newNetworkData["eth0"] = NetworkData{
		time: newTimestamp,
		rxBytes   : 10,
		rxDropped : 20,
		rxErrors  : 30,
		rxPackets : 40,
		txBytes   : 50,
		txDropped : 60,
		txErrors  : 70,
		txPackets : 80,
	}
	newNetworkData["em1"] = NetworkData{
		time: newTimestamp,
		rxBytes   : 90,
		rxDropped : 100,
		rxErrors  : 110,
		rxPackets : 120,
		txBytes   : 130,
		txDropped : 140,
		txErrors  : 150,
		txPackets : 160,
	}

	// second - instantiate mock
	mockedCalculatorFactory := new(MockedCalculatorFactory)
	mockedNetworkCalculatorEth0 := getMockedNetworkCalculator(1.0)
	mockedNetworkCalculatorEm1 := getMockedNetworkCalculator(2.0)
	mockedCalculatorFactory.On("newNetworkCalculator", oldNetworkData[containerId]["eth0"], newNetworkData["eth0"]).Return(mockedNetworkCalculatorEth0)
	mockedCalculatorFactory.On("newNetworkCalculator", oldNetworkData[containerId]["em1"], newNetworkData["em1"]).Return(mockedNetworkCalculatorEm1)

	// expected events
	expectedEvents := []common.MapStr{}
	expectedEvents = append(expectedEvents,
		common.MapStr{
			"@timestamp":    common.Time(newTimestamp),
			"type":          "net",
			"containerID":   container.ID,
			"containerName": "name1",
			"net": common.MapStr{
				"name":         "eth0",
				"rxBytes_ps":   mockedNetworkCalculatorEth0.getRxBytesPerSecond(),
				"rxDropped_ps": mockedNetworkCalculatorEth0.getRxDroppedPerSecond(),
				"rxErrors_ps":  mockedNetworkCalculatorEth0.getRxErrorsPerSecond(),
				"rxPackets_ps": mockedNetworkCalculatorEth0.getRxPacketsPerSecond(),
				"txBytes_ps":   mockedNetworkCalculatorEth0.getTxBytesPerSecond(),
				"txDropped_ps": mockedNetworkCalculatorEth0.getTxDroppedPerSecond(),
				"txErrors_ps":  mockedNetworkCalculatorEth0.getTxErrorsPerSecond(),
				"txPackets_ps": mockedNetworkCalculatorEth0.getTxPacketsPerSecond(),
			}},
		common.MapStr{
			"@timestamp":    common.Time(newTimestamp),
			"type":          "net",
			"containerID":   container.ID,
			"containerName": "name1",
			"net": common.MapStr{
				"name":         "em1",
				"rxBytes_ps":   mockedNetworkCalculatorEm1.getRxBytesPerSecond(),
				"rxDropped_ps": mockedNetworkCalculatorEm1.getRxDroppedPerSecond(),
				"rxErrors_ps":  mockedNetworkCalculatorEm1.getRxErrorsPerSecond(),
				"rxPackets_ps": mockedNetworkCalculatorEm1.getRxPacketsPerSecond(),
				"txBytes_ps":   mockedNetworkCalculatorEm1.getTxBytesPerSecond(),
				"txDropped_ps": mockedNetworkCalculatorEm1.getTxDroppedPerSecond(),
				"txErrors_ps":  mockedNetworkCalculatorEm1.getTxErrorsPerSecond(),
				"txPackets_ps": mockedNetworkCalculatorEm1.getTxPacketsPerSecond(),
			}})

	// the eventGenerator to test
	var eventGenerator = EventGenerator{oldNetworkData, nil, mockedCalculatorFactory}

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
	assert.Equal(t, eventGenerator.networkStats[container.ID]["eth0"], newNetworkData["eth0"])
	assert.Equal(t, eventGenerator.networkStats[container.ID]["em1"], newNetworkData["em1"])
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

// METHODS

func getMockedNetworkCalculator(number float64) *MockedNetworkCalculator {
	mock := new(MockedNetworkCalculator)
	mock.On("getRxBytesPerSecond").Return(number)
	mock.On("getRxDroppedPerSecond").Return(number * 2)
	mock.On("getRxErrorsPerSecond").Return(number * 3)
	mock.On("getRxPacketsPerSecond").Return(number * 4)
	mock.On("getTxBytesPerSecond").Return(number * 5)
	mock.On("getTxDroppedPerSecond").Return(number * 6)
	mock.On("getTxErrorsPerSecond").Return(number * 7)
	mock.On("getTxPacketsPerSecond").Return(number * 8)
	return mock
}