// Code generated by mockery v2.22.1. DO NOT EDIT.

package mocks

import (
	transaction "github.com/0chain/gosdk/core/transaction"
	mock "github.com/stretchr/testify/mock"

	zcncore "github.com/0chain/gosdk/zcncore"
)

// TransactionCommon is an autogenerated mock type for the TransactionCommon type
type TransactionCommon struct {
	mock.Mock
}

// CancelAllocation provides a mock function with given fields: allocID
func (_m *TransactionCommon) CancelAllocation(allocID string) error {
	ret := _m.Called(allocID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(allocID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateAllocation provides a mock function with given fields: car, lock
func (_m *TransactionCommon) CreateAllocation(car *zcncore.CreateAllocationRequest, lock uint64) error {
	ret := _m.Called(car, lock)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.CreateAllocationRequest, uint64) error); ok {
		r0 = rf(car, lock)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateReadPool provides a mock function with given fields:
func (_m *TransactionCommon) CreateReadPool() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ExecuteSmartContract provides a mock function with given fields: address, methodName, input, val, feeOpts
func (_m *TransactionCommon) ExecuteSmartContract(address string, methodName string, input interface{}, val uint64, feeOpts ...zcncore.FeeOption) (*transaction.Transaction, error) {
	_va := make([]interface{}, len(feeOpts))
	for _i := range feeOpts {
		_va[_i] = feeOpts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, address, methodName, input, val)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *transaction.Transaction
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, interface{}, uint64, ...zcncore.FeeOption) (*transaction.Transaction, error)); ok {
		return rf(address, methodName, input, val, feeOpts...)
	}
	if rf, ok := ret.Get(0).(func(string, string, interface{}, uint64, ...zcncore.FeeOption) *transaction.Transaction); ok {
		r0 = rf(address, methodName, input, val, feeOpts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*transaction.Transaction)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, interface{}, uint64, ...zcncore.FeeOption) error); ok {
		r1 = rf(address, methodName, input, val, feeOpts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FaucetUpdateConfig provides a mock function with given fields: _a0
func (_m *TransactionCommon) FaucetUpdateConfig(_a0 *zcncore.InputMap) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.InputMap) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FinalizeAllocation provides a mock function with given fields: allocID
func (_m *TransactionCommon) FinalizeAllocation(allocID string) error {
	ret := _m.Called(allocID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(allocID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetVerifyConfirmationStatus provides a mock function with given fields:
func (_m *TransactionCommon) GetVerifyConfirmationStatus() zcncore.ConfirmationStatus {
	ret := _m.Called()

	var r0 zcncore.ConfirmationStatus
	if rf, ok := ret.Get(0).(func() zcncore.ConfirmationStatus); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(zcncore.ConfirmationStatus)
	}

	return r0
}

// MinerSCCollectReward provides a mock function with given fields: providerID, providerType
func (_m *TransactionCommon) MinerSCCollectReward(providerID string, providerType zcncore.Provider) error {
	ret := _m.Called(providerID, providerType)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, zcncore.Provider) error); ok {
		r0 = rf(providerID, providerType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MinerSCDeleteMiner provides a mock function with given fields: _a0
func (_m *TransactionCommon) MinerSCDeleteMiner(_a0 *zcncore.MinerSCMinerInfo) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.MinerSCMinerInfo) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MinerSCDeleteSharder provides a mock function with given fields: _a0
func (_m *TransactionCommon) MinerSCDeleteSharder(_a0 *zcncore.MinerSCMinerInfo) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.MinerSCMinerInfo) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MinerSCKill provides a mock function with given fields: providerID, providerType
func (_m *TransactionCommon) MinerSCKill(providerID string, providerType zcncore.Provider) error {
	ret := _m.Called(providerID, providerType)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, zcncore.Provider) error); ok {
		r0 = rf(providerID, providerType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MinerSCLock provides a mock function with given fields: providerId, providerType, lock
func (_m *TransactionCommon) MinerSCLock(providerId string, providerType zcncore.Provider, lock uint64) error {
	ret := _m.Called(providerId, providerType, lock)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, zcncore.Provider, uint64) error); ok {
		r0 = rf(providerId, providerType, lock)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MinerSCMinerSettings provides a mock function with given fields: _a0
func (_m *TransactionCommon) MinerSCMinerSettings(_a0 *zcncore.MinerSCMinerInfo) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.MinerSCMinerInfo) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MinerSCSharderSettings provides a mock function with given fields: _a0
func (_m *TransactionCommon) MinerSCSharderSettings(_a0 *zcncore.MinerSCMinerInfo) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.MinerSCMinerInfo) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MinerSCUnlock provides a mock function with given fields: providerId, providerType
func (_m *TransactionCommon) MinerSCUnlock(providerId string, providerType zcncore.Provider) error {
	ret := _m.Called(providerId, providerType)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, zcncore.Provider) error); ok {
		r0 = rf(providerId, providerType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MinerScUpdateConfig provides a mock function with given fields: _a0
func (_m *TransactionCommon) MinerScUpdateConfig(_a0 *zcncore.InputMap) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.InputMap) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MinerScUpdateGlobals provides a mock function with given fields: _a0
func (_m *TransactionCommon) MinerScUpdateGlobals(_a0 *zcncore.InputMap) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.InputMap) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ReadPoolLock provides a mock function with given fields: allocID, blobberID, duration, lock
func (_m *TransactionCommon) ReadPoolLock(allocID string, blobberID string, duration int64, lock uint64) error {
	ret := _m.Called(allocID, blobberID, duration, lock)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, int64, uint64) error); ok {
		r0 = rf(allocID, blobberID, duration, lock)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ReadPoolUnlock provides a mock function with given fields:
func (_m *TransactionCommon) ReadPoolUnlock() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RegisterMultiSig provides a mock function with given fields: walletstr, mswallet
func (_m *TransactionCommon) RegisterMultiSig(walletstr string, mswallet string) error {
	ret := _m.Called(walletstr, mswallet)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(walletstr, mswallet)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Send provides a mock function with given fields: toClientID, val, desc
func (_m *TransactionCommon) Send(toClientID string, val uint64, desc string) error {
	ret := _m.Called(toClientID, val, desc)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, uint64, string) error); ok {
		r0 = rf(toClientID, val, desc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StakePoolLock provides a mock function with given fields: providerId, providerType, lock
func (_m *TransactionCommon) StakePoolLock(providerId string, providerType zcncore.Provider, lock uint64) error {
	ret := _m.Called(providerId, providerType, lock)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, zcncore.Provider, uint64) error); ok {
		r0 = rf(providerId, providerType, lock)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StakePoolUnlock provides a mock function with given fields: providerId, providerType
func (_m *TransactionCommon) StakePoolUnlock(providerId string, providerType zcncore.Provider) error {
	ret := _m.Called(providerId, providerType)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, zcncore.Provider) error); ok {
		r0 = rf(providerId, providerType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StorageSCCollectReward provides a mock function with given fields: providerID, providerType
func (_m *TransactionCommon) StorageSCCollectReward(providerID string, providerType zcncore.Provider) error {
	ret := _m.Called(providerID, providerType)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, zcncore.Provider) error); ok {
		r0 = rf(providerID, providerType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StorageScUpdateConfig provides a mock function with given fields: _a0
func (_m *TransactionCommon) StorageScUpdateConfig(_a0 *zcncore.InputMap) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.InputMap) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateAllocation provides a mock function with given fields: allocID, sizeDiff, expirationDiff, lock
func (_m *TransactionCommon) UpdateAllocation(allocID string, sizeDiff int64, expirationDiff int64, lock uint64) error {
	ret := _m.Called(allocID, sizeDiff, expirationDiff, lock)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64, int64, uint64) error); ok {
		r0 = rf(allocID, sizeDiff, expirationDiff, lock)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateBlobberSettings provides a mock function with given fields: blobber
func (_m *TransactionCommon) UpdateBlobberSettings(blobber *zcncore.Blobber) error {
	ret := _m.Called(blobber)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.Blobber) error); ok {
		r0 = rf(blobber)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateValidatorSettings provides a mock function with given fields: validator
func (_m *TransactionCommon) UpdateValidatorSettings(validator *zcncore.Validator) error {
	ret := _m.Called(validator)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.Validator) error); ok {
		r0 = rf(validator)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// VestingAdd provides a mock function with given fields: ar, value
func (_m *TransactionCommon) VestingAdd(ar *zcncore.VestingAddRequest, value uint64) error {
	ret := _m.Called(ar, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.VestingAddRequest, uint64) error); ok {
		r0 = rf(ar, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// VestingUpdateConfig provides a mock function with given fields: _a0
func (_m *TransactionCommon) VestingUpdateConfig(_a0 *zcncore.InputMap) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.InputMap) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WritePoolLock provides a mock function with given fields: allocID, blobberID, duration, lock
func (_m *TransactionCommon) WritePoolLock(allocID string, blobberID string, duration int64, lock uint64) error {
	ret := _m.Called(allocID, blobberID, duration, lock)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, int64, uint64) error); ok {
		r0 = rf(allocID, blobberID, duration, lock)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WritePoolUnlock provides a mock function with given fields: allocID
func (_m *TransactionCommon) WritePoolUnlock(allocID string) error {
	ret := _m.Called(allocID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(allocID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ZCNSCAddAuthorizer provides a mock function with given fields: _a0
func (_m *TransactionCommon) ZCNSCAddAuthorizer(_a0 *zcncore.AddAuthorizerPayload) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.AddAuthorizerPayload) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ZCNSCAuthorizerHealthCheck provides a mock function with given fields: _a0
func (_m *TransactionCommon) ZCNSCAuthorizerHealthCheck(_a0 *zcncore.AuthorizerHealthCheckPayload) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.AuthorizerHealthCheckPayload) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ZCNSCDeleteAuthorizer provides a mock function with given fields: _a0
func (_m *TransactionCommon) ZCNSCDeleteAuthorizer(_a0 *zcncore.DeleteAuthorizerPayload) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.DeleteAuthorizerPayload) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ZCNSCUpdateAuthorizerConfig provides a mock function with given fields: _a0
func (_m *TransactionCommon) ZCNSCUpdateAuthorizerConfig(_a0 *zcncore.AuthorizerNode) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.AuthorizerNode) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ZCNSCUpdateGlobalConfig provides a mock function with given fields: _a0
func (_m *TransactionCommon) ZCNSCUpdateGlobalConfig(_a0 *zcncore.InputMap) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.InputMap) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewTransactionCommon interface {
	mock.TestingT
	Cleanup(func())
}

// NewTransactionCommon creates a new instance of TransactionCommon. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTransactionCommon(t mockConstructorTestingTNewTransactionCommon) *TransactionCommon {
	mock := &TransactionCommon{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// ZCNSCCollectReward provides a mock function with given fields: providerID, providerType
func (_m *TransactionCommon) ZCNSCCollectReward(providerID string, providerType zcncore.Provider) error {
	ret := _m.Called(providerID, providerType)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, zcncore.Provider) error); ok {
		r0 = rf(providerID, providerType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}