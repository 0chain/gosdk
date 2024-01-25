// Code generated by mockery v2.36.0. DO NOT EDIT.

package mocks

import (
	context "context"

	transaction "github.com/0chain/gosdk/zcnbridge/transaction"
	mock "github.com/stretchr/testify/mock"

	zcncore "github.com/0chain/gosdk/zcncore"
)

// Transaction is an autogenerated mock type for the Transaction type
type Transaction struct {
	mock.Mock
}

// ExecuteSmartContract provides a mock function with given fields: ctx, address, funcName, input, val
func (_m *Transaction) ExecuteSmartContract(ctx context.Context, address string, funcName string, input interface{}, val uint64) (string, error) {
	ret := _m.Called(ctx, address, funcName, input, val)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, interface{}, uint64) (string, error)); ok {
		return rf(ctx, address, funcName, input, val)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, interface{}, uint64) string); ok {
		r0 = rf(ctx, address, funcName, input, val)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, interface{}, uint64) error); ok {
		r1 = rf(ctx, address, funcName, input, val)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCallback provides a mock function with given fields:
func (_m *Transaction) GetCallback() transaction.TransactionCallbackAwaitable {
	ret := _m.Called()

	var r0 transaction.TransactionCallbackAwaitable
	if rf, ok := ret.Get(0).(func() transaction.TransactionCallbackAwaitable); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(transaction.TransactionCallbackAwaitable)
		}
	}

	return r0
}

// GetHash provides a mock function with given fields:
func (_m *Transaction) GetHash() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetScheme provides a mock function with given fields:
func (_m *Transaction) GetScheme() zcncore.TransactionScheme {
	ret := _m.Called()

	var r0 zcncore.TransactionScheme
	if rf, ok := ret.Get(0).(func() zcncore.TransactionScheme); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(zcncore.TransactionScheme)
		}
	}

	return r0
}

// GetTransactionOutput provides a mock function with given fields:
func (_m *Transaction) GetTransactionOutput() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// SetHash provides a mock function with given fields: _a0
func (_m *Transaction) SetHash(_a0 string) {
	_m.Called(_a0)
}

// Verify provides a mock function with given fields: ctx
func (_m *Transaction) Verify(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewTransaction creates a new instance of Transaction. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTransaction(t interface {
	mock.TestingT
	Cleanup(func())
}) *Transaction {
	mock := &Transaction{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
