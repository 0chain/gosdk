// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// M3u8Writer is an autogenerated mock type for the M3u8Writer type
type M3u8Writer struct {
	mock.Mock
}

// Seek provides a mock function with given fields: offset, whence
func (_m *M3u8Writer) Seek(offset int64, whence int) (int64, error) {
	ret := _m.Called(offset, whence)

	var r0 int64
	if rf, ok := ret.Get(0).(func(int64, int) int64); ok {
		r0 = rf(offset, whence)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64, int) error); ok {
		r1 = rf(offset, whence)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Sync provides a mock function with given fields:
func (_m *M3u8Writer) Sync() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Truncate provides a mock function with given fields: size
func (_m *M3u8Writer) Truncate(size int64) error {
	ret := _m.Called(size)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64) error); ok {
		r0 = rf(size)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Write provides a mock function with given fields: p
func (_m *M3u8Writer) Write(p []byte) (int, error) {
	ret := _m.Called(p)

	var r0 int
	if rf, ok := ret.Get(0).(func([]byte) int); ok {
		r0 = rf(p)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(p)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewM3u8Writer interface {
	mock.TestingT
	Cleanup(func())
}

// NewM3u8Writer creates a new instance of M3u8Writer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewM3u8Writer(t mockConstructorTestingTNewM3u8Writer) *M3u8Writer {
	mock := &M3u8Writer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
