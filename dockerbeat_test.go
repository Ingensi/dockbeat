package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"time"
	"github.com/fsouza/go-dockerclient"
)

type MockedDockerClient struct {
	mock.Mock
}

type MockedEventPublisher struct {
	mock.Mock
}

func TestDockerbeatRun(t *testing.T) {
	// GIVEN
	dockerClient := new(MockedDockerClient)

	// mock list container method
	apiContainers := []docker.APIContainers{docker.APIContainers{
		"123456789",
		"image_name",
		"container command",
		9876543210,
		"Up for 2 days",
		[]docker.APIPort{},
		0,
		0,
		[]string{"name1", "name2"},


	}}
	dockerClient.On("ListContainers").Return(apiContainers, nil)

	networkStats := make(map[string]NetworkData)
	events := new(MockedEventPublisher)
	dockerBeat := Dockerbeat{
		false,
		time.Duration(1000000000),
		"/path/to/docker/socket",
		nil,
		dockerClient,
		networkStats,
		events,
	}

	// WHEN
	dockerBeat.Run(nil)

	// THEN



}