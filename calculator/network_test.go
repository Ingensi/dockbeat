package calculator

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNetworkGetRxBytesPerSecond(t *testing.T) {
	// GIVEN
	var oldDate = time.Now()
	// 1000000000 nanoseconds = 1 second
	var newDate = oldDate.Add(time.Duration(1000000000))

	// old rxBytes = 10
	var oldData = NetworkData{oldDate, 10, 0, 0, 0, 0, 0, 0, 0}
	// new rxBytes = 110
	var newData = NetworkData{newDate, 110, 0, 0, 0, 0, 0, 0, 0}
	var calculator = NetworkCalculatorImpl{oldData, newData}

	// WHEN
	value := calculator.GetRxBytesPerSecond()

	// THEN
	// value should be 100 bytes / second
	assert.Equal(t, float64(100), value)
}

func TestNetworkGetRxDroppedPerSecond(t *testing.T) {
	// GIVEN
	var oldDate = time.Now()
	// 2000000000 nanoseconds = 2 seconds
	var newDate = oldDate.Add(time.Duration(2000000000))

	// old rxDropped = 20
	var oldData = NetworkData{oldDate, 0, 20, 0, 0, 0, 0, 0, 0}
	// new rxDropped = 120
	var newData = NetworkData{newDate, 0, 120, 0, 0, 0, 0, 0, 0}
	var calculator = NetworkCalculatorImpl{oldData, newData}

	// WHEN
	value := calculator.GetRxDroppedPerSecond()

	// THEN
	// value should be 50 dropped / second
	assert.Equal(t, float64(50), value)
}

func TestNetworkGetRxErrorsPerSecond(t *testing.T) {
	// GIVEN
	var oldDate = time.Now()
	// 10000000000 nanoseconds = 10 seconds
	var newDate = oldDate.Add(time.Duration(10000000000))

	// old rxErrors = 30
	var oldData = NetworkData{oldDate, 0, 0, 30, 0, 0, 0, 0, 0}
	// new rxErrors = 31
	var newData = NetworkData{newDate, 0, 0, 31, 0, 0, 0, 0, 0}
	var calculator = NetworkCalculatorImpl{oldData, newData}

	// WHEN
	value := calculator.GetRxErrorsPerSecond()

	// THEN
	// value should be 0.1 error / second
	assert.Equal(t, float64(0.1), value)
}

func TestNetworkGetRxPacketsPerSecond(t *testing.T) {
	// GIVEN
	var oldDate = time.Now()
	// 10000000000 nanoseconds = 10 seconds
	var newDate = oldDate.Add(time.Duration(10000000000))

	// old rxPackets = 40
	var oldData = NetworkData{oldDate, 0, 0, 0, 40, 0, 0, 0, 0}
	// new rxErrors = 140
	var newData = NetworkData{newDate, 0, 0, 0, 140, 0, 0, 0, 0}
	var calculator = NetworkCalculatorImpl{oldData, newData}

	// WHEN
	value := calculator.GetRxPacketsPerSecond()

	// THEN
	// value should be 10 packets / second
	assert.Equal(t, float64(10), value)
}

func TestNetworkGetTxBytesPerSecond(t *testing.T) {
	// GIVEN
	var oldDate = time.Now()
	// 1000000000 nanoseconds = 1 second
	var newDate = oldDate.Add(time.Duration(1000000000))

	// old txBytes = 10
	var oldData = NetworkData{oldDate, 0, 0, 0, 0, 10, 0, 0, 0}
	// new txBytes = 10
	var newData = NetworkData{newDate, 0, 0, 0, 0, 10, 0, 0, 0}
	var calculator = NetworkCalculatorImpl{oldData, newData}

	// WHEN
	value := calculator.GetTxBytesPerSecond()

	// THEN
	// value should be 0 bytes / second
	assert.Equal(t, float64(0), value)
}

func TestNetworkGetTxDroppedPerSecond(t *testing.T) {
	// GIVEN
	var oldDate = time.Now()
	// 2000000000 nanoseconds = 2 seconds
	var newDate = oldDate.Add(time.Duration(2000000000))

	// old txDropped = 0
	var oldData = NetworkData{oldDate, 0, 0, 0, 0, 0, 0, 0, 0}
	// new txDropped = 0
	var newData = NetworkData{newDate, 0, 0, 0, 0, 0, 0, 0, 0}
	var calculator = NetworkCalculatorImpl{oldData, newData}

	// WHEN
	value := calculator.GetTxDroppedPerSecond()

	// THEN
	// value should be 0 dropped / second
	assert.Equal(t, float64(0), value)
}

func TestNetworkGetTxErrorsPerSecond(t *testing.T) {
	// GIVEN
	var oldDate = time.Now()
	// 10000000000 nanoseconds = 10 seconds
	var newDate = oldDate.Add(time.Duration(10000000000))

	// old txErrors = 70
	var oldData = NetworkData{oldDate, 0, 0, 0, 0, 0, 0, 70, 0}
	// new txErrors = 170
	var newData = NetworkData{newDate, 0, 0, 0, 0, 0, 0, 170, 0}
	var calculator = NetworkCalculatorImpl{oldData, newData}

	// WHEN
	value := calculator.GetTxErrorsPerSecond()

	// THEN
	// value should be 10 error / second
	assert.Equal(t, float64(10), value)
}

func TestNetworkGetTxPacketsPerSecond(t *testing.T) {
	// GIVEN
	var oldDate = time.Now()
	// 10000000000 nanoseconds = 10 seconds
	var newDate = oldDate.Add(time.Duration(10000000000))

	// old txPackets = 80
	var oldData = NetworkData{oldDate, 0, 0, 0, 0, 0, 0, 0, 80}
	// new txErrors = 92
	var newData = NetworkData{newDate, 0, 0, 0, 0, 0, 0, 0, 92}
	var calculator = NetworkCalculatorImpl{oldData, newData}

	// WHEN
	value := calculator.GetTxPacketsPerSecond()

	// THEN
	// value should be 0 packets / second when old value > new value
	assert.Equal(t, float64(1.2), value)
}
