package main

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

func (c *CPUCalculator) perCpuUsage() []uint64 {
	var output []uint64
	if cap(c.new.perCpuUsage) == cap(c.old.perCpuUsage) {
		output = make([]uint64, cap(c.new.perCpuUsage))
		for index := range c.new.perCpuUsage {
			output[index] = c.calculateLoad(c.new.perCpuUsage[index] - c.old.perCpuUsage[index])
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
	// TODO save the old stat timestamp and reuse here in case of docker read time changes...
	// value * 100 / 1000000000
	return value / 10000000
}