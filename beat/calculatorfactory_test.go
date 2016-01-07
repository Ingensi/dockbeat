package beat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBlkioCalculator(t *testing.T) {
	// GIVEN
	// a factory
	factory := CalculatorFactory{}
	new := BlkioData{}
	old := BlkioData{}

	// WHEN
	calculator := factory.newBlkioCalculator(old, new)

	// THEN
	// calculator is not null and data stored are correct
	assert.Equal(t, new, calculator.new)
	assert.Equal(t, old, calculator.old)
}

func TestNewCPUCalculator(t *testing.T) {
	// GIVEN
	// a factory
	factory := CalculatorFactory{}
	new := CPUData{}
	old := CPUData{}

	// WHEN
	calculator := factory.newCPUCalculator(old, new)

	// THEN
	// calculator is not null and data stored are correct
	assert.Equal(t, new, calculator.new)
	assert.Equal(t, old, calculator.old)
}

func TestNewNetworkCalculator(t *testing.T) {
	// GIVEN
	// a factory
	factory := CalculatorFactory{}
	new := NetworkData{}
	old := NetworkData{}

	// WHEN
	calculator := factory.newNetworkCalculator(old, new)

	// THEN
	// calculator is not null and data stored are correct
	assert.Equal(t, new, calculator.new)
	assert.Equal(t, old, calculator.old)
}
