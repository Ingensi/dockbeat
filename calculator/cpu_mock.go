package calculator

import "github.com/stretchr/testify/mock"

import "github.com/elastic/beats/libbeat/common"

type MockedCPUCalculator struct {
	mock.Mock
}

func (_m *MockedCPUCalculator) perCpuUsage() common.MapStr {
	ret := _m.Called()

	var r0 common.MapStr
	if rf, ok := ret.Get(0).(func() common.MapStr); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(common.MapStr)
	}

	return r0
}
func (_m *MockedCPUCalculator) totalUsage() float64 {
	ret := _m.Called()

	var r0 float64
	if rf, ok := ret.Get(0).(func() float64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}
func (_m *MockedCPUCalculator) usageInKernelmode() float64 {
	ret := _m.Called()

	var r0 float64
	if rf, ok := ret.Get(0).(func() float64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}
func (_m *MockedCPUCalculator) usageInUsermode() float64 {
	ret := _m.Called()

	var r0 float64
	if rf, ok := ret.Get(0).(func() float64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}
