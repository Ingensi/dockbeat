package beat

import (
	"time"
)

type NetworkCalculator interface {
	getRxBytesPerSecond() float64
	getRxDroppedPerSecond() float64
	getRxErrorsPerSecond() float64
	getRxPacketsPerSecond() float64
	getTxBytesPerSecond() float64
	getTxDroppedPerSecond() float64
	getTxErrorsPerSecond() float64
	getTxPacketsPerSecond() float64
}

type NetworkCalculatorImpl struct {
	old NetworkData
	new NetworkData
}

type NetworkData struct {
	time      time.Time
	rxBytes   uint64
	rxDropped uint64
	rxErrors  uint64
	rxPackets uint64
	txBytes   uint64
	txDropped uint64
	txErrors  uint64
	txPackets uint64
}

func (c NetworkCalculatorImpl) getRxBytesPerSecond() float64 {
	return c.calculatePerSecond(c.old.rxBytes, c.new.rxBytes)
}

func (c NetworkCalculatorImpl) getRxDroppedPerSecond() float64 {
	return c.calculatePerSecond(c.old.rxDropped, c.new.rxDropped)
}

func (c NetworkCalculatorImpl) getRxErrorsPerSecond() float64 {
	return c.calculatePerSecond(c.old.rxErrors, c.new.rxErrors)
}

func (c NetworkCalculatorImpl) getRxPacketsPerSecond() float64 {
	return c.calculatePerSecond(c.old.rxPackets, c.new.rxPackets)
}

func (c NetworkCalculatorImpl) getTxBytesPerSecond() float64 {
	return c.calculatePerSecond(c.old.txBytes, c.new.txBytes)
}

func (c NetworkCalculatorImpl) getTxDroppedPerSecond() float64 {
	return c.calculatePerSecond(c.old.txDropped, c.new.txDropped)
}

func (c NetworkCalculatorImpl) getTxErrorsPerSecond() float64 {
	return c.calculatePerSecond(c.old.txErrors, c.new.txErrors)
}

func (c NetworkCalculatorImpl) getTxPacketsPerSecond() float64 {
	return c.calculatePerSecond(c.old.txPackets, c.new.txPackets)
}

func (c NetworkCalculatorImpl) calculatePerSecond(oldValue uint64, newValue uint64) float64 {
	duration := c.new.time.Sub(c.old.time)
	return float64((newValue - oldValue)) / duration.Seconds()
}
