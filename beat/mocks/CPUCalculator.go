package mocks

import "github.com/stretchr/testify/mock"

import "github.com/elastic/beats/libbeat/common"

type CPUCalculator struct {
	mock.Mock
}

func (_m *CPUCalculator) perCpuUsage() common.MapStr {
	ret := _m.Called()

	var r0 common.MapStr
	if rf, ok := ret.Get(0).(func() common.MapStr); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(common.MapStr)
	}

	return r0
}
func (_m *CPUCalculator) totalUsage() float64 {
	ret := _m.Called()

	var r0 float64
	if rf, ok := ret.Get(0).(func() float64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}
func (_m *CPUCalculator) usageInKernelmode() float64 {
	ret := _m.Called()

	var r0 float64
	if rf, ok := ret.Get(0).(func() float64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}
func (_m *CPUCalculator) usageInUsermode() float64 {
	ret := _m.Called()

	var r0 float64
	if rf, ok := ret.Get(0).(func() float64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}
func (_m *CPUCalculator) calculateLoad(value uint64) float64 {
	ret := _m.Called(value)

	var r0 float64
	if rf, ok := ret.Get(0).(func(uint64) float64); ok {
		r0 = rf(value)
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}
