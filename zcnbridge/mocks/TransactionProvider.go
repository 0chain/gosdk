// Code generated by mockery v2.28.1. DO NOT EDIT.

package mocks

import (
	transaction "github.com/0chain/gosdk/zcnbridge/transaction"
	mock "github.com/stretchr/testify/mock"
)

// TransactionProvider is an autogenerated mock type for the TransactionProvider type
type TransactionProvider struct {
	mock.Mock
}

// NewTransactionEntity provides a mock function with given fields: txnFee
func (_m *TransactionProvider) NewTransactionEntity(txnFee uint64) (transaction.Transaction, error) {
	ret := _m.Called(txnFee)

	var r0 transaction.Transaction
	var r1 error
	if rf, ok := ret.Get(0).(func(uint64) (transaction.Transaction, error)); ok {
		return rf(txnFee)
	}
	if rf, ok := ret.Get(0).(func(uint64) transaction.Transaction); ok {
		r0 = rf(txnFee)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(transaction.Transaction)
		}
	}

	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(txnFee)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewTransactionProvider interface {
	mock.TestingT
	Cleanup(func())
}

// NewTransactionProvider creates a new instance of TransactionProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTransactionProvider(t mockConstructorTestingTNewTransactionProvider) *TransactionProvider {
	mock := &TransactionProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
