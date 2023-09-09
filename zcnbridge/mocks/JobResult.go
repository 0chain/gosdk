// Code generated by mockery v2.22.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// JobResult is an autogenerated mock type for the JobResult type
type JobResult struct {
	mock.Mock
}

// Data provides a mock function with given fields:
func (_m *JobResult) Data() interface{} {
	ret := _m.Called()

	var r0 interface{}
	if rf, ok := ret.Get(0).(func() interface{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

// Error provides a mock function with given fields:
func (_m *JobResult) Error() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAuthorizerID provides a mock function with given fields:
func (_m *JobResult) GetAuthorizerID() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// SetAuthorizerID provides a mock function with given fields: ID
func (_m *JobResult) SetAuthorizerID(ID string) {
	_m.Called(ID)
}

type mockConstructorTestingTNewJobResult interface {
	mock.TestingT
	Cleanup(func())
}

// NewJobResult creates a new instance of JobResult. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewJobResult(t mockConstructorTestingTNewJobResult) *JobResult {
	mock := &JobResult{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
