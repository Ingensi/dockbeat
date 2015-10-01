package main

import (
	"github.com/elastic/libbeat/common"
	"strconv"
)

type CPUData struct {
	perCpuUsage       []uint64
	totalUsage        uint64
	usageInKernelmode uint64
	usageInUsermode   uint64
}

type CPUCalculator struct {
	old CPUData
	new CPUData
}

func (c *CPUCalculator) perCpuUsage() common.MapStr {
	var output common.MapStr
	if cap(c.new.perCpuUsage) == cap(c.old.perCpuUsage) {
		output = common.MapStr{}
		for index := range c.new.perCpuUsage {
			output["cpu" + strconv.Itoa(index)] = c.calculateLoad(c.new.perCpuUsage[index] - c.old.perCpuUsage[index])
		}
	}
	return output
}

func (c *CPUCalculator) totalUsage() uint64 {
	return c.calculateLoad(c.new.totalUsage - c.old.totalUsage)
}

func (c *CPUCalculator) usageInKernelmode() uint64 {
	return c.calculateLoad(c.new.usageInKernelmode - c.old.usageInKernelmode)
}

func (c *CPUCalculator) usageInUsermode() uint64 {
	return c.calculateLoad(c.new.usageInUsermode - c.old.usageInUsermode)
}

func (c *CPUCalculator) calculateLoad(value uint64) uint64 {
	// value is the count of CPU nanosecond in 1sec
	// value * 100 / 1000000000
	return value / 10000000
}