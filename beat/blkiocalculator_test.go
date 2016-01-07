package beat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlkioRead(t *testing.T) {
	// GIVEN
	old := BlkioData{1000, 2000, 3000}
	new := BlkioData{1010, 2020, 3030}

	var calculator = BlkioCalculator{old, new}

	// WHEN
	value := calculator.getRead()

	// THEN
	// value should be 1010-1000
	assert.Equal(t, uint64(10), value)
}

func TestBlkioWrite(t *testing.T) {
	// GIVEN
	old := BlkioData{1000, 2000, 3000}
	new := BlkioData{1010, 2020, 3030}

	var calculator = BlkioCalculator{old, new}

	// WHEN
	value := calculator.getWrite()

	// THEN
	// value should be 2020-2000
	assert.Equal(t, uint64(20), value)
}

func TestBlkioTotal(t *testing.T) {
	// GIVEN
	old := BlkioData{1000, 2000, 3000}
	new := BlkioData{1010, 2020, 3030}

	var calculator = BlkioCalculator{old, new}

	// WHEN
	value := calculator.getTotal()

	// THEN
	// value should be 3030-3000
	assert.Equal(t, uint64(30), value)
}
