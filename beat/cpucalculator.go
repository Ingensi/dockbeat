package beat

import (
	"github.com/elastic/beats/libbeat/common"
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
			output["cpu"+strconv.Itoa(index)] = c.calculateLoad(c.new.perCpuUsage[index] - c.old.perCpuUsage[index])
		}
	}
	return output
}

func (c *CPUCalculator) totalUsage() float64  {
	return c.calculateLoad(c.new.totalUsage - c.old.totalUsage)
}

func (c *CPUCalculator) usageInKernelmode() float64 {
	return c.calculateLoad(c.new.usageInKernelmode - c.old.usageInKernelmode)
}

func (c *CPUCalculator) usageInUsermode() float64  {
	return c.calculateLoad(c.new.usageInUsermode - c.old.usageInUsermode)
}

func (c *CPUCalculator) calculateLoad(value uint64) float64 {
	// value is the count of CPU nanosecond in 1sec
	// TODO save the old stat timestamp and reuse here in case of docker read time changes...
	// 1s = 1000000000 ns
	// value / 1000000000
	return float64 (value) / float64 (1000000000)
}
