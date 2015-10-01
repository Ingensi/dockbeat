package main

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/fsouza/go-dockerclient"
	"github.com/elastic/libbeat/common"
	"github.com/stretchr/testify/mock"
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
		[]string{"name1", "name2"},
		labels,
	}

	timestamp := time.Now()
	var stats = new(docker.Stats)
	stats.Read = timestamp
	var eventGenerator = EventGenerator{nil}

	// expected output
	expectedEvent := common.MapStr{
		"timestamp":      common.Time(timestamp),
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
			"ports":      []map[string]interface{}{common.MapStr{
				"ip": container.Ports[0].IP,
				"privatePort": container.Ports[0].PrivatePort,
				"publicPort": container.Ports[0].PublicPort,
				"type": container.Ports[0].Type,
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