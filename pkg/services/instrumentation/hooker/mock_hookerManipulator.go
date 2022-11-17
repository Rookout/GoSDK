

package hooker

import (
	types "github.com/Rookout/GoSDK/pkg/types"
	mock "github.com/stretchr/testify/mock"
)


type mockHookerManipulator struct {
	mock.Mock
}


func (_m *mockHookerManipulator) addBreakpoint(bpAddress uint64, functionEntry uint64, functionEnd uint64) {
	_m.Called(bpAddress, functionEntry, functionEnd)
}


func (_m *mockHookerManipulator) getActiveBreakpointsWithNew(functionEntry uint64, newBreakpoint uint64) []uint64 {
	ret := _m.Called(functionEntry, newBreakpoint)

	var r0 []uint64
	if rf, ok := ret.Get(0).(func(uint64, uint64) []uint64); ok {
		r0 = rf(functionEntry, newBreakpoint)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]uint64)
		}
	}

	return r0
}


func (_m *mockHookerManipulator) getActiveBreakpointsWithoutOld(functionEntry uint64, oldBreakpoint uint64) []uint64 {
	ret := _m.Called(functionEntry, oldBreakpoint)

	var r0 []uint64
	if rf, ok := ret.Get(0).(func(uint64, uint64) []uint64); ok {
		r0 = rf(functionEntry, oldBreakpoint)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]uint64)
		}
	}

	return r0
}


func (_m *mockHookerManipulator) getNativeAPI() types.NativeHookerAPI {
	ret := _m.Called()

	var r0 types.NativeHookerAPI
	if rf, ok := ret.Get(0).(func() types.NativeHookerAPI); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.NativeHookerAPI)
		}
	}

	return r0
}


func (_m *mockHookerManipulator) removeBreakpoint(bpAddress uint64, functionEntry uint64) {
	_m.Called(bpAddress, functionEntry)
}

type mockConstructorTestingTnewMockHookerManipulator interface {
	mock.TestingT
	Cleanup(func())
}


func newMockHookerManipulator(t mockConstructorTestingTnewMockHookerManipulator) *mockHookerManipulator {
	mock := &mockHookerManipulator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
