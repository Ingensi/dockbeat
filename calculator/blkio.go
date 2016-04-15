package calculator

import (
	"time"
)

type BlkioCalculator interface {
	GetReadPs() float64
	GetWritePs() float64
	GetTotalPs() float64
}

type BlkioCalculatorImpl struct {
	Old BlkioData
	New BlkioData
}

type BlkioData struct {
	Time   time.Time
	Reads  uint64
	Writes uint64
	Totals uint64
}

func (c BlkioCalculatorImpl) GetReadPs() float64 {
	return c.calculatePerSecond(c.Old.Reads, c.New.Reads)
}

func (c BlkioCalculatorImpl) GetWritePs() float64 {
	return c.calculatePerSecond(c.Old.Writes, c.New.Writes)
}

func (c BlkioCalculatorImpl) GetTotalPs() float64 {
	return c.calculatePerSecond(c.Old.Totals, c.New.Totals)
}

func (c BlkioCalculatorImpl) calculatePerSecond(oldValue uint64, newValue uint64) float64 {
	duration := c.New.Time.Sub(c.Old.Time)
	return float64(newValue-oldValue) / duration.Seconds()
}
