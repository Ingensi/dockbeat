package beat

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
		time:   oldTimestamp,
		reads:  1000,
		writes: 2000,
		totals: 3000,
	}
	new := BlkioData{
		time:   newTimestamp,
		reads:  1010,
		writes: 2020,
		totals: 3030,
	}

	var calculator = BlkioCalculatorImpl{old, new}

	// WHEN
	value := calculator.getReadPs()

	// THEN
	// value should be (1010-1000)/2
	assert.Equal(t, float64(5), value)
}

func TestBlkioWrite(t *testing.T) {
	// GIVEN
	oldTimestamp := time.Now()
	newTimestamp := oldTimestamp.Add(2 * time.Second)

	old := BlkioData{
		time:   oldTimestamp,
		reads:  1000,
		writes: 2000,
		totals: 3000,
	}
	new := BlkioData{
		time:   newTimestamp,
		reads:  1010,
		writes: 2020,
		totals: 3030,
	}

	var calculator = BlkioCalculatorImpl{old, new}

	// WHEN
	value := calculator.getWritePs()

	// THEN
	// value should be (2020-2000)/2
	assert.Equal(t, float64(10), value)
}

func TestBlkioTotal(t *testing.T) {
	// GIVEN
	oldTimestamp := time.Now()
	newTimestamp := oldTimestamp.Add(2 * time.Second)

	old := BlkioData{
		time:   oldTimestamp,
		reads:  1000,
		writes: 2000,
		totals: 3000,
	}
	new := BlkioData{
		time:   newTimestamp,
		reads:  1010,
		writes: 2020,
		totals: 3030,
	}

	var calculator = BlkioCalculatorImpl{old, new}

	// WHEN
	value := calculator.getTotalPs()

	// THEN
	// value should be (3030-3000)/2
	assert.Equal(t, float64(15), value)
}
