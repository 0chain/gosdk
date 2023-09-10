// Code generated by mockery v2.22.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// WalletCallback is an autogenerated mock type for the WalletCallback type
type WalletCallback struct {
	mock.Mock
}

// OnWalletCreateComplete provides a mock function with given fields: status, wallet, err
func (_m *WalletCallback) OnWalletCreateComplete(status int, wallet string, err string) {
	_m.Called(status, wallet, err)
}

type mockConstructorTestingTNewWalletCallback interface {
	mock.TestingT
	Cleanup(func())
}

// NewWalletCallback creates a new instance of WalletCallback. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewWalletCallback(t mockConstructorTestingTNewWalletCallback) *WalletCallback {
	mock := &WalletCallback{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
