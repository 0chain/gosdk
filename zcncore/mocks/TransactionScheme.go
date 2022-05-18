// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	zcncore "github.com/0chain/gosdk/zcncore"
	mock "github.com/stretchr/testify/mock"
)

// TransactionScheme is an autogenerated mock type for the TransactionScheme type
type TransactionScheme struct {
	mock.Mock
}

// CancelAllocation provides a mock function with given fields: allocID, fee
func (_m *TransactionScheme) CancelAllocation(allocID string, fee int64) error {
	ret := _m.Called(allocID, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64) error); ok {
		r0 = rf(allocID, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateAllocation provides a mock function with given fields: car, lock, fee
func (_m *TransactionScheme) CreateAllocation(car *zcncore.CreateAllocationRequest, lock int64, fee int64) error {
	ret := _m.Called(car, lock, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.CreateAllocationRequest, int64, int64) error); ok {
		r0 = rf(car, lock, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateReadPool provides a mock function with given fields: fee
func (_m *TransactionScheme) CreateReadPool(fee int64) error {
	ret := _m.Called(fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64) error); ok {
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

// ExecuteSmartContract provides a mock function with given fields: address, methodName, jsoninput, val
func (_m *TransactionScheme) ExecuteSmartContract(address string, methodName string, jsoninput string, val int64) error {
	ret := _m.Called(address, methodName, jsoninput, val)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, int64) error); ok {
		r0 = rf(address, methodName, jsoninput, val)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FinalizeAllocation provides a mock function with given fields: allocID, fee
func (_m *TransactionScheme) FinalizeAllocation(allocID string, fee int64) error {
	ret := _m.Called(allocID, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64) error); ok {
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

// LockTokens provides a mock function with given fields: val, durationHr, durationMin
func (_m *TransactionScheme) LockTokens(val int64, durationHr int64, durationMin int) error {
	ret := _m.Called(val, durationHr, durationMin)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64, int64, int) error); ok {
		r0 = rf(val, durationHr, durationMin)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MienrSCUnlock provides a mock function with given fields: minerID, poolID
func (_m *TransactionScheme) MinerSCUnlock(minerID string, poolID string) error {
	ret := _m.Called(minerID, poolID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(minerID, poolID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MinerSCLock provides a mock function with given fields: minerID, lock
func (_m *TransactionScheme) MinerSCLock(minerID string, lock int64) error {
	ret := _m.Called(minerID, lock)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64) error); ok {
		r0 = rf(minerID, lock)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MinerSCSettings provides a mock function with given fields: _a0
func (_m *TransactionScheme) MinerSCSettings(_a0 *zcncore.MinerSCMinerInfo) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.MinerSCMinerInfo) error); ok {
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
func (_m *TransactionScheme) ReadPoolLock(allocID string, blobberID string, duration int64, lock int64, fee int64) error {
	ret := _m.Called(allocID, blobberID, duration, lock, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, int64, int64, int64) error); ok {
		r0 = rf(allocID, blobberID, duration, lock, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ReadPoolUnlock provides a mock function with given fields: poolID, fee
func (_m *TransactionScheme) ReadPoolUnlock(poolID string, fee int64) error {
	ret := _m.Called(poolID, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64) error); ok {
		r0 = rf(poolID, fee)
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
func (_m *TransactionScheme) Send(toClientID string, val int64, desc string) error {
	ret := _m.Called(toClientID, val, desc)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64, string) error); ok {
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
func (_m *TransactionScheme) SetTransactionFee(txnFee int64) error {
	ret := _m.Called(txnFee)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64) error); ok {
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

// StakePoolLock provides a mock function with given fields: blobberID, lock, fee
func (_m *TransactionScheme) StakePoolLock(blobberID string, lock int64, fee int64) error {
	ret := _m.Called(blobberID, lock, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64, int64) error); ok {
		r0 = rf(blobberID, lock, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StakePoolUnlock provides a mock function with given fields: blobberID, poolID, fee
func (_m *TransactionScheme) StakePoolUnlock(blobberID string, poolID string, fee int64) error {
	ret := _m.Called(blobberID, poolID, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, int64) error); ok {
		r0 = rf(blobberID, poolID, fee)
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

// UnlockTokens provides a mock function with given fields: poolID
func (_m *TransactionScheme) UnlockTokens(poolID string) error {
	ret := _m.Called(poolID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(poolID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateAllocation provides a mock function with given fields: allocID, sizeDiff, expirationDiff, lock, fee
func (_m *TransactionScheme) UpdateAllocation(allocID string, sizeDiff int64, expirationDiff int64, lock int64, fee int64) error {
	ret := _m.Called(allocID, sizeDiff, expirationDiff, lock, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64, int64, int64, int64) error); ok {
		r0 = rf(allocID, sizeDiff, expirationDiff, lock, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateBlobberSettings provides a mock function with given fields: blobber, fee
func (_m *TransactionScheme) UpdateBlobberSettings(blobber *zcncore.Blobber, fee int64) error {
	ret := _m.Called(blobber, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.Blobber, int64) error); ok {
		r0 = rf(blobber, fee)
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
func (_m *TransactionScheme) VestingAdd(ar *zcncore.VestingAddRequest, value int64) error {
	ret := _m.Called(ar, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.VestingAddRequest, int64) error); ok {
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

// VestingUpdateConfig provides a mock function with given fields: vscc
func (_m *TransactionScheme) VestingUpdateConfig(vscc *zcncore.InputMap) error {
	ret := _m.Called(vscc)

	var r0 error
	if rf, ok := ret.Get(0).(func(*zcncore.InputMap) error); ok {
		r0 = rf(vscc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WritePoolLock provides a mock function with given fields: allocID, blobberID, duration, lock, fee
func (_m *TransactionScheme) WritePoolLock(allocID string, blobberID string, duration int64, lock int64, fee int64) error {
	ret := _m.Called(allocID, blobberID, duration, lock, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, int64, int64, int64) error); ok {
		r0 = rf(allocID, blobberID, duration, lock, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WritePoolUnlock provides a mock function with given fields: poolID, fee
func (_m *TransactionScheme) WritePoolUnlock(poolID string, fee int64) error {
	ret := _m.Called(poolID, fee)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64) error); ok {
		r0 = rf(poolID, fee)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
