package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/fsouza/go-dockerclient"
)

func TestBlkioRead(t *testing.T) {
	// GIVEN
	var data []docker.BlkioStatsEntry
	data = make([]docker.BlkioStatsEntry, 20, 20)
	data[0] = docker.BlkioStatsEntry{0, 0, "Read", 1000}
	data[1] = docker.BlkioStatsEntry{0, 0, "Write", 2000}
	data[2] = docker.BlkioStatsEntry{0, 0, "Total", 3000}
	data[3] = docker.BlkioStatsEntry{0, 1, "Read", 10}
	data[4] = docker.BlkioStatsEntry{0, 1, "Write", 20}
	data[5] = docker.BlkioStatsEntry{0, 1, "Total", 30}
	var calculator = BlkioCalculator{data}

	// WHEN
	value := calculator.getRead()

	// THEN
	// value should be 1000+10
	assert.Equal(t, uint64(1010), value)
}

func TestBlkioWrite(t *testing.T) {
	// GIVEN
	var data []docker.BlkioStatsEntry
	data = make([]docker.BlkioStatsEntry, 20, 20)
	data[0] = docker.BlkioStatsEntry{0, 0, "Read", 1000}
	data[1] = docker.BlkioStatsEntry{0, 0, "Write", 2000}
	data[2] = docker.BlkioStatsEntry{0, 0, "Total", 3000}
	data[3] = docker.BlkioStatsEntry{0, 1, "Read", 10}
	data[4] = docker.BlkioStatsEntry{0, 1, "Write", 20}
	data[5] = docker.BlkioStatsEntry{0, 1, "Total", 30}
	var calculator = BlkioCalculator{data}

	// WHEN
	value := calculator.getWrite()

	// THEN
	// value should be 2000+20
	assert.Equal(t, uint64(2020), value)
}

func TestBlkioTotal(t *testing.T) {
	// GIVEN
	var data []docker.BlkioStatsEntry
	data = make([]docker.BlkioStatsEntry, 20, 20)
	data[0] = docker.BlkioStatsEntry{0, 0, "Read", 1000}
	data[1] = docker.BlkioStatsEntry{0, 0, "Write", 2000}
	data[2] = docker.BlkioStatsEntry{0, 0, "Total", 3000}
	data[3] = docker.BlkioStatsEntry{0, 1, "Read", 10}
	data[4] = docker.BlkioStatsEntry{0, 1, "Write", 20}
	data[5] = docker.BlkioStatsEntry{0, 1, "Total", 30}
	var calculator = BlkioCalculator{data}

	// WHEN
	value := calculator.getTotal()

	// THEN
	// value should be 3000+30
	assert.Equal(t, uint64(3030), value)
}