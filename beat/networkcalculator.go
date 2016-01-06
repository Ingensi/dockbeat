package beat

import (
	"os"
	"strconv"
	"time"
)

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

type NetworkCalculator struct {
	old NetworkData
	new NetworkData
}

func (c *NetworkCalculator) getRxBytesPerSecond() float64 {
	return c.calculatePerSecond(c.old.rxBytes, c.new.rxBytes)
}

func (c *NetworkCalculator) getRxDroppedPerSecond() float64 {
	return c.calculatePerSecond(c.old.rxDropped, c.new.rxDropped)
}

func (c *NetworkCalculator) getRxErrorsPerSecond() float64 {
	return c.calculatePerSecond(c.old.rxErrors, c.new.rxErrors)
}

func (c *NetworkCalculator) getRxPacketsPerSecond() float64 {
	return c.calculatePerSecond(c.old.rxPackets, c.new.rxPackets)
}

func (c *NetworkCalculator) getTxBytesPerSecond() float64 {
	return c.calculatePerSecond(c.old.txBytes, c.new.txBytes)
}

func (c *NetworkCalculator) getTxDroppedPerSecond() float64 {
	return c.calculatePerSecond(c.old.txDropped, c.new.txDropped)
}

func (c *NetworkCalculator) getTxErrorsPerSecond() float64 {
	return c.calculatePerSecond(c.old.txErrors, c.new.txErrors)
}

func (c *NetworkCalculator) getTxPacketsPerSecond() float64 {
	os.Stdout.WriteString("\nin")
	os.Stdout.WriteString("\n" + strconv.FormatUint(c.old.txPackets, 10))
	os.Stdout.WriteString("\n" + strconv.FormatUint(c.new.txPackets, 10))
	os.Stdout.WriteString("\nout")
	return c.calculatePerSecond(c.old.txPackets, c.new.txPackets)
}

func (c *NetworkCalculator) calculatePerSecond(oldValue uint64, newValue uint64) float64 {
	duration := c.new.time.Sub(c.old.time)
	return float64((newValue - oldValue)) / duration.Seconds()
}
