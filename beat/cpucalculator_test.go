package beat

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCPUperCpuUsage(t *testing.T) {
	// GIVEN
	var oldData = CPUData{[]uint64{1, 2, 3, 4}, 0, 0, 0}
	var newData = CPUData{[]uint64{100000001, 200000002, 300000003, 400000004}, 0, 0, 0}
	var calculator = CPUCalculator{oldData, newData}

	// WHEN
	value := calculator.perCpuUsage()

	// THEN
	// value should be 10%, 20%, 30% and 40%
	assert.Equal(t, common.MapStr{
		"cpu0": float64 (0.10),
		"cpu1": float64 (0.20),
		"cpu2": float64 (0.30),
		"cpu3": float64 (0.40),
	}, value)
}

func TestCPUTotalUsage(t *testing.T) {
	// GIVEN
	var oldData = CPUData{nil, 50, 0, 0}
	var newData = CPUData{nil, 500000050, 0, 0}
	var calculator = CPUCalculator{oldData, newData}

	// WHEN
	value := calculator.totalUsage()

	// THEN
	// value should be 50%
	assert.Equal(t, 0.50, value)
}

func TestCPUUsageInKernelmode(t *testing.T) {
	// GIVEN
	var oldData = CPUData{nil, 0, 0, 0}
	var newData = CPUData{nil, 0, 800000000, 0}
	var calculator = CPUCalculator{oldData, newData}

	// WHEN
	value := calculator.usageInKernelmode()

	// THEN
	// value should be 80%
	assert.Equal(t, float64(0.80), value)
}

func TestCPUUsageInUsermode(t *testing.T) {
	// GIVEN
	var oldData = CPUData{nil, 0, 0, 800000000}
	var newData = CPUData{nil, 0, 0, 800000000}
	var calculator = CPUCalculator{oldData, newData}

	// WHEN
	value := calculator.usageInUsermode()

	// THEN
	// value should be 0%
	assert.Equal(t, float64(0), value)
}
