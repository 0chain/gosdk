//go:build mobile
// +build mobile

package zcncore

import (
	"encoding/json"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/transaction"
)

func (ta *TransactionWithAuth) ExecuteSmartContract(address, methodName string, input string, val string) error {
	v, err := parseCoinStr(val)
	if err != nil {
		return err
	}

	err = ta.t.createSmartContractTxn(address, methodName, json.RawMessage(input), v)
	if err != nil {
		return err
	}
	go func() {
		ta.submitTxn()
	}()
	return nil
}

func (ta *TransactionWithAuth) SetTransactionFee(txnFee string) error {
	return ta.t.SetTransactionFee(txnFee)
}

func (ta *TransactionWithAuth) Send(toClientID string, val string, desc string) error {
	v, err := parseCoinStr(val)
	if err != nil {
		return err
	}

	txnData, err := json.Marshal(SendTxnData{Note: desc})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}
	go func() {
		ta.t.txn.TransactionType = transaction.TxnTypeSend
		ta.t.txn.ToClientID = toClientID
		ta.t.txn.Value = v
		ta.t.txn.TransactionData = string(txnData)
		ta.submitTxn()
	}()
	return nil
}

func (ta *TransactionWithAuth) VestingAdd(ar *VestingAddRequest, value string) error {
	v, err := parseCoinStr(value)
	if err != nil {
		return err
	}

	err = ta.t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_ADD, ar, v)
	if err != nil {
		Logger.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

func (ta *TransactionWithAuth) MinerSCLock(minerID string, lock string) error {
	v, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	var mscl MinerSCLock
	mscl.ID = minerID

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, &mscl, v)
	if err != nil {
		Logger.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

// FinalizeAllocation transaction.
func (ta *TransactionWithAuth) FinalizeAllocation(allocID string, fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	type finiRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_FINALIZE_ALLOCATION, &finiRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return err
	}
	ta.t.setTransactionFee(v)
	go func() { ta.submitTxn() }()
	return nil
}

// CancelAllocation transaction.
func (ta *TransactionWithAuth) CancelAllocation(allocID string, fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	type cancelRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CANCEL_ALLOCATION, &cancelRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return err
	}
	ta.t.setTransactionFee(v)
	go func() { ta.submitTxn() }()
	return nil
}

// CreateAllocation transaction.
func (ta *TransactionWithAuth) CreateAllocation(car *CreateAllocationRequest,
	lock, fee string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	fv, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_ALLOCATION, car, lv)
	if err != nil {
		Logger.Error(err)
		return err
	}
	ta.t.setTransactionFee(fv)
	go func() { ta.submitTxn() }()
	return nil
}

// CreateReadPool for current user.
func (ta *TransactionWithAuth) CreateReadPool(fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_READ_POOL, nil, 0)
	if err != nil {
		Logger.Error(err)
		return err
	}
	ta.t.setTransactionFee(v)
	go func() { ta.submitTxn() }()
	return nil
}

// ReadPoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (ta *TransactionWithAuth) ReadPoolLock(allocID, blobberID string,
	duration int64, lock, fee string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	fv, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	type lockRequest struct {
		Duration     time.Duration `json:"duration"`
		AllocationID string        `json:"allocation_id"`
		BlobberID    string        `json:"blobber_id,omitempty"`
	}

	var lr lockRequest
	lr.Duration = time.Duration(duration)
	lr.AllocationID = allocID
	lr.BlobberID = blobberID

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_LOCK, &lr, lv)
	if err != nil {
		Logger.Error(err)
		return err
	}
	ta.t.setTransactionFee(fv)
	go func() { ta.submitTxn() }()
	return nil
}

// ReadPoolUnlock for current user and given pool.
func (ta *TransactionWithAuth) ReadPoolUnlock(poolID string, fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}
	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_UNLOCK, &unlockRequest{
			PoolID: poolID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return err
	}
	ta.t.setTransactionFee(v)
	go func() { ta.submitTxn() }()
	return nil
}

// StakePoolLock used to lock tokens in a stake pool of a blobber.
func (ta *TransactionWithAuth) StakePoolLock(blobberID string,
	lock, fee string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	fv, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	type stakePoolRequest struct {
		BlobberID string `json:"blobber_id"`
	}

	var spr stakePoolRequest
	spr.BlobberID = blobberID

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_LOCK, &spr, lv)
	if err != nil {
		Logger.Error(err)
		return err
	}
	ta.t.setTransactionFee(fv)
	go func() { ta.submitTxn() }()
	return nil
}

// StakePoolUnlock by blobberID and poolID.
func (ta *TransactionWithAuth) StakePoolUnlock(blobberID, poolID string,
	fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	type stakePoolRequest struct {
		BlobberID string `json:"blobber_id"`
		PoolID    string `json:"pool_id"`
	}

	var spr stakePoolRequest
	spr.BlobberID = blobberID
	spr.PoolID = poolID

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_UNLOCK, &spr, 0)
	if err != nil {
		Logger.Error(err)
		return err
	}
	ta.t.setTransactionFee(v)
	go func() { ta.submitTxn() }()
	return nil
}

// UpdateBlobberSettings update settings of a blobber.
func (ta *TransactionWithAuth) UpdateBlobberSettings(blob *Blobber, fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS, blob, 0)
	if err != nil {
		Logger.Error(err)
		return err
	}
	ta.t.setTransactionFee(v)
	go func() { ta.submitTxn() }()
	return nil
}

// UpdateAllocation transaction.
func (ta *TransactionWithAuth) UpdateAllocation(allocID string, sizeDiff int64,
	expirationDiff int64, lock, fee string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	fv, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	type updateAllocationRequest struct {
		ID         string `json:"id"`              // allocation id
		Size       int64  `json:"size"`            // difference
		Expiration int64  `json:"expiration_date"` // difference
	}

	var uar updateAllocationRequest
	uar.ID = allocID
	uar.Size = sizeDiff
	uar.Expiration = expirationDiff

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_ALLOCATION, &uar, lv)
	if err != nil {
		Logger.Error(err)
		return err
	}
	ta.t.setTransactionFee(fv)
	go func() { ta.submitTxn() }()
	return nil
}

// WritePoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (ta *TransactionWithAuth) WritePoolLock(allocID, blobberID string,
	duration int64, lock, fee string) error {
	lv, err := parseCoinStr(lock)
	if err != nil {
		return err
	}

	fv, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	type lockRequest struct {
		Duration     time.Duration `json:"duration"`
		AllocationID string        `json:"allocation_id"`
		BlobberID    string        `json:"blobber_id,omitempty"`
	}

	var lr lockRequest
	lr.Duration = time.Duration(duration)
	lr.AllocationID = allocID
	lr.BlobberID = blobberID

	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_LOCK, &lr, lv)
	if err != nil {
		Logger.Error(err)
		return err
	}
	ta.t.setTransactionFee(fv)
	go func() { ta.submitTxn() }()
	return nil
}

// WritePoolUnlock for current user and given pool.
func (ta *TransactionWithAuth) WritePoolUnlock(poolID string, fee string) error {
	v, err := parseCoinStr(fee)
	if err != nil {
		return err
	}

	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}
	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_UNLOCK, &unlockRequest{
			PoolID: poolID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return err
	}
	ta.t.setTransactionFee(v)
	go func() { ta.submitTxn() }()
	return nil
}
