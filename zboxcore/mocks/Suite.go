// Code generated by mockery v2.22.1. DO NOT EDIT.

package mocks

import (
	cipher "crypto/cipher"

	kyber "go.dedis.ch/kyber/v3"

	mock "github.com/stretchr/testify/mock"
)

// Suite is an autogenerated mock type for the Suite type
type Suite struct {
	mock.Mock
}

// Point provides a mock function with given fields:
func (_m *Suite) Point() kyber.Point {
	ret := _m.Called()

	var r0 kyber.Point
	if rf, ok := ret.Get(0).(func() kyber.Point); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(kyber.Point)
		}
	}

	return r0
}

// PointLen provides a mock function with given fields:
func (_m *Suite) PointLen() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// RandomStream provides a mock function with given fields:
func (_m *Suite) RandomStream() cipher.Stream {
	ret := _m.Called()

	var r0 cipher.Stream
	if rf, ok := ret.Get(0).(func() cipher.Stream); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cipher.Stream)
		}
	}

	return r0
}

// Scalar provides a mock function with given fields:
func (_m *Suite) Scalar() kyber.Scalar {
	ret := _m.Called()

	var r0 kyber.Scalar
	if rf, ok := ret.Get(0).(func() kyber.Scalar); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(kyber.Scalar)
		}
	}

	return r0
}

// ScalarLen provides a mock function with given fields:
func (_m *Suite) ScalarLen() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// String provides a mock function with given fields:
func (_m *Suite) String() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

type mockConstructorTestingTNewSuite interface {
	mock.TestingT
	Cleanup(func())
}

// NewSuite creates a new instance of Suite. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewSuite(t mockConstructorTestingTNewSuite) *Suite {
	mock := &Suite{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
