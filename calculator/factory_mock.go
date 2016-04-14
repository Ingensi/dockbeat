package calculator

import "github.com/stretchr/testify/mock"

type MockedCalculatorFactory struct {
	mock.Mock
}

func (_m *MockedCalculatorFactory) NewBlkioCalculator(old BlkioData, new BlkioData) BlkioCalculator {
	ret := _m.Called(old, new)

	var r0 BlkioCalculator
	if rf, ok := ret.Get(0).(func(BlkioData, BlkioData) BlkioCalculator); ok {
		r0 = rf(old, new)
	} else {
		r0 = ret.Get(0).(BlkioCalculator)
	}

	return r0
}
func (_m *MockedCalculatorFactory) NewCPUCalculator(old CPUData, new CPUData) CPUCalculator {
	ret := _m.Called(old, new)

	var r0 CPUCalculator
	if rf, ok := ret.Get(0).(func(CPUData, CPUData) CPUCalculator); ok {
		r0 = rf(old, new)
	} else {
		r0 = ret.Get(0).(CPUCalculator)
	}

	return r0
}
func (_m *MockedCalculatorFactory) NewNetworkCalculator(old NetworkData, new NetworkData) NetworkCalculator {
	ret := _m.Called(old, new)

	var r0 NetworkCalculator
	if rf, ok := ret.Get(0).(func(NetworkData, NetworkData) NetworkCalculator); ok {
		r0 = rf(old, new)
	} else {
		r0 = ret.Get(0).(NetworkCalculator)
	}

	return r0
}
