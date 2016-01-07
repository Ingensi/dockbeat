package beat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBlkioCalculator(t *testing.T) {
	// GIVEN
	// a factory
	factory := CalculatorFactoryImpl{}
	new := BlkioData{}
	old := BlkioData{}

	// WHEN
	calculator := factory.newBlkioCalculator(old, new)

	// THEN
	// calculator is not null and data stored are correct
	assert.Equal(t, new, calculator.(BlkioCalculatorImpl).new)
	assert.Equal(t, old, calculator.(BlkioCalculatorImpl).old)
}

func TestNewCPUCalculator(t *testing.T) {
	// GIVEN
	// a factory
	factory := CalculatorFactoryImpl{}
	new := CPUData{}
	old := CPUData{}

	// WHEN
	calculator := factory.newCPUCalculator(old, new)

	// THEN
	// calculator is not null and data stored are correct
	assert.Equal(t, new, calculator.(CPUCalculatorImpl).new)
	assert.Equal(t, old, calculator.(CPUCalculatorImpl).old)
}

func TestNewNetworkCalculator(t *testing.T) {
	// GIVEN
	// a factory
	factory := CalculatorFactoryImpl{}
	new := NetworkData{}
	old := NetworkData{}

	// WHEN
	calculator := factory.newNetworkCalculator(old, new)

	// THEN
	// calculator is not null and data stored are correct
	assert.Equal(t, new, calculator.(NetworkCalculatorImpl).new)
	assert.Equal(t, old, calculator.(NetworkCalculatorImpl).old)
}
