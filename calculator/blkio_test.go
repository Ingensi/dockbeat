package calculator

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBlkioRead(t *testing.T) {
	// GIVEN
	oldTimestamp := time.Now()
	newTimestamp := oldTimestamp.Add(2 * time.Second)

	old := BlkioData{
		Time:   oldTimestamp,
		Reads:  1000,
		Writes: 2000,
		Totals: 3000,
	}
	new := BlkioData{
		Time:   newTimestamp,
		Reads:  1010,
		Writes: 2020,
		Totals: 3030,
	}

	var calculator = BlkioCalculatorImpl{old, new}

	// WHEN
	value := calculator.GetReadPs()

	// THEN
	// value should be (1010-1000)/2
	assert.Equal(t, float64(5), value)
}

func TestBlkioWrite(t *testing.T) {
	// GIVEN
	oldTimestamp := time.Now()
	newTimestamp := oldTimestamp.Add(2 * time.Second)

	old := BlkioData{
		Time:   oldTimestamp,
		Reads:  1000,
		Writes: 2000,
		Totals: 3000,
	}
	new := BlkioData{
		Time:   newTimestamp,
		Reads:  1010,
		Writes: 2020,
		Totals: 3030,
	}

	var calculator = BlkioCalculatorImpl{old, new}

	// WHEN
	value := calculator.GetWritePs()

	// THEN
	// value should be (2020-2000)/2
	assert.Equal(t, float64(10), value)
}

func TestBlkioTotal(t *testing.T) {
	// GIVEN
	oldTimestamp := time.Now()
	newTimestamp := oldTimestamp.Add(2 * time.Second)

	old := BlkioData{
		Time:   oldTimestamp,
		Reads:  1000,
		Writes: 2000,
		Totals: 3000,
	}
	new := BlkioData{
		Time:   newTimestamp,
		Reads:  1010,
		Writes: 2020,
		Totals: 3030,
	}

	var calculator = BlkioCalculatorImpl{old, new}

	// WHEN
	value := calculator.GetTotalPs()

	// THEN
	// value should be (3030-3000)/2
	assert.Equal(t, float64(15), value)
}
