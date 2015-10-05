package main

import (
	"github.com/fsouza/go-dockerclient"
)

type BlkioCalculator struct {
	ioServiceBytes []docker.BlkioStatsEntry
}

func (c *BlkioCalculator) getRead() uint64 {
	var total uint64
	total = 0
	for _, iostat := range c.ioServiceBytes {
		if iostat.Op == "Read" {
			total += iostat.Value
		}
	}
	return total
}

func (c *BlkioCalculator) getWrite() uint64 {
	var total uint64
	total = 0
	for _, iostat := range c.ioServiceBytes {
		if iostat.Op == "Write" {
			total += iostat.Value
		}
	}
	return total
}

func (c *BlkioCalculator) getTotal() uint64 {
	var total uint64
	total = 0
	for _, iostat := range c.ioServiceBytes {
		if iostat.Op == "Total" {
			total += iostat.Value
		}
	}
	return total
}
