// Code generated by mockery v2.22.1. DO NOT EDIT.

package mocks

import (
	zcncore "github.com/0chain/gosdk/zcncore"
	mock "github.com/stretchr/testify/mock"
)

// GetNotProcessedZCNBurnTicketsCallback is an autogenerated mock type for the GetNotProcessedZCNBurnTicketsCallback type
type GetNotProcessedZCNBurnTicketsCallback struct {
	mock.Mock
}

// OnAddBurnTicket provides a mock function with given fields: value
func (_m *GetNotProcessedZCNBurnTicketsCallback) OnAddBurnTicket(value *zcncore.BurnTicket) {
	_m.Called(value)
}

// OnResponseAvailable provides a mock function with given fields: status, info
func (_m *GetNotProcessedZCNBurnTicketsCallback) OnResponseAvailable(status int, info string) {
	_m.Called(status, info)
}

type mockConstructorTestingTNewGetNotProcessedZCNBurnTicketsCallback interface {
	mock.TestingT
	Cleanup(func())
}

// NewGetNotProcessedZCNBurnTicketsCallback creates a new instance of GetNotProcessedZCNBurnTicketsCallback. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewGetNotProcessedZCNBurnTicketsCallback(t mockConstructorTestingTNewGetNotProcessedZCNBurnTicketsCallback) *GetNotProcessedZCNBurnTicketsCallback {
	mock := &GetNotProcessedZCNBurnTicketsCallback{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
