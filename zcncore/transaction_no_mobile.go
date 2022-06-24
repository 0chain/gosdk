//go:build !mobile
// +build !mobile

package zcncore

import (
	"encoding/json"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/transaction"
)

type TransactionCommon interface {
	// ExecuteSmartContract implements wrapper for smart contract function
	ExecuteSmartContract(address, methodName string, input interface{}, val uint64) error
	// Send implements sending token to a given clientid
	Send(toClientID string, val uint64, desc string) error
	// SetTransactionFee implements method to set the transaction fee
	SetTransactionFee(txnFee uint64) error

	VestingAdd(ar *VestingAddRequest, value uint64) error

	MinerSCLock(minerID string, lock uint64) error

	FinalizeAllocation(allocID string, fee uint64) error
	CancelAllocation(allocID string, fee uint64) error
	CreateAllocation(car *CreateAllocationRequest, lock uint64, fee uint64) error //
	CreateReadPool(fee uint64) error
	ReadPoolLock(allocID string, blobberID string, duration int64, lock uint64, fee uint64) error
	ReadPoolUnlock(poolID string, fee uint64) error
	StakePoolLock(blobberID string, lock uint64, fee uint64) error
	StakePoolUnlock(blobberID string, poolID string, fee uint64) error
	UpdateBlobberSettings(blobber *Blobber, fee uint64) error
	UpdateAllocation(allocID string, sizeDiff int64, expirationDiff int64, lock uint64, fee uint64) error
	WritePoolLock(allocID string, blobberID string, duration int64, lock uint64, fee uint64) error
	WritePoolUnlock(poolID string, fee uint64) error
}

// NewTransaction allocation new generic transaction object for any operation
func NewTransaction(cb TransactionCallback, txnFee uint64, nonce int64) (TransactionScheme, error) {
	err := CheckConfig()
	if err != nil {
		return nil, err
	}
	if _config.isSplitWallet {
		if _config.authUrl == "" {
			return nil, errors.New("", "auth url not set")
		}
		Logger.Info("New transaction interface with auth")
		return newTransactionWithAuth(cb, txnFee, nonce)
	}
	Logger.Info("New transaction interface")
	t, err := newTransaction(cb, txnFee, nonce)
	return t, err
}

func (t *Transaction) ExecuteSmartContract(address, methodName string, input interface{}, val uint64) error {
	err := t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return err
	}
	go func() {
		t.setNonceAndSubmit()
	}()
	return nil
}

func (t *Transaction) SetTransactionFee(txnFee uint64) error {
	if t.txnStatus != StatusUnknown {
		return errors.New("", "transaction already exists. cannot set transaction fee.")
	}
	t.txn.TransactionFee = txnFee
	return nil
}

func (t *Transaction) Send(toClientID string, val uint64, desc string) error {
	txnData, err := json.Marshal(SendTxnData{Note: desc})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSend
		t.txn.ToClientID = toClientID
		t.txn.Value = val
		t.txn.TransactionData = string(txnData)
		t.setNonceAndSubmit()
	}()
	return nil
}

func (t *Transaction) SendWithSignatureHash(toClientID string, val uint64, desc string, sig string, CreationDate int64, hash string) error {
	txnData, err := json.Marshal(SendTxnData{Note: desc})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}
	go func() {
		t.txn.TransactionType = transaction.TxnTypeSend
		t.txn.ToClientID = toClientID
		t.txn.Value = val
		t.txn.Hash = hash
		t.txn.TransactionData = string(txnData)
		t.txn.Signature = sig
		t.txn.CreationDate = CreationDate
		t.setNonceAndSubmit()
	}()
	return nil
}

func (t *Transaction) VestingAdd(ar *VestingAddRequest, value uint64) (
	err error) {

	err = t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_ADD, ar, value)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

func (t *Transaction) MinerSCLock(nodeID string, lock uint64) (err error) {

	var mscl MinerSCLock
	mscl.ID = nodeID

	err = t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, &mscl, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	go func() { t.setNonceAndSubmit() }()
	return
}

// FinalizeAllocation transaction.
func (t *Transaction) FinalizeAllocation(allocID string, fee uint64) (
	err error) {

	type finiRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_FINALIZE_ALLOCATION, &finiRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// CancelAllocation transaction.
func (t *Transaction) CancelAllocation(allocID string, fee uint64) (
	err error) {

	type cancelRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CANCEL_ALLOCATION, &cancelRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// CreateAllocation transaction.
func (t *Transaction) CreateAllocation(car *CreateAllocationRequest,
	lock uint64, fee uint64) (err error) {

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_ALLOCATION, car, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// CreateReadPool for current user.
func (t *Transaction) CreateReadPool(fee uint64) (err error) {

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_READ_POOL, nil, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// ReadPoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (t *Transaction) ReadPoolLock(allocID, blobberID string,
	duration int64, lock, fee uint64) (err error) {

	type lockRequest struct {
		Duration     time.Duration `json:"duration"`
		AllocationID string        `json:"allocation_id"`
		BlobberID    string        `json:"blobber_id,omitempty"`
	}

	var lr lockRequest
	lr.Duration = time.Duration(duration)
	lr.AllocationID = allocID
	lr.BlobberID = blobberID

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_LOCK, &lr, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// ReadPoolUnlock for current user and given pool.
func (t *Transaction) ReadPoolUnlock(poolID string, fee uint64) (err error) {
	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_UNLOCK, &unlockRequest{
			PoolID: poolID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// StakePoolLock used to lock tokens in a stake pool of a blobber.
func (t *Transaction) StakePoolLock(blobberID string, lock, fee uint64) (
	err error) {

	type stakePoolRequest struct {
		BlobberID string `json:"blobber_id"`
	}

	var spr stakePoolRequest
	spr.BlobberID = blobberID

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_LOCK, &spr, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// StakePoolUnlock by blobberID and poolID.
func (t *Transaction) StakePoolUnlock(blobberID, poolID string,
	fee uint64) (err error) {

	type stakePoolRequest struct {
		BlobberID string `json:"blobber_id"`
		PoolID    string `json:"pool_id"`
	}

	var spr stakePoolRequest
	spr.BlobberID = blobberID
	spr.PoolID = poolID

	err = t.createSmartContractTxn(StorageSmartContractAddress, transaction.STORAGESC_STAKE_POOL_UNLOCK, &spr, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// UpdateBlobberSettings update settings of a blobber.
func (t *Transaction) UpdateBlobberSettings(b *Blobber, fee uint64) (err error) {

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS, b, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// UpdateAllocation transaction.
func (t *Transaction) UpdateAllocation(allocID string, sizeDiff int64,
	expirationDiff int64, lock, fee uint64) (err error) {

	type updateAllocationRequest struct {
		ID         string `json:"id"`              // allocation id
		Size       int64  `json:"size"`            // difference
		Expiration int64  `json:"expiration_date"` // difference
	}

	var uar updateAllocationRequest
	uar.ID = allocID
	uar.Size = sizeDiff
	uar.Expiration = expirationDiff

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_ALLOCATION, &uar, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// WritePoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (t *Transaction) WritePoolLock(allocID, blobberID string, duration int64,
	lock, fee uint64) (err error) {

	type lockRequest struct {
		Duration     time.Duration `json:"duration"`
		AllocationID string        `json:"allocation_id"`
		BlobberID    string        `json:"blobber_id,omitempty"`
	}

	var lr lockRequest
	lr.Duration = time.Duration(duration)
	lr.AllocationID = allocID
	lr.BlobberID = blobberID

	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_LOCK, &lr, lock)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// WritePoolUnlock for current user and given pool.
func (t *Transaction) WritePoolUnlock(poolID string, fee uint64) (
	err error) {

	type unlockRequest struct {
		PoolID string `json:"pool_id"`
	}
	err = t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_UNLOCK, &unlockRequest{
			PoolID: poolID,
		}, 0)
	if err != nil {
		Logger.Error(err)
		return
	}
	t.SetTransactionFee(fee)
	go func() { t.setNonceAndSubmit() }()
	return
}

// ConvertToValue converts ZCN tokens to value
func ConvertToValue(token float64) uint64 {
	return uint64(token * float64(TOKEN_UNIT))
}
