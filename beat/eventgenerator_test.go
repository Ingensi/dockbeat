package beat

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type MockedStats struct {
	mock.Mock
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
	var eventGenerator = EventGenerator{nil, nil}

	// expected output
	expectedEvent := common.MapStr{
		"timestamp":     common.Time(timestamp),
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
	var eventGenerator = EventGenerator{nil, nil}
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
	var eventGenerator = EventGenerator{nil, nil}
	expectedName := "containerName"

	// WHEN
	name := eventGenerator.extractContainerName([]string{"/" + expectedName})

	// THEN
	assert.Equal(t, expectedName, name)
}

func TestExtractContainerNameMultiple(t *testing.T) {
	// GIVEN
	var eventGenerator = EventGenerator{nil, nil}
	expectedName := "containerName"

	// WHEN
	name := eventGenerator.extractContainerName([]string{"/name1/fake", "/" + expectedName, "/name3/fake"})

	// THEN
	assert.Equal(t, expectedName, name)
}
