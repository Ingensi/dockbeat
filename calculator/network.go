package calculator

import (
	"time"
)

type NetworkCalculator interface {
	GetRxBytesPerSecond() float64
	GetRxDroppedPerSecond() float64
	GetRxErrorsPerSecond() float64
	GetRxPacketsPerSecond() float64
	GetTxBytesPerSecond() float64
	GetTxDroppedPerSecond() float64
	GetTxErrorsPerSecond() float64
	GetTxPacketsPerSecond() float64
}

type NetworkCalculatorImpl struct {
	old NetworkData
	new NetworkData
}

type NetworkData struct {
	Time      time.Time
	RxBytes   uint64
	RxDropped uint64
	RxErrors  uint64
	RxPackets uint64
	TxBytes   uint64
	TxDropped uint64
	TxErrors  uint64
	TxPackets uint64
}

func (c NetworkCalculatorImpl) GetRxBytesPerSecond() float64 {
	return c.calculatePerSecond(c.old.RxBytes, c.new.RxBytes)
}

func (c NetworkCalculatorImpl) GetRxDroppedPerSecond() float64 {
	return c.calculatePerSecond(c.old.RxDropped, c.new.RxDropped)
}

func (c NetworkCalculatorImpl) GetRxErrorsPerSecond() float64 {
	return c.calculatePerSecond(c.old.RxErrors, c.new.RxErrors)
}

func (c NetworkCalculatorImpl) GetRxPacketsPerSecond() float64 {
	return c.calculatePerSecond(c.old.RxPackets, c.new.RxPackets)
}

func (c NetworkCalculatorImpl) GetTxBytesPerSecond() float64 {
	return c.calculatePerSecond(c.old.TxBytes, c.new.TxBytes)
}

func (c NetworkCalculatorImpl) GetTxDroppedPerSecond() float64 {
	return c.calculatePerSecond(c.old.TxDropped, c.new.TxDropped)
}

func (c NetworkCalculatorImpl) GetTxErrorsPerSecond() float64 {
	return c.calculatePerSecond(c.old.TxErrors, c.new.TxErrors)
}

func (c NetworkCalculatorImpl) GetTxPacketsPerSecond() float64 {
	return c.calculatePerSecond(c.old.TxPackets, c.new.TxPackets)
}

func (c NetworkCalculatorImpl) calculatePerSecond(oldValue uint64, newValue uint64) float64 {
	duration := c.new.Time.Sub(c.old.Time)
	return float64((newValue - oldValue)) / duration.Seconds()
}
