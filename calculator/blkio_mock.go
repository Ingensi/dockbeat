package calculator

import "github.com/stretchr/testify/mock"

type MockedBlkioCalculator struct {
	mock.Mock
}

func (_m *MockedBlkioCalculator) GetReadPs() float64 {
	ret := _m.Called()

	var r0 float64
	if rf, ok := ret.Get(0).(func() float64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}
func (_m *MockedBlkioCalculator) GetWritePs() float64 {
	ret := _m.Called()

	var r0 float64
	if rf, ok := ret.Get(0).(func() float64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}
func (_m *MockedBlkioCalculator) GetTotalPs() float64 {
	ret := _m.Called()

	var r0 float64
	if rf, ok := ret.Get(0).(func() float64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}
