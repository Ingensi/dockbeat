package beat

import (
	"time"
)

type BlkioCalculator interface {
	getReadPs() float64
	getWritePs() float64
	getTotalPs() float64
}

type BlkioCalculatorImpl struct {
	old BlkioData
	new BlkioData
}

type BlkioData struct {
	time   time.Time
	reads  uint64
	writes uint64
	totals uint64
}

func (c BlkioCalculatorImpl) getReadPs() float64 {
	return c.calculatePerSecond(c.old.reads, c.new.reads)
}

func (c BlkioCalculatorImpl) getWritePs() float64 {
	return c.calculatePerSecond(c.old.writes, c.new.writes)
}

func (c BlkioCalculatorImpl) getTotalPs() float64 {
	return c.calculatePerSecond(c.old.totals, c.new.totals)
}

func (c BlkioCalculatorImpl) calculatePerSecond(oldValue uint64, newValue uint64) float64 {
	duration := c.new.time.Sub(c.old.time)
	return float64(newValue-oldValue) / duration.Seconds()
}
