// Code generated by mockery v2.22.1. DO NOT EDIT.

package mocks

import (
	allocationchange "github.com/0chain/gosdk/zboxcore/allocationchange"
	fileref "github.com/0chain/gosdk/zboxcore/fileref"

	mock "github.com/stretchr/testify/mock"
)

// AllocationChange is an autogenerated mock type for the AllocationChange type
type AllocationChange struct {
	mock.Mock
}

// GetAffectedPath provides a mock function with given fields:
func (_m *AllocationChange) GetAffectedPath() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// GetSize provides a mock function with given fields:
func (_m *AllocationChange) GetSize() int64 {
	ret := _m.Called()

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// ProcessChange provides a mock function with given fields: rootRef
func (_m *AllocationChange) ProcessChange(rootRef *fileref.Ref) (allocationchange.CommitParams, error) {
	ret := _m.Called(rootRef)

	var r0 allocationchange.CommitParams
	var r1 error
	if rf, ok := ret.Get(0).(func(*fileref.Ref) (allocationchange.CommitParams, error)); ok {
		return rf(rootRef)
	}
	if rf, ok := ret.Get(0).(func(*fileref.Ref) allocationchange.CommitParams); ok {
		r0 = rf(rootRef)
	} else {
		r0 = ret.Get(0).(allocationchange.CommitParams)
	}

	if rf, ok := ret.Get(1).(func(*fileref.Ref) error); ok {
		r1 = rf(rootRef)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewAllocationChange interface {
	mock.TestingT
	Cleanup(func())
}

// NewAllocationChange creates a new instance of AllocationChange. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAllocationChange(t mockConstructorTestingTNewAllocationChange) *AllocationChange {
	mock := &AllocationChange{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
