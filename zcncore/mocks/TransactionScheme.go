// Code generated by mockery v2.22.1. DO NOT EDIT.

package mocks

import (
	transaction "github.com/0chain/gosdk/core/transaction"
	mock "github.com/stretchr/testify/mock"

	zcncore "github.com/0chain/gosdk/zcncore"
)

// TransactionScheme is an autogenerated mock type for the TransactionScheme type
type TransactionScheme struct {
	mock.Mock
}

// CancelAllocation provides a mock function with given fields: allocID, fee
func (_m *TransactionScheme) CancelAllocation(allocID string, fee uint64) error {
	ret := _m.Called(allocID, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, uint64) error); ok {
		r0 = rf(allocID, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateAllocation provides a mock function with given fields: car, lock, fee
func (_m *TransactionScheme) CreateAllocation(car *zcncore.CreateAllocationRequest, lock uint64, fee uint64) error {
	ret := _m.Called(car, lock, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.CreateAllocationRequest, uint64, uint64) error); ok {
		r0 = rf(car, lock, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateReadPool provides a mock function with given fields: fee
func (_m *TransactionScheme) CreateReadPool(fee uint64) error {
	ret := _m.Called(fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint64) error); ok {
		r0 = rf(fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ExecuteFaucetSCWallet provides a mock function with given fields: walletStr, methodName, input
func (_m *TransactionScheme) ExecuteFaucetSCWallet(walletStr string, methodName string, input []byte) error {
	ret := _m.Called(walletStr, methodName, input)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, []byte) error); ok {
		r0 = rf(walletStr, methodName, input)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ExecuteSmartContract provides a mock function with given fields: address, methodName, input, val
func (_m *TransactionScheme) ExecuteSmartContract(address string, methodName string, input interface{}, val uint64) (*transaction.Transaction, error) {
	ret := _m.Called(address, methodName, input, val)

	var r0 *transaction.Transaction
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, interface{}, uint64) (*transaction.Transaction, error)); ok {
		return rf(address, methodName, input, val)
	}
	if rf, ok := ret.Get(0).(func(string, string, interface{}, uint64) *transaction.Transaction); ok {
		r0 = rf(address, methodName, input, val)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*transaction.Transaction)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, interface{}, uint64) error); ok {
		r1 = rf(address, methodName, input, val)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FaucetUpdateConfig provides a mock function with given fields: _a0
func (_m *TransactionScheme) FaucetUpdateConfig(_a0 *zcncore.InputMap) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.InputMap) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FinalizeAllocation provides a mock function with given fields: allocID, fee
func (_m *TransactionScheme) FinalizeAllocation(allocID string, fee uint64) error {
	ret := _m.Called(allocID, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, uint64) error); ok {
		r0 = rf(allocID, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetTransactionError provides a mock function with given fields:
func (_m *TransactionScheme) GetTransactionError() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetTransactionHash provides a mock function with given fields:
func (_m *TransactionScheme) GetTransactionHash() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetTransactionNonce provides a mock function with given fields:
func (_m *TransactionScheme) GetTransactionNonce() int64 {
	ret := _m.Called()

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// GetVerifyConfirmationStatus provides a mock function with given fields:
func (_m *TransactionScheme) GetVerifyConfirmationStatus() zcncore.ConfirmationStatus {
	ret := _m.Called()

	var r0 zcncore.ConfirmationStatus
	if rf, ok := ret.Get(0).(func() zcncore.ConfirmationStatus); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(zcncore.ConfirmationStatus)
	}

	return r0
}

// GetVerifyError provides a mock function with given fields:
func (_m *TransactionScheme) GetVerifyError() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetVerifyOutput provides a mock function with given fields:
func (_m *TransactionScheme) GetVerifyOutput() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Hash provides a mock function with given fields:
func (_m *TransactionScheme) Hash() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MinerSCCollectReward provides a mock function with given fields: providerID, providerType
func (_m *TransactionScheme) MinerSCCollectReward(providerID string, providerType zcncore.Provider) error {
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
func (_m *TransactionScheme) MinerSCDeleteMiner(_a0 *zcncore.MinerSCMinerInfo) error {
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
func (_m *TransactionScheme) MinerSCDeleteSharder(_a0 *zcncore.MinerSCMinerInfo) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.MinerSCMinerInfo) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MinerSCLock provides a mock function with given fields: providerId, providerType, lock
func (_m *TransactionScheme) MinerSCLock(providerId string, providerType zcncore.Provider, lock uint64) error {
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
func (_m *TransactionScheme) MinerSCMinerSettings(_a0 *zcncore.MinerSCMinerInfo) error {
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
func (_m *TransactionScheme) MinerSCSharderSettings(_a0 *zcncore.MinerSCMinerInfo) error {
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
func (_m *TransactionScheme) MinerSCUnlock(providerId string, providerType zcncore.Provider) error {
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
func (_m *TransactionScheme) MinerScUpdateConfig(_a0 *zcncore.InputMap) error {
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
func (_m *TransactionScheme) MinerScUpdateGlobals(_a0 *zcncore.InputMap) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.InputMap) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Output provides a mock function with given fields:
func (_m *TransactionScheme) Output() []byte {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	return r0
}

// ReadPoolLock provides a mock function with given fields: allocID, blobberID, duration, lock, fee
func (_m *TransactionScheme) ReadPoolLock(allocID string, blobberID string, duration int64, lock uint64, fee uint64) error {
	ret := _m.Called(allocID, blobberID, duration, lock, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, int64, uint64, uint64) error); ok {
		r0 = rf(allocID, blobberID, duration, lock, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ReadPoolUnlock provides a mock function with given fields: fee
func (_m *TransactionScheme) ReadPoolUnlock(fee uint64) error {
	ret := _m.Called(fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint64) error); ok {
		r0 = rf(fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RegisterMultiSig provides a mock function with given fields: walletstr, mswallet
func (_m *TransactionScheme) RegisterMultiSig(walletstr string, mswallet string) error {
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
func (_m *TransactionScheme) Send(toClientID string, val uint64, desc string) error {
	ret := _m.Called(toClientID, val, desc)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, uint64, string) error); ok {
		r0 = rf(toClientID, val, desc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetTransactionCallback provides a mock function with given fields: cb
func (_m *TransactionScheme) SetTransactionCallback(cb zcncore.TransactionCallback) error {
	ret := _m.Called(cb)

	var r0 error
	if rf, ok := ret.Get(0).(func(zcncore.TransactionCallback) error); ok {
		r0 = rf(cb)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetTransactionFee provides a mock function with given fields: txnFee
func (_m *TransactionScheme) SetTransactionFee(txnFee uint64) error {
	ret := _m.Called(txnFee)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint64) error); ok {
		r0 = rf(txnFee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetTransactionHash provides a mock function with given fields: hash
func (_m *TransactionScheme) SetTransactionHash(hash string) error {
	ret := _m.Called(hash)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(hash)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetTransactionNonce provides a mock function with given fields: txnNonce
func (_m *TransactionScheme) SetTransactionNonce(txnNonce int64) error {
	ret := _m.Called(txnNonce)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64) error); ok {
		r0 = rf(txnNonce)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StakePoolLock provides a mock function with given fields: providerId, providerType, lock, fee
func (_m *TransactionScheme) StakePoolLock(providerId string, providerType zcncore.Provider, lock uint64, fee uint64) error {
	ret := _m.Called(providerId, providerType, lock, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, zcncore.Provider, uint64, uint64) error); ok {
		r0 = rf(providerId, providerType, lock, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StakePoolUnlock provides a mock function with given fields: providerId, providerType, fee
func (_m *TransactionScheme) StakePoolUnlock(providerId string, providerType zcncore.Provider, fee uint64) error {
	ret := _m.Called(providerId, providerType, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, zcncore.Provider, uint64) error); ok {
		r0 = rf(providerId, providerType, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StorageSCCollectReward provides a mock function with given fields: providerID, providerType
func (_m *TransactionScheme) StorageSCCollectReward(providerID string, providerType zcncore.Provider) error {
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
func (_m *TransactionScheme) StorageScUpdateConfig(_a0 *zcncore.InputMap) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.InputMap) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StoreData provides a mock function with given fields: data
func (_m *TransactionScheme) StoreData(data string) error {
	ret := _m.Called(data)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateAllocation provides a mock function with given fields: allocID, sizeDiff, expirationDiff, lock, fee
func (_m *TransactionScheme) UpdateAllocation(allocID string, sizeDiff int64, expirationDiff int64, lock uint64, fee uint64) error {
	ret := _m.Called(allocID, sizeDiff, expirationDiff, lock, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64, int64, uint64, uint64) error); ok {
		r0 = rf(allocID, sizeDiff, expirationDiff, lock, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateBlobberSettings provides a mock function with given fields: blobber, fee
func (_m *TransactionScheme) UpdateBlobberSettings(blobber *zcncore.Blobber, fee uint64) error {
	ret := _m.Called(blobber, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.Blobber, uint64) error); ok {
		r0 = rf(blobber, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateValidatorSettings provides a mock function with given fields: validator, fee
func (_m *TransactionScheme) UpdateValidatorSettings(validator *zcncore.Validator, fee uint64) error {
	ret := _m.Called(validator, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.Validator, uint64) error); ok {
		r0 = rf(validator, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Verify provides a mock function with given fields:
func (_m *TransactionScheme) Verify() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// VestingAdd provides a mock function with given fields: ar, value
func (_m *TransactionScheme) VestingAdd(ar *zcncore.VestingAddRequest, value uint64) error {
	ret := _m.Called(ar, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.VestingAddRequest, uint64) error); ok {
		r0 = rf(ar, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// VestingDelete provides a mock function with given fields: poolID
func (_m *TransactionScheme) VestingDelete(poolID string) error {
	ret := _m.Called(poolID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(poolID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// VestingStop provides a mock function with given fields: sr
func (_m *TransactionScheme) VestingStop(sr *zcncore.VestingStopRequest) error {
	ret := _m.Called(sr)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.VestingStopRequest) error); ok {
		r0 = rf(sr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// VestingTrigger provides a mock function with given fields: poolID
func (_m *TransactionScheme) VestingTrigger(poolID string) error {
	ret := _m.Called(poolID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(poolID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// VestingUnlock provides a mock function with given fields: poolID
func (_m *TransactionScheme) VestingUnlock(poolID string) error {
	ret := _m.Called(poolID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(poolID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// VestingUpdateConfig provides a mock function with given fields: _a0
func (_m *TransactionScheme) VestingUpdateConfig(_a0 *zcncore.InputMap) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.InputMap) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WritePoolLock provides a mock function with given fields: allocID, blobberID, duration, lock, fee
func (_m *TransactionScheme) WritePoolLock(allocID string, blobberID string, duration int64, lock uint64, fee uint64) error {
	ret := _m.Called(allocID, blobberID, duration, lock, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, int64, uint64, uint64) error); ok {
		r0 = rf(allocID, blobberID, duration, lock, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WritePoolUnlock provides a mock function with given fields: allocID, fee
func (_m *TransactionScheme) WritePoolUnlock(allocID string, fee uint64) error {
	ret := _m.Called(allocID, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, uint64) error); ok {
		r0 = rf(allocID, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ZCNSCAddAuthorizer provides a mock function with given fields: _a0
func (_m *TransactionScheme) ZCNSCAddAuthorizer(_a0 *zcncore.AddAuthorizerPayload) error {
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
func (_m *TransactionScheme) ZCNSCAuthorizerHealthCheck(_a0 *zcncore.AuthorizerHealthCheckPayload) error {
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
func (_m *TransactionScheme) ZCNSCDeleteAuthorizer(_a0 *zcncore.DeleteAuthorizerPayload) error {
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
func (_m *TransactionScheme) ZCNSCUpdateAuthorizerConfig(_a0 *zcncore.AuthorizerNode) error {
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
func (_m *TransactionScheme) ZCNSCUpdateGlobalConfig(_a0 *zcncore.InputMap) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.InputMap) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewTransactionScheme interface {
	mock.TestingT
	Cleanup(func())
}

// NewTransactionScheme creates a new instance of TransactionScheme. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTransactionScheme(t mockConstructorTestingTNewTransactionScheme) *TransactionScheme {
	mock := &TransactionScheme{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
