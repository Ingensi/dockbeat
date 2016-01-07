package mocks

import "github.com/ingensi/dockerbeat/beat"
import "github.com/stretchr/testify/mock"

type CalculatorFactory struct {
	mock.Mock
}

func (_m *CalculatorFactory) newBlkioCalculator(old beat.BlkioData, new beat.BlkioData) beat.BlkioCalculator {
	ret := _m.Called(old, new)

	var r0 beat.BlkioCalculator
	if rf, ok := ret.Get(0).(func(beat.BlkioData, beat.BlkioData) beat.BlkioCalculator); ok {
		r0 = rf(old, new)
	} else {
		r0 = ret.Get(0).(beat.BlkioCalculator)
	}

	return r0
}
func (_m *CalculatorFactory) newCPUCalculator(old beat.CPUData, new beat.CPUData) beat.CPUCalculator {
	ret := _m.Called(old, new)

	var r0 beat.CPUCalculator
	if rf, ok := ret.Get(0).(func(beat.CPUData, beat.CPUData) beat.CPUCalculator); ok {
		r0 = rf(old, new)
	} else {
		r0 = ret.Get(0).(beat.CPUCalculator)
	}

	return r0
}
func (_m *CalculatorFactory) newNetworkCalculator(old beat.NetworkData, new beat.NetworkData) beat.NetworkCalculator {
	ret := _m.Called(old, new)

	var r0 beat.NetworkCalculator
	if rf, ok := ret.Get(0).(func(beat.NetworkData, beat.NetworkData) beat.NetworkCalculator); ok {
		r0 = rf(old, new)
	} else {
		r0 = ret.Get(0).(beat.NetworkCalculator)
	}

	return r0
}
