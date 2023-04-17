//go:build !mobile
// +build !mobile

package zcncore

import (
	"encoding/json"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/transaction"
)

func (ta *TransactionWithAuth) ExecuteSmartContract(address, methodName string,
	input interface{}, val uint64, feeOpts ...FeeOption) (*transaction.Transaction, error) {
	err := ta.createSmartContractTxn(address, methodName, input, val, feeOpts...)
	if err != nil {
		return nil, err
	}
	go func() {
		ta.submitTxn()
	}()
	return ta.txn, nil
}

//func (ta *TransactionWithAuth) SetTransactionFee(txnFee uint64) error {
//	return ta.SetTransactionFee(txnFee)
//}

func (ta *TransactionWithAuth) Send(toClientID string, val uint64, desc string) error {
	txnData, err := json.Marshal(SendTxnData{Note: desc})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}
	go func() {
		ta.txn.TransactionType = transaction.TxnTypeSend
		ta.txn.ToClientID = toClientID
		ta.txn.Value = val
		ta.txn.TransactionData = string(txnData)
		ta.submitTxn()
	}()
	return nil
}

func (ta *TransactionWithAuth) VestingAdd(ar *VestingAddRequest,
	value uint64) (err error) {

	err = ta.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_ADD, ar, value)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCLock(providerId string, providerType Provider, lock uint64) error {
	pr := &stakePoolRequest{
		ProviderID:   providerId,
		ProviderType: providerType,
	}
	err := ta.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, pr, lock)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

func (ta *TransactionWithAuth) MinerSCUnlock(providerId string, providerType Provider) error {
	pr := &stakePoolRequest{
		ProviderID:   providerId,
		ProviderType: providerType,
	}
	err := ta.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, pr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return err
}

// FinalizeAllocation transaction.
func (ta *TransactionWithAuth) FinalizeAllocation(allocID string) (
	err error) {

	type finiRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err = ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_FINALIZE_ALLOCATION, &finiRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// CancelAllocation transaction.
func (ta *TransactionWithAuth) CancelAllocation(allocID string) (
	err error) {

	type cancelRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err = ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CANCEL_ALLOCATION, &cancelRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// CreateAllocation transaction.
func (ta *TransactionWithAuth) CreateAllocation(car *CreateAllocationRequest,
	lock uint64) (err error) {

	err = ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_ALLOCATION, car, lock)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// CreateReadPool for current user.
func (ta *TransactionWithAuth) CreateReadPool() (err error) {

	err = ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_READ_POOL, nil, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// ReadPoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (ta *TransactionWithAuth) ReadPoolLock(allocID, blobberID string,
	duration int64, lock uint64) (err error) {

	type lockRequest struct {
		Duration     time.Duration `json:"duration"`
		AllocationID string        `json:"allocation_id"`
		BlobberID    string        `json:"blobber_id,omitempty"`
	}

	var lr lockRequest
	lr.Duration = time.Duration(duration)
	lr.AllocationID = allocID
	lr.BlobberID = blobberID

	err = ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_LOCK, &lr, lock)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// ReadPoolUnlock for current user and given pool.
func (ta *TransactionWithAuth) ReadPoolUnlock() (
	err error) {

	err = ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_UNLOCK, nil, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// StakePoolLock used to lock tokens in a stake pool of a blobber.
func (ta *TransactionWithAuth) StakePoolLock(providerId string, providerType Provider, lock uint64) error {

	type stakePoolRequest struct {
		ProviderType Provider `json:"provider_type,omitempty"`
		ProviderID   string   `json:"provider_id,omitempty"`
	}

	spr := stakePoolRequest{
		ProviderType: providerType,
		ProviderID:   providerId,
	}

	err := ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_LOCK, &spr, lock)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

// StakePoolUnlock by blobberID
func (ta *TransactionWithAuth) StakePoolUnlock(providerId string, providerType Provider) error {

	type stakePoolRequest struct {
		ProviderType Provider `json:"provider_type,omitempty"`
		ProviderID   string   `json:"provider_id,omitempty"`
	}

	spr := stakePoolRequest{
		ProviderType: providerType,
		ProviderID:   providerId,
	}

	err := ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_UNLOCK, &spr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

// UpdateBlobberSettings update settings of a blobber.
func (ta *TransactionWithAuth) UpdateBlobberSettings(blob *Blobber) (
	err error) {

	err = ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS, blob, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// UpdateValidatorSettings update settings of a validator.
func (ta *TransactionWithAuth) UpdateValidatorSettings(v *Validator) (
	err error) {

	err = ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_VALIDATOR_SETTINGS, v, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// UpdateAllocation transaction.
func (ta *TransactionWithAuth) UpdateAllocation(allocID string, sizeDiff int64,
	expirationDiff int64, lock uint64) (err error) {

	type updateAllocationRequest struct {
		ID         string `json:"id"`              // allocation id
		Size       int64  `json:"size"`            // difference
		Expiration int64  `json:"expiration_date"` // difference
	}

	var uar updateAllocationRequest
	uar.ID = allocID
	uar.Size = sizeDiff
	uar.Expiration = expirationDiff

	err = ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_ALLOCATION, &uar, lock)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// WritePoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (ta *TransactionWithAuth) WritePoolLock(allocID, blobberID string,
	duration int64, lock uint64) (err error) {

	type lockRequest struct {
		Duration     time.Duration `json:"duration"`
		AllocationID string        `json:"allocation_id"`
		BlobberID    string        `json:"blobber_id,omitempty"`
	}

	var lr lockRequest
	lr.Duration = time.Duration(duration)
	lr.AllocationID = allocID
	lr.BlobberID = blobberID

	err = ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_LOCK, &lr, lock)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// WritePoolUnlock for current user and given pool.
func (ta *TransactionWithAuth) WritePoolUnlock(allocID string) (err error) {
	type unlockRequest struct {
		AllocationID string `json:"allocation_id"`
	}

	err = ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_UNLOCK, &unlockRequest{
			AllocationID: allocID,
		}, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCCollectReward(providerId string, providerType Provider) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: int(providerType),
	}
	err := ta.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_COLLECT_REWARD, pr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go ta.submitTxn()
	return err
}

func (ta *TransactionWithAuth) StorageSCCollectReward(providerId string, providerType Provider) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: int(providerType),
	}
	err := ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_COLLECT_REWARD, pr, 0)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return err
}

func (ta *TransactionWithAuth) VestingUpdateConfig(ip *InputMap) (err error) {
	err = ta.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// faucet smart contract

func (ta *TransactionWithAuth) FaucetUpdateConfig(ip *InputMap) (err error) {
	err = ta.createSmartContractTxn(FaucetSmartContractAddress,
		transaction.FAUCETSC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerScUpdateConfig(ip *InputMap) (err error) {
	err = ta.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerScUpdateGlobals(ip *InputMap) (err error) {
	err = ta.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_GLOBALS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) StorageScUpdateConfig(ip *InputMap) (err error) {
	err = ta.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_SETTINGS, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) ZCNSCUpdateGlobalConfig(ip *InputMap) (err error) {
	err = ta.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_UPDATE_GLOBAL_CONFIG, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.submitTxn()
	return
}

func (ta *TransactionWithAuth) GetVerifyConfirmationStatus() ConfirmationStatus {
	return ta.GetVerifyConfirmationStatus()
}

func (ta *TransactionWithAuth) MinerSCMinerSettings(info *MinerSCMinerInfo) (
	err error) {

	err = ta.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_SETTINGS, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCSharderSettings(info *MinerSCMinerInfo) (
	err error) {

	err = ta.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_SETTINGS, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCDeleteMiner(info *MinerSCMinerInfo) (
	err error) {

	err = ta.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_DELETE, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCDeleteSharder(info *MinerSCMinerInfo) (
	err error) {

	err = ta.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_DELETE, info, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) ZCNSCUpdateAuthorizerConfig(ip *AuthorizerNode) (err error) {
	err = ta.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_UPDATE_AUTHORIZER_CONFIG, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.submitTxn()
	return
}

func (ta *TransactionWithAuth) ZCNSCAddAuthorizer(ip *AddAuthorizerPayload) (err error) {
	err = ta.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_ADD_AUTHORIZER, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.submitTxn()
	return
}

func (ta *TransactionWithAuth) ZCNSCAuthorizerHealthCheck(ip *AuthorizerHealthCheckPayload) (err error) {
	err = ta.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_AUTHORIZER_HEALTH_CHECK, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.setNonceAndSubmit()
	return
}

func (ta *TransactionWithAuth) ZCNSCDeleteAuthorizer(ip *DeleteAuthorizerPayload) (err error) {
	err = ta.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_DELETE_AUTHORIZER, ip, 0)
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.submitTxn()
	return
}
