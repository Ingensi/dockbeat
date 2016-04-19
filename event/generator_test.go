package event

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"github.com/ingensi/dockerbeat/calculator"
)

// NETWORK EVENT GENERATION

/*
TestEventGeneratorGetNetworksEventFirstPass simulates the case when a new network event should be generated

It simulates following status:
  - a common container
  - network stats with two networks "eth0" and "em1"

The network "eth0" already have an saved status from previous tick.

This test checks that it generate two network events:
  - an event for "eth0" with calculated data from saved stats (+ new stats saved)
  - an event for "em1" with zeros values (+ new stats saved)
*/
func TestEventGeneratorGetNetworksEventFirstPass(t *testing.T) {
	// GIVEN
	// docker socket
	socket := "unix:///some/docker/socket"
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
		RxBytes:   10,
		RxDropped: 20,
		RxErrors:  30,
		RxPackets: 40,
		TxBytes:   50,
		TxDropped: 60,
		TxErrors:  70,
		TxPackets: 80,
	}
	networkStatsMap["em1"] = docker.NetworkStats{
		RxBytes:   90,
		RxDropped: 100,
		RxErrors:  110,
		RxPackets: 120,
		TxBytes:   130,
		TxDropped: 140,
		TxErrors:  150,
		TxPackets: 160,
	}

	// main stats object
	var stats = new(docker.Stats)
	stats.Read = newTimestamp
	stats.Networks = networkStatsMap

	// saved network status (em1 does not already exists)
	oldNetworkData := map[string]map[string]calculator.NetworkData{}
	oldNetworkData[containerId] = map[string]calculator.NetworkData{}
	oldNetworkData[containerId]["eth0"] = calculator.NetworkData{
		Time:      oldTimestamp,
		RxBytes:   1,
		RxDropped: 2,
		RxErrors:  3,
		RxPackets: 4,
		TxBytes:   5,
		TxDropped: 6,
		TxErrors:  7,
		TxPackets: 8,
	}

	// mocking calculators
	// first - generate expected calls (NetworkStats to NetworkData conversion)
	newNetworkData := map[string]calculator.NetworkData{}
	newNetworkData["eth0"] = calculator.NetworkData{
		Time:      newTimestamp,
		RxBytes:   10,
		RxDropped: 20,
		RxErrors:  30,
		RxPackets: 40,
		TxBytes:   50,
		TxDropped: 60,
		TxErrors:  70,
		TxPackets: 80,
	}
	newNetworkData["em1"] = calculator.NetworkData{
		Time:      newTimestamp,
		TxBytes:   90,
		TxDropped: 100,
		TxErrors:  110,
		TxPackets: 120,
		TxBytes:   130,
		TxDropped: 140,
		TxErrors:  150,
		TxPackets: 160,
	}

	// second - instantiate mock
	// calculator will no be called for em1 network, it will generate zero-values event for em1
	mockedCalculatorFactory := new(calculator.MockedCalculatorFactory)
	mockedNetworkCalculatorEth0 := getMockedNetworkCalculator(1.0)
	mockedCalculatorFactory.On("newNetworkCalculator", oldNetworkData[containerId]["eth0"], newNetworkData["eth0"]).Return(mockedNetworkCalculatorEth0)

	// expected events
	expectedEvents := []common.MapStr{}
	expectedEvents = append(expectedEvents,
		common.MapStr{
			"@timestamp":    common.Time(newTimestamp),
			"type":          "net",
			"containerID":   container.ID,
			"containerName": "name1",
			"dockerSocket":  &socket,
			"net": common.MapStr{
				"name":         "eth0",
				"rxBytes_ps":   mockedNetworkCalculatorEth0.GetRxBytesPerSecond(),
				"rxDropped_ps": mockedNetworkCalculatorEth0.GetRxDroppedPerSecond(),
				"rxErrors_ps":  mockedNetworkCalculatorEth0.GetRxErrorsPerSecond(),
				"rxPackets_ps": mockedNetworkCalculatorEth0.GetRxPacketsPerSecond(),
				"txBytes_ps":   mockedNetworkCalculatorEth0.GetTxBytesPerSecond(),
				"txDropped_ps": mockedNetworkCalculatorEth0.GetTxDroppedPerSecond(),
				"txErrors_ps":  mockedNetworkCalculatorEth0.GetTxErrorsPerSecond(),
				"txPackets_ps": mockedNetworkCalculatorEth0.GetTxPacketsPerSecond(),
			}},
		common.MapStr{
			"@timestamp":    common.Time(newTimestamp),
			"type":          "net",
			"containerID":   container.ID,
			"containerName": "name1",
			"dockerSocket":  &socket,
			"net": common.MapStr{
				"name":         "em1",
				"rxBytes_ps":   0,
				"rxDropped_ps": 0,
				"rxErrors_ps":  0,
				"rxPackets_ps": 0,
				"txBytes_ps":   0,
				"txDropped_ps": 0,
				"txErrors_ps":  0,
				"txPackets_ps": 0,
			}})

	// the eventGenerator to test
	var eventGenerator = EventGenerator{&socket, EGNetworkStats{M: oldNetworkData}, EGBlkioStats{}, mockedCalculatorFactory, period}

	// WHEN
	events := eventGenerator.GetNetworksEvent(&container, stats)

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
	assert.Equal(t, eventGenerator.NetworkStats.M[container.ID]["eth0"], newNetworkData["eth0"])
	assert.Equal(t, eventGenerator.NetworkStats.M[container.ID]["em1"], newNetworkData["em1"])
}

/*
TestEventGeneratorGetNetworksEvent simulates the case when all networks are already in saved status

It simulates following status:
  - a common container
  - network stats with two networks "eth0" and "em1"

Networks "eth0" and "em1" already have an saved status from previous tick.

This test checks that it generate two network events:
  - an event for "eth0" with calculated data from saved stats (+ new stats saved)
  - an event for "em1" with calculated data from saved stats (+ new stats saved)
*/
func TestEventGeneratorGetNetworksEvent(t *testing.T) {
	// GIVEN
	// docker socket
	socket := "unix:///some/docker/socket"

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
		RxBytes:   10,
		RxDropped: 20,
		RxErrors:  30,
		RxPackets: 40,
		TxBytes:   50,
		TxDropped: 60,
		TxErrors:  70,
		TxPackets: 80,
	}
	networkStatsMap["em1"] = docker.NetworkStats{
		RxBytes:   90,
		RxDropped: 100,
		RxErrors:  110,
		RxPackets: 120,
		TxBytes:   130,
		TxDropped: 140,
		TxErrors:  150,
		TxPackets: 160,
	}

	// main stats object
	var stats = new(docker.Stats)
	stats.Read = newTimestamp
	stats.Networks = networkStatsMap

	// saved network status
	oldNetworkData := map[string]map[string]calculator.NetworkData{}
	oldNetworkData[containerId] = map[string]calculator.NetworkData{}
	oldNetworkData[containerId]["eth0"] = calculator.NetworkData{
		Time:      oldTimestamp,
		RxBytes:   1,
		RxDropped: 2,
		RxErrors:  3,
		RxPackets: 4,
		TxBytes:   5,
		TxDropped: 6,
		TxErrors:  7,
		TxPackets: 8,
	}
	oldNetworkData[containerId]["em1"] = calculator.NetworkData{
		Time:      oldTimestamp,
		RxBytes:   9,
		RxDropped: 10,
		RxErrors:  11,
		RxPackets: 12,
		TxBytes:   13,
		TxDropped: 14,
		TxErrors:  15,
		TxPackets: 16,
	}

	// mocking calculators
	// first - generate expected calls (NetworkStats to NetworkData conversion)
	newNetworkData := map[string]calculator.NetworkData{}
	newNetworkData["eth0"] = calculator.NetworkData{
		Time:      newTimestamp,
		RxBytes:   10,
		RxDropped: 20,
		RxErrors:  30,
		RxPackets: 40,
		TxBytes:   50,
		TxDropped: 60,
		TxErrors:  70,
		TxPackets: 80,
	}
	newNetworkData["em1"] = calculator.NetworkData{
		Time:      newTimestamp,
		RxBytes:   90,
		RxDropped: 100,
		RxErrors:  110,
		RxPackets: 120,
		TxBytes:   130,
		TxDropped: 140,
		TxErrors:  150,
		TxPackets: 160,
	}

	// second - instantiate mock
	mockedCalculatorFactory := new(calculator.MockedCalculatorFactory)
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
			"dockerSocket":  &socket,
			"net": common.MapStr{
				"name":         "eth0",
				"rxBytes_ps":   mockedNetworkCalculatorEth0.GetRxBytesPerSecond(),
				"rxDropped_ps": mockedNetworkCalculatorEth0.GetRxDroppedPerSecond(),
				"rxErrors_ps":  mockedNetworkCalculatorEth0.GetRxErrorsPerSecond(),
				"rxPackets_ps": mockedNetworkCalculatorEth0.GetRxPacketsPerSecond(),
				"txBytes_ps":   mockedNetworkCalculatorEth0.GetTxBytesPerSecond(),
				"txDropped_ps": mockedNetworkCalculatorEth0.GetTxDroppedPerSecond(),
				"txErrors_ps":  mockedNetworkCalculatorEth0.GetTxErrorsPerSecond(),
				"txPackets_ps": mockedNetworkCalculatorEth0.GetTxPacketsPerSecond(),
			}},
		common.MapStr{
			"@timestamp":    common.Time(newTimestamp),
			"type":          "net",
			"containerID":   container.ID,
			"containerName": "name1",
			"dockerSocket":  &socket,
			"net": common.MapStr{
				"name":         "em1",
				"rxBytes_ps":   mockedNetworkCalculatorEm1.GetRxBytesPerSecond(),
				"rxDropped_ps": mockedNetworkCalculatorEm1.GetRxDroppedPerSecond(),
				"rxErrors_ps":  mockedNetworkCalculatorEm1.GetRxErrorsPerSecond(),
				"rxPackets_ps": mockedNetworkCalculatorEm1.GetRxPacketsPerSecond(),
				"txBytes_ps":   mockedNetworkCalculatorEm1.GetTxBytesPerSecond(),
				"txDropped_ps": mockedNetworkCalculatorEm1.GetTxDroppedPerSecond(),
				"txErrors_ps":  mockedNetworkCalculatorEm1.GetTxErrorsPerSecond(),
				"txPackets_ps": mockedNetworkCalculatorEm1.GetTxPacketsPerSecond(),
			}})

	// the eventGenerator to test
	var eventGenerator = EventGenerator{&socket, EGNetworkStats{M: oldNetworkData}, EGBlkioStats{}, mockedCalculatorFactory, period}

	// WHEN
	events := eventGenerator.GetNetworksEvent(&container, stats)

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
	assert.Equal(t, eventGenerator.NetworkStats.M[container.ID]["eth0"], newNetworkData["eth0"])
	assert.Equal(t, eventGenerator.NetworkStats.M[container.ID]["em1"], newNetworkData["em1"])
}

/*
TestEventGeneratorGetNetworksEvent simulates the case when a saved network should be cleaned from saved status

It simulates following status:
  - a common container
  - network stats with one network "eth0"

Networks "eth0" have an saved status from previous tick.
An existing saved status for "em1" network is too old and should be removed.

This test checks that it generate one network event:
  - an event for "eth0" with calculated data from saved stats (+ new stats saved)
  - the "em1" saved network status should be removed
*/
func TestEventGeneratorGetNetworksEventCleanSavedEvents(t *testing.T) {
	// GIVEN
	// docker socket
	socket := "unix:///some/docker/socket"

	// old and current timestamps
	oldTimestamp := time.Now()
	veryOldTimestamp := oldTimestamp.AddDate(0, -1, 0)
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
		RxBytes:   10,
		RxDropped: 20,
		RxErrors:  30,
		RxPackets: 40,
		TxBytes:   50,
		TxDropped: 60,
		TxErrors:  70,
		TxPackets: 80,
	}

	// main stats object
	var stats = new(docker.Stats)
	stats.Read = newTimestamp
	stats.Networks = networkStatsMap

	// saved network status
	oldNetworkData := map[string]map[string]calculator.NetworkData{}
	oldNetworkData[containerId] = map[string]calculator.NetworkData{}
	oldNetworkData[containerId]["eth0"] = calculator.NetworkData{
		Time:      oldTimestamp,
		RxBytes:   1,
		RxDropped: 2,
		RxErrors:  3,
		RxPackets: 4,
		TxBytes:   5,
		TxDropped: 6,
		TxErrors:  7,
		TxPackets: 8,
	}
	// em1 has a very old timestamp, and should be removed because no em1 event come from stats API
	oldNetworkData[containerId]["em1"] = calculator.NetworkData{
		Time:      veryOldTimestamp,
		RxBytes:   9,
		RxDropped: 10,
		RxErrors:  11,
		RxPackets: 12,
		TxBytes:   13,
		TxDropped: 14,
		TxErrors:  15,
		TxPackets: 16,
	}

	// mocking calculators
	// first - generate expected calls (NetworkStats to NetworkData conversion)
	newNetworkData := map[string]calculator.NetworkData{}
	newNetworkData["eth0"] = calculator.NetworkData{
		Time:      newTimestamp,
		RxBytes:   10,
		RxDropped: 20,
		RxErrors:  30,
		RxPackets: 40,
		TxBytes:   50,
		TxDropped: 60,
		TxErrors:  70,
		TxPackets: 80,
	}

	// second - instantiate mock
	mockedCalculatorFactory := new(calculator.MockedCalculatorFactory)
	mockedNetworkCalculatorEth0 := getMockedNetworkCalculator(1.0)
	mockedCalculatorFactory.On("newNetworkCalculator", oldNetworkData[containerId]["eth0"], newNetworkData["eth0"]).Return(mockedNetworkCalculatorEth0)

	// expected events
	expectedEvents := []common.MapStr{}
	expectedEvents = append(expectedEvents,
		common.MapStr{
			"@timestamp":    common.Time(newTimestamp),
			"type":          "net",
			"containerID":   container.ID,
			"containerName": "name1",
			"dockerSocket":  &socket,
			"net": common.MapStr{
				"name":         "eth0",
				"rxBytes_ps":   mockedNetworkCalculatorEth0.GetRxBytesPerSecond(),
				"rxDropped_ps": mockedNetworkCalculatorEth0.GetRxDroppedPerSecond(),
				"rxErrors_ps":  mockedNetworkCalculatorEth0.GetRxErrorsPerSecond(),
				"rxPackets_ps": mockedNetworkCalculatorEth0.GetRxPacketsPerSecond(),
				"txBytes_ps":   mockedNetworkCalculatorEth0.GetTxBytesPerSecond(),
				"txDropped_ps": mockedNetworkCalculatorEth0.GetTxDroppedPerSecond(),
				"txErrors_ps":  mockedNetworkCalculatorEth0.GetTxErrorsPerSecond(),
				"txPackets_ps": mockedNetworkCalculatorEth0.GetTxPacketsPerSecond(),
			}})

	// the eventGenerator to test
	var eventGenerator = EventGenerator{&socket, EGNetworkStats{M: oldNetworkData}, EGBlkioStats{}, mockedCalculatorFactory, period}

	// WHEN
	events := eventGenerator.GetNetworksEvent(&container, stats)

	// THEN
	// check returned events
	assert.Equal(t, expectedEvents, events)

	// check that new stats saved
	assert.Equal(t, eventGenerator.NetworkStats.M[container.ID]["eth0"], newNetworkData["eth0"])

	// check that expired state has been deleted
	_, ok := eventGenerator.NetworkStats.M[container.ID]["em1"]
	if ok {
		assert.Fail(t, "Expired event has not been deleted")
	}
}

// CONTAINER EVENT GENERATION

/*
TestEventGeneratorGetContainerEvent simulates the case when a container event is generated

This test checks that the generated event is well formatted according to the incoming container stats.
*/

func TestEventGeneratorGetContainerEvent(t *testing.T) {
	// GIVEN
	// docker socket
	socket := "unix:///some/docker/socket"

	labels := make(map[string]string)
	labels["label1"] = "value1"
	labels["label2"] = "value2"
	labels["label3.with.dots"] = "value3"
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
	var eventGenerator = EventGenerator{&socket, EGNetworkStats{}, EGBlkioStats{}, calculator.CalculatorFactoryImpl{}, time.Second}

	// expected output
	// sanitized lables expected
	labels_expected := make(map[string]string)
	labels_expected["label1"] = "value1"
	labels_expected["label2"] = "value2"
	labels_expected["label3_with_dots"] = "value3"

	expectedEvent := common.MapStr{
		"@timestamp":    common.Time(timestamp),
		"type":          "container",
		"containerID":   container.ID,
		"containerName": "name1",
		"dockerSocket":  &socket,
		"container": common.MapStr{
			"id":      container.ID,
			"command": container.Command,
			"created": time.Unix(container.Created, 0),
			"image":   container.Image,
			"labels":  labels_expected,
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
	event := eventGenerator.GetContainerEvent(&container, stats)

	// THEN
	assert.Equal(t, expectedEvent, event)
}

func TestEventGeneratorGetContainerEventWithNoPorts(t *testing.T) {
	// GIVEN
	// docker socket
	socket := "unix:///some/docker/socket"

	labels := make(map[string]string)
	labels["label1"] = "value1"
	labels["label2"] = "value2"
	labels["label3.with.dots"] = "value3"
	var container = docker.APIContainers{
		"container_id",
		"container_image",
		"container command",
		9876543210,
		"Up",
		[]docker.APIPort{}, // no port
		123,
		456,
		[]string{"/name1", "name1/fake"},
		labels,
	}

	timestamp := time.Now()
	var stats = new(docker.Stats)
	stats.Read = timestamp
	var eventGenerator = EventGenerator{&socket, EGNetworkStats{}, EGBlkioStats{}, calculator.CalculatorFactoryImpl{}, time.Second}

	// expected output
	// sanitized lables expected
	labels_expected := make(map[string]string)
	labels_expected["label1"] = "value1"
	labels_expected["label2"] = "value2"
	labels_expected["label3_with_dots"] = "value3"

	expectedEvent := common.MapStr{
		"@timestamp":    common.Time(timestamp),
		"type":          "container",
		"containerID":   container.ID,
		"containerName": "name1",
		"dockerSocket":  &socket,
		"container": common.MapStr{
			"id":         container.ID,
			"command":    container.Command,
			"created":    time.Unix(container.Created, 0),
			"image":      container.Image,
			"labels":     labels_expected,
			"names":      container.Names,
			"ports":      []map[string]interface{}{},
			"sizeRootFs": container.SizeRootFs,
			"sizeRw":     container.SizeRw,
			"status":     container.Status,
		},
	}

	// WHEN
	event := eventGenerator.GetContainerEvent(&container, stats)

	// THEN
	assert.Equal(t, expectedEvent, event)
}

// CPU EVENT GENERATION

/*
TestEventGeneratorGetCpuEventFirstPass simulates the case when a cpu event should be generated

It simulates following status:
  - a common container
  - a common CPU stats

This test checks parameters passed to the calculator and checks that the event generated is well formatted.
*/
func TestEventGeneratorGetCpuEvent(t *testing.T) {
	// GIVEN
	// docker socket
	socket := "unix:///some/docker/socket"

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

	// CPU stats from Docker API
	preCPUStats := getCPUStats(1)
	cpuStats := getCPUStats(2)

	// main stats object
	var stats = new(docker.Stats)
	stats.Read = time.Now()
	stats.CPUStats = cpuStats
	stats.PreCPUStats = preCPUStats

	// mocking calculator
	// first - generate expected calls (CPUStats to CPUData conversion)
	cpuData := calculator.CPUData{
		PerCpuUsage:       cpuStats.CPUUsage.PercpuUsage,
		TotalUsage:        cpuStats.CPUUsage.TotalUsage,
		UsageInKernelmode: cpuStats.CPUUsage.UsageInKernelmode,
		UsageInUsermode:   cpuStats.CPUUsage.UsageInUsermode,
	}

	preCPUData := calculator.CPUData{
		PerCpuUsage:       preCPUStats.CPUUsage.PercpuUsage,
		TotalUsage:        preCPUStats.CPUUsage.TotalUsage,
		UsageInKernelmode: preCPUStats.CPUUsage.UsageInKernelmode,
		UsageInUsermode:   preCPUStats.CPUUsage.UsageInUsermode,
	}

	// second - instantiate mock
	// calculator will no be called for em1 network, it will generate zero-values event for em1
	mockedCalculatorFactory := new(calculator.MockedCalculatorFactory)
	mockedCPUCalculator := getMockedCPUCalculator(1.0)
	mockedCalculatorFactory.On("newCPUCalculator", preCPUData, cpuData).Return(mockedCPUCalculator)

	// expected events
	expectedEvent := common.MapStr{
		"@timestamp":    common.Time(stats.Read),
		"type":          "cpu",
		"containerID":   container.ID,
		"containerName": "name1",
		"dockerSocket":  &socket,
		"cpu": common.MapStr{
			"percpuUsage":       mockedCPUCalculator.PerCpuUsage(),
			"totalUsage":        mockedCPUCalculator.TotalUsage(),
			"usageInKernelmode": mockedCPUCalculator.UsageInKernelmode(),
			"usageInUsermode":   mockedCPUCalculator.UsageInUsermode(),
		},
	}

	// the eventGenerator to test
	var eventGenerator = EventGenerator{&socket, EGNetworkStats{}, EGBlkioStats{}, mockedCalculatorFactory, time.Second}

	// WHEN
	event := eventGenerator.GetCpuEvent(&container, stats)

	// THEN
	// check returned events
	assert.Equal(t, expectedEvent, event)
}

// MEMORY EVENT GENERATION

/* TestEventGeneratorGetMemoryEvent simulates the case when a memory event should be generated

It checks the event format, according to the incoming memory stats
*/

func TestEventGeneratorGetMemoryEvent(t *testing.T) {
	// GIVEN
	// docker socket
	socket := "unix:///some/docker/socket"

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

	// main stats object
	var stats = getMemoryStats(time.Now(), 1)

	// expected events
	expectedEvent := common.MapStr{
		"@timestamp":    common.Time(stats.Read),
		"type":          "memory",
		"containerID":   container.ID,
		"containerName": "name1",
		"dockerSocket":  &socket,
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

	// the eventGenerator to test
	var eventGenerator = EventGenerator{&socket, EGNetworkStats{}, EGBlkioStats{}, nil, time.Second}

	// WHEN
	event := eventGenerator.GetMemoryEvent(&container, &stats)

	// THEN
	// check returned events
	assert.Equal(t, expectedEvent, event)
}

// BLKIO EVENT GENERATION

/*
TestEventGeneratorGetBlkioEventFirstPass simulates the case when a new blkio event should be generated

It simulates following status:
  - a common container
  - blkio stats without saved status

The blkio stats for this container already have an saved status from previous tick.

This test checks that it generate a well formatted Blkio stats event.
*/
func TestEventGeneratorGetBlkioEventFirstPass(t *testing.T) {
	// GIVEN
	// docker socket
	socket := "unix:///some/docker/socket"

	// old and current timestamps
	newTimestamp := time.Now()

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
		[]string{"name1", "name1/fake"},
		labels,
	}

	// main stats object
	var stats = getBlkioStats(newTimestamp, 10, 20, 30)

	// saved network status (blkio stats does not exist for this container)
	oldBlkioData := map[string]calculator.BlkioData{}

	// mocking calculators
	// first - generate expected calls (NetworkStats to NetworkData conversion)
	newBlkioData := calculator.BlkioData{
		Time:   newTimestamp,
		Reads:  10,
		Writes: 20,
		Totals: 30,
	}

	// second - instantiate mock
	// calculator will no be called, it will generate zero-values event

	// expected events
	expectedEvent := common.MapStr{
		"@timestamp":    common.Time(stats.Read),
		"type":          "blkio",
		"containerID":   container.ID,
		"containerName": "name1",
		"dockerSocket":  &socket,
		"blkio": common.MapStr{
			"read_ps":  float64(0),
			"write_ps": float64(0),
			"total_ps": float64(0),
		},
	}

	// the eventGenerator to test
	var eventGenerator = EventGenerator{&socket, EGNetworkStats{}, EGBlkioStats{M: oldBlkioData}, nil, time.Second}

	// WHEN
	event := eventGenerator.GetBlkioEvent(&container, &stats)

	// THEN
	// check returned events
	assert.Equal(t, expectedEvent, event)

	assert.Equal(t, eventGenerator.BlkioStats.M[container.ID], newBlkioData)
}

/*
TestEventGeneratorGetBlkioEventFirstPass simulates the case when a new blkio event should be generated

It simulates following status:
  - a common container
  - blkio stats

The blkio stats for this container already have an saved status from previous tick.

This test checks that it generate a well formatted Blkio stats event.
*/
func TestEventGeneratorGetBlkioEvent(t *testing.T) {
	// GIVEN
	// docker socket
	socket := "unix:///some/docker/socket"

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

	// main stats object
	var stats = getBlkioStats(newTimestamp, 10, 20, 30)

	// saved network status
	oldBlkioData := map[string]calculator.BlkioData{}
	oldBlkioData[containerId] = calculator.BlkioData{
		Time:   oldTimestamp,
		Reads:  1,
		Writes: 2,
		Totals: 3,
	}

	// mocking calculators
	// first - generate expected calls (NetworkStats to NetworkData conversion)
	newBlkioData := calculator.BlkioData{
		Time:   newTimestamp,
		Reads:  10,
		Writes: 20,
		Totals: 30,
	}

	// second - instantiate mock
	mockedCalculatorFactory := new(calculator.MockedCalculatorFactory)
	mockedBlkioCalculator := getMockedBlkioCalculator(1)
	mockedCalculatorFactory.On("newBlkioCalculator", oldBlkioData[containerId], newBlkioData).Return(mockedBlkioCalculator)

	// expected events
	expectedEvent := common.MapStr{
		"@timestamp":    common.Time(stats.Read),
		"type":          "blkio",
		"containerID":   container.ID,
		"containerName": "name1",
		"dockerSocket":  &socket,
		"blkio": common.MapStr{
			"read_ps":  mockedBlkioCalculator.GetReadPs(),
			"write_ps": mockedBlkioCalculator.GetWritePs(),
			"total_ps": mockedBlkioCalculator.GetTotalPs(),
		},
	}

	// the eventGenerator to test
	var eventGenerator = EventGenerator{&socket, EGNetworkStats{}, EGBlkioStats{M: oldBlkioData}, mockedCalculatorFactory, time.Second}

	// WHEN
	event := eventGenerator.GetBlkioEvent(&container, &stats)

	// THEN
	// check returned events
	assert.Equal(t, expectedEvent, event)

	// check that new stats saved
	assert.Equal(t, eventGenerator.BlkioStats.M[container.ID], newBlkioData)
}

/*
TestEventGeneratorGetBlkioEventCleanSavedEvents simulates the case when method should clean old blkio stats

It simulates following status:
  - a common container
  - blkio stats with saved status

The blkio stats for this container already have an saved status from previous tick.
A saved event is too old and should be remove from saved stats.
*/
func TestEventGeneratorGetBlkioEventCleanSavedEvents(t *testing.T) {
	// GIVEN
	// docker socket
	socket := "unix:///some/docker/socket"

	// old and current timestamps
	oldTimestamp := time.Now()
	veryOldTimestamp := oldTimestamp.AddDate(0, -1, 0)
	period := time.Second
	newTimestamp := oldTimestamp.Add(period)

	// a container
	labels := make(map[string]string)
	labels["label1"] = "value1"
	labels["label2"] = "value2"
	containerId := "container_id"
	anotherContainerId := "container_id2"
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

	// main stats object
	var stats = getBlkioStats(newTimestamp, 10, 20, 30)

	// saved blkio stats
	oldBlkioData := map[string]calculator.BlkioData{}
	oldBlkioData[containerId] = calculator.BlkioData{
		Time:   oldTimestamp,
		Reads:  1,
		Writes: 2,
		Totals: 3,
	}
	// another container has a very old blkio stats
	oldBlkioData[anotherContainerId] = calculator.BlkioData{
		Time:   veryOldTimestamp,
		Reads:  4,
		Writes: 5,
		Totals: 6,
	}

	// mocking calculators
	// first - generate expected calls (BlkioStats to BlkioData conversion)
	newBlkioData := calculator.BlkioData{
		Time:   newTimestamp,
		Reads:  10,
		Writes: 20,
		Totals: 30,
	}

	// second - instantiate mock
	mockedCalculatorFactory := new(calculator.MockedCalculatorFactory)
	mockedBlkioCalculator := getMockedBlkioCalculator(1)
	mockedCalculatorFactory.On("newBlkioCalculator", oldBlkioData[containerId], newBlkioData).Return(mockedBlkioCalculator)

	// expected events
	expectedEvent := common.MapStr{
		"@timestamp":    common.Time(newTimestamp),
		"type":          "blkio",
		"containerID":   container.ID,
		"containerName": "name1",
		"dockerSocket":  &socket,
		"blkio": common.MapStr{
			"read_ps":  mockedBlkioCalculator.GetReadPs(),
			"write_ps": mockedBlkioCalculator.GetWritePs(),
			"total_ps": mockedBlkioCalculator.GetTotalPs(),
		},
	}

	// the eventGenerator to test
	var eventGenerator = EventGenerator{&socket, EGNetworkStats{}, EGBlkioStats{M: oldBlkioData}, mockedCalculatorFactory, period}

	// WHEN
	event := eventGenerator.GetBlkioEvent(&container, &stats)

	// THEN
	// check returned events
	assert.Equal(t, expectedEvent, event)

	// check that new stats saved
	assert.Equal(t, eventGenerator.BlkioStats.M[container.ID], newBlkioData)

	// check that expired state has been deleted
	_, ok := eventGenerator.BlkioStats.M[anotherContainerId]
	if ok {
		assert.Fail(t, "Expired event has not been deleted")
	}
}

// DAEMON EVENT GENERATION

/*
TestEventGeneratorGetLogEvent check that a well formatted event is generated from a level and message.
*/
func TestEventGeneratorGetLogEvent(t *testing.T) {
	// GIVEN
	// an error
	message := "Some error message"
	level := "Some level"

	// docker socket
	socket := "unix:///some/docker/socket"

	// expected event
	expectedEvent := common.MapStr{
		"@timestamp":   nil,
		"type":         "log",
		"dockerSocket": &socket,
		"log": common.MapStr{
			"level":   level,
			"message": message,
		},
	}

	// the eventGenerator to test
	var eventGenerator = EventGenerator{&socket, EGNetworkStats{}, EGBlkioStats{}, nil, time.Second}

	// WHEN
	event := eventGenerator.GetLogEvent(level, message)

	// get the event time and set value to the expectedEvent
	expectedEvent["@timestamp"] = event["@timestamp"]

	// THEN
	// check returned events
	assert.Equal(t, expectedEvent, event)
}

// NEEDED TYPES

type MemoryStats struct {
	Stats    struct {
			 TotalPgmafault          uint64 `json:"total_pgmafault,omitempty" yaml:"total_pgmafault,omitempty"`
			 Cache                   uint64 `json:"cache,omitempty" yaml:"cache,omitempty"`
			 MappedFile              uint64 `json:"mapped_file,omitempty" yaml:"mapped_file,omitempty"`
			 TotalInactiveFile       uint64 `json:"total_inactive_file,omitempty" yaml:"total_inactive_file,omitempty"`
			 Pgpgout                 uint64 `json:"pgpgout,omitempty" yaml:"pgpgout,omitempty"`
			 Rss                     uint64 `json:"rss,omitempty" yaml:"rss,omitempty"`
			 TotalMappedFile         uint64 `json:"total_mapped_file,omitempty" yaml:"total_mapped_file,omitempty"`
			 Writeback               uint64 `json:"writeback,omitempty" yaml:"writeback,omitempty"`
			 Unevictable             uint64 `json:"unevictable,omitempty" yaml:"unevictable,omitempty"`
			 Pgpgin                  uint64 `json:"pgpgin,omitempty" yaml:"pgpgin,omitempty"`
			 TotalUnevictable        uint64 `json:"total_unevictable,omitempty" yaml:"total_unevictable,omitempty"`
			 Pgmajfault              uint64 `json:"pgmajfault,omitempty" yaml:"pgmajfault,omitempty"`
			 TotalRss                uint64 `json:"total_rss,omitempty" yaml:"total_rss,omitempty"`
			 TotalRssHuge            uint64 `json:"total_rss_huge,omitempty" yaml:"total_rss_huge,omitempty"`
			 TotalWriteback          uint64 `json:"total_writeback,omitempty" yaml:"total_writeback,omitempty"`
			 TotalInactiveAnon       uint64 `json:"total_inactive_anon,omitempty" yaml:"total_inactive_anon,omitempty"`
			 RssHuge                 uint64 `json:"rss_huge,omitempty" yaml:"rss_huge,omitempty"`
			 HierarchicalMemoryLimit uint64 `json:"hierarchical_memory_limit,omitempty" yaml:"hierarchical_memory_limit,omitempty"`
			 TotalPgfault            uint64 `json:"total_pgfault,omitempty" yaml:"total_pgfault,omitempty"`
			 TotalActiveFile         uint64 `json:"total_active_file,omitempty" yaml:"total_active_file,omitempty"`
			 ActiveAnon              uint64 `json:"active_anon,omitempty" yaml:"active_anon,omitempty"`
			 TotalActiveAnon         uint64 `json:"total_active_anon,omitempty" yaml:"total_active_anon,omitempty"`
			 TotalPgpgout            uint64 `json:"total_pgpgout,omitempty" yaml:"total_pgpgout,omitempty"`
			 TotalCache              uint64 `json:"total_cache,omitempty" yaml:"total_cache,omitempty"`
			 InactiveAnon            uint64 `json:"inactive_anon,omitempty" yaml:"inactive_anon,omitempty"`
			 ActiveFile              uint64 `json:"active_file,omitempty" yaml:"active_file,omitempty"`
			 Pgfault                 uint64 `json:"pgfault,omitempty" yaml:"pgfault,omitempty"`
			 InactiveFile            uint64 `json:"inactive_file,omitempty" yaml:"inactive_file,omitempty"`
			 TotalPgpgin             uint64 `json:"total_pgpgin,omitempty" yaml:"total_pgpgin,omitempty"`
		 } `json:"stats,omitempty" yaml:"stats,omitempty"`
	MaxUsage uint64 `json:"max_usage,omitempty" yaml:"max_usage,omitempty"`
	Usage    uint64 `json:"usage,omitempty" yaml:"usage,omitempty"`
	Failcnt  uint64 `json:"failcnt,omitempty" yaml:"failcnt,omitempty"`
	Limit    uint64 `json:"limit,omitempty" yaml:"limit,omitempty"`
}

// UTILITY METHODS

func getMockedNetworkCalculator(number float64) *calculator.MockedNetworkCalculator {
	mock := new(calculator.MockedNetworkCalculator)
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

func getCPUStats(number uint64) docker.CPUStats {
	return docker.CPUStats{
		CPUUsage: struct {
			PercpuUsage       []uint64 `json:"percpu_usage,omitempty" yaml:"percpu_usage,omitempty"`
			UsageInUsermode   uint64   `json:"usage_in_usermode,omitempty" yaml:"usage_in_usermode,omitempty"`
			TotalUsage        uint64   `json:"total_usage,omitempty" yaml:"total_usage,omitempty"`
			UsageInKernelmode uint64   `json:"usage_in_kernelmode,omitempty" yaml:"usage_in_kernelmode,omitempty"`
		}{
			PercpuUsage:       []uint64{number, number * 2, number * 3, number * 4},
			UsageInUsermode:   number * 5,
			TotalUsage:        number * 6,
			UsageInKernelmode: number * 7,
		},
		SystemCPUUsage: number * 8,
		ThrottlingData: struct {
			Periods          uint64 `json:"periods,omitempty"`
			ThrottledPeriods uint64 `json:"throttled_periods,omitempty"`
			ThrottledTime    uint64 `json:"throttled_time,omitempty"`
		}{
			Periods:          number * 9,
			ThrottledPeriods: number * 10,
			ThrottledTime:    number * 11,
		},
	}
}

func getMockedCPUCalculator(number float64) calculator.CPUCalculator {
	mock := new(calculator.MockedCPUCalculator)
	perCPUUsage := common.MapStr{
		"cpu0": number,
		"cpu1": number,
		"cpu2": number,
		"cpu3": number,
	}
	mock.On("perCpuUsage").Return(perCPUUsage)
	mock.On("totalUsage").Return(number * 2)
	mock.On("usageInKernelmode").Return(number * 3)
	mock.On("usageInUsermode").Return(number * 4)
	mock.On("calculateLoad").Return(number * 5)

	return mock
}

func getMemoryStats(read time.Time, number uint64) docker.Stats {
	type memoryStats struct {
		Stats    struct {
				 TotalPgmafault          uint64 `json:"total_pgmafault,omitempty" yaml:"total_pgmafault,omitempty"`
				 Cache                   uint64 `json:"cache,omitempty" yaml:"cache,omitempty"`
				 MappedFile              uint64 `json:"mapped_file,omitempty" yaml:"mapped_file,omitempty"`
				 TotalInactiveFile       uint64 `json:"total_inactive_file,omitempty" yaml:"total_inactive_file,omitempty"`
				 Pgpgout                 uint64 `json:"pgpgout,omitempty" yaml:"pgpgout,omitempty"`
				 Rss                     uint64 `json:"rss,omitempty" yaml:"rss,omitempty"`
				 TotalMappedFile         uint64 `json:"total_mapped_file,omitempty" yaml:"total_mapped_file,omitempty"`
				 Writeback               uint64 `json:"writeback,omitempty" yaml:"writeback,omitempty"`
				 Unevictable             uint64 `json:"unevictable,omitempty" yaml:"unevictable,omitempty"`
				 Pgpgin                  uint64 `json:"pgpgin,omitempty" yaml:"pgpgin,omitempty"`
				 TotalUnevictable        uint64 `json:"total_unevictable,omitempty" yaml:"total_unevictable,omitempty"`
				 Pgmajfault              uint64 `json:"pgmajfault,omitempty" yaml:"pgmajfault,omitempty"`
				 TotalRss                uint64 `json:"total_rss,omitempty" yaml:"total_rss,omitempty"`
				 TotalRssHuge            uint64 `json:"total_rss_huge,omitempty" yaml:"total_rss_huge,omitempty"`
				 TotalWriteback          uint64 `json:"total_writeback,omitempty" yaml:"total_writeback,omitempty"`
				 TotalInactiveAnon       uint64 `json:"total_inactive_anon,omitempty" yaml:"total_inactive_anon,omitempty"`
				 RssHuge                 uint64 `json:"rss_huge,omitempty" yaml:"rss_huge,omitempty"`
				 HierarchicalMemoryLimit uint64 `json:"hierarchical_memory_limit,omitempty" yaml:"hierarchical_memory_limit,omitempty"`
				 TotalPgfault            uint64 `json:"total_pgfault,omitempty" yaml:"total_pgfault,omitempty"`
				 TotalActiveFile         uint64 `json:"total_active_file,omitempty" yaml:"total_active_file,omitempty"`
				 ActiveAnon              uint64 `json:"active_anon,omitempty" yaml:"active_anon,omitempty"`
				 TotalActiveAnon         uint64 `json:"total_active_anon,omitempty" yaml:"total_active_anon,omitempty"`
				 TotalPgpgout            uint64 `json:"total_pgpgout,omitempty" yaml:"total_pgpgout,omitempty"`
				 TotalCache              uint64 `json:"total_cache,omitempty" yaml:"total_cache,omitempty"`
				 InactiveAnon            uint64 `json:"inactive_anon,omitempty" yaml:"inactive_anon,omitempty"`
				 ActiveFile              uint64 `json:"active_file,omitempty" yaml:"active_file,omitempty"`
				 Pgfault                 uint64 `json:"pgfault,omitempty" yaml:"pgfault,omitempty"`
				 InactiveFile            uint64 `json:"inactive_file,omitempty" yaml:"inactive_file,omitempty"`
				 TotalPgpgin             uint64 `json:"total_pgpgin,omitempty" yaml:"total_pgpgin,omitempty"`
			 } `json:"stats,omitempty" yaml:"stats,omitempty"`
		MaxUsage uint64 `json:"max_usage,omitempty" yaml:"max_usage,omitempty"`
		Usage    uint64 `json:"usage,omitempty" yaml:"usage,omitempty"`
		Failcnt  uint64 `json:"failcnt,omitempty" yaml:"failcnt,omitempty"`
		Limit    uint64 `json:"limit,omitempty" yaml:"limit,omitempty"`
	}

	testStats := docker.Stats{
		Read: read,
		MemoryStats: memoryStats{
			MaxUsage: number,
			Usage:    number * 2,
			Failcnt:  number * 3,
			Limit:    number * 4,
		},
	}

	testStats.MemoryStats.Stats.TotalRss = number * 5

	return testStats
}

func getMockedBlkioCalculator(number float64) *calculator.MockedBlkioCalculator {
	mock := new(calculator.MockedBlkioCalculator)
	mock.On("getReadPs").Return(number)
	mock.On("getWritePs").Return(number * 2)
	mock.On("getTotalPs").Return(number * 3)
	return mock
}

func getBlkioStats(read time.Time, reads uint64, writes uint64, total uint64) docker.Stats {
	type blkioStats struct {
		IOServiceBytesRecursive []docker.BlkioStatsEntry `json:"io_service_bytes_recursive,omitempty" yaml:"io_service_bytes_recursive,omitempty"`
		IOServicedRecursive     []docker.BlkioStatsEntry `json:"io_serviced_recursive,omitempty" yaml:"io_serviced_recursive,omitempty"`
		IOQueueRecursive        []docker.BlkioStatsEntry `json:"io_queue_recursive,omitempty" yaml:"io_queue_recursive,omitempty"`
		IOServiceTimeRecursive  []docker.BlkioStatsEntry `json:"io_service_time_recursive,omitempty" yaml:"io_service_time_recursive,omitempty"`
		IOWaitTimeRecursive     []docker.BlkioStatsEntry `json:"io_wait_time_recursive,omitempty" yaml:"io_wait_time_recursive,omitempty"`
		IOMergedRecursive       []docker.BlkioStatsEntry `json:"io_merged_recursive,omitempty" yaml:"io_merged_recursive,omitempty"`
		IOTimeRecursive         []docker.BlkioStatsEntry `json:"io_time_recursive,omitempty" yaml:"io_time_recursive,omitempty"`
		SectorsRecursive        []docker.BlkioStatsEntry `json:"sectors_recursive,omitempty" yaml:"sectors_recursive,omitempty"`
	}
	return docker.Stats{
		Read: read,
		BlkioStats: blkioStats{
			IOServicedRecursive: []docker.BlkioStatsEntry{
				docker.BlkioStatsEntry{0, 0, "Read", reads},
				docker.BlkioStatsEntry{0, 0, "Write", writes},
				docker.BlkioStatsEntry{0, 0, "Total", total},
			},
		},
	}
}
