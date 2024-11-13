//go:build mobile
// +build mobile

package zcncore

import (
	"encoding/json"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/transaction"
)

func newTransactionWithAuth(cb TransactionCallback, txnFee string, nonce int64) (*TransactionWithAuth, error) {
	ta := &TransactionWithAuth{}
	var err error
	ta.t, err = newTransaction(cb, txnFee, nonce)
	return ta, err
}

func (ta *TransactionWithAuth) GetDetails() *transaction.Transaction {
	return ta.t.txn
}

func (ta *TransactionWithAuth) ExecuteSmartContract(address string, methodName string, input string, val string) error {
	err := ta.t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return err
	}
	go func() {
		ta.submitTxn()
	}()

	return nil
}

func (ta *TransactionWithAuth) Send(toClientID string, val string, desc string) error {
	txnData, err := json.Marshal(SendTxnData{Note: desc})
	if err != nil {
		return errors.New("", "Could not serialize description to transaction_data")
	}
	go func() {
		ta.t.txn.TransactionType = transaction.TxnTypeSend
		ta.t.txn.ToClientID = toClientID
		ta.t.txn.Value = val
		ta.t.txn.TransactionData = string(txnData)
		ta.submitTxn()
	}()
	return nil
}

func (ta *TransactionWithAuth) VestingAdd(ar VestingAddRequest, value string) error {
	err := ta.t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_ADD, ar, value)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

func (ta *TransactionWithAuth) MinerSCLock(providerId string, providerType int, lock string) error {
	pr := stakePoolRequest{
		ProviderType: providerType,
		ProviderID:   providerId,
	}

	err := ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, &pr, lock)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

func (ta *TransactionWithAuth) MinerSCUnlock(providerId string, providerType int) error {
	pr := &stakePoolRequest{
		ProviderID:   providerId,
		ProviderType: providerType,
	}
	err := ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_LOCK, pr, "0")
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return err
}

// FinalizeAllocation transaction.
func (ta *TransactionWithAuth) FinalizeAllocation(allocID string) error {
	type finiRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_FINALIZE_ALLOCATION, &finiRequest{
			AllocationID: allocID,
		}, "0")
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

// CancelAllocation transaction.
func (ta *TransactionWithAuth) CancelAllocation(allocID string) error {
	type cancelRequest struct {
		AllocationID string `json:"allocation_id"`
	}
	err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CANCEL_ALLOCATION, &cancelRequest{
			AllocationID: allocID,
		}, "0")
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

// CreateAllocation transaction.
func (ta *TransactionWithAuth) CreateAllocation(car *CreateAllocationRequest,
	lock string) error {
	err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_ALLOCATION, car, lock)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

// CreateReadPool for current user.
func (ta *TransactionWithAuth) CreateReadPool() error {
	if err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_CREATE_READ_POOL, nil, "0"); err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

// ReadPoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (ta *TransactionWithAuth) ReadPoolLock(allocID, blobberID string,
	duration int64, lock string) error {
	type lockRequest struct {
		Duration     time.Duration `json:"duration"`
		AllocationID string        `json:"allocation_id"`
		BlobberID    string        `json:"blobber_id,omitempty"`
	}

	var lr lockRequest
	lr.Duration = time.Duration(duration)
	lr.AllocationID = allocID
	lr.BlobberID = blobberID

	err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_LOCK, &lr, lock)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

// ReadPoolUnlock for current user and given pool.
func (ta *TransactionWithAuth) ReadPoolUnlock() error {
	if err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_READ_POOL_UNLOCK, nil, "0"); err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

// StakePoolLock used to lock tokens in a stake pool of a blobber.
func (ta *TransactionWithAuth) StakePoolLock(providerId string, providerType int,
	lock string) error {
	type stakePoolRequest struct {
		ProviderType int    `json:"provider_type,omitempty"`
		ProviderID   string `json:"provider_id,omitempty"`
	}

	spr := stakePoolRequest{
		ProviderType: providerType,
		ProviderID:   providerId,
	}
	err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_LOCK, &spr, lock)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

// StakePoolUnlock by blobberID
func (ta *TransactionWithAuth) StakePoolUnlock(providerId string, providerType int) error {
	spr := stakePoolRequest{
		ProviderType: providerType,
		ProviderID:   providerId,
	}

	if err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_STAKE_POOL_UNLOCK, &spr, "0"); err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

// UpdateBlobberSettings update settings of a blobber.
func (ta *TransactionWithAuth) UpdateBlobberSettings(blob Blobber) error {
	if err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_BLOBBER_SETTINGS, blob, "0"); err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

// UpdateAllocation transaction.
func (ta *TransactionWithAuth) UpdateAllocation(allocID string, sizeDiff int64,
	expirationDiff int64, lock string) error {
	type updateAllocationRequest struct {
		ID         string `json:"id"`              // allocation id
		Size       int64  `json:"size"`            // difference
		Expiration int64  `json:"expiration_date"` // difference
	}

	var uar updateAllocationRequest
	uar.ID = allocID
	uar.Size = sizeDiff
	uar.Expiration = expirationDiff

	err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_ALLOCATION, &uar, lock)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

// WritePoolLock locks tokens for current user and given allocation, using given
// duration. If blobberID is not empty, then tokens will be locked for given
// allocation->blobber only.
func (ta *TransactionWithAuth) WritePoolLock(allocID, lock string) error {
	var lr = struct {
		AllocationID string `json:"allocation_id"`
	}{
		AllocationID: allocID,
	}

	err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_LOCK, &lr, lock)
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

// WritePoolUnlock for current user and given pool.
func (ta *TransactionWithAuth) WritePoolUnlock(allocID string) error {
	var ur = struct {
		AllocationID string `json:"allocation_id"`
	}{
		AllocationID: allocID,
	}

	if err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_WRITE_POOL_UNLOCK, &ur, "0"); err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return nil
}

func (ta *TransactionWithAuth) MinerSCCollectReward(providerId string, providerType int) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: providerType,
	}
	err := ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_COLLECT_REWARD, pr, "0")
	if err != nil {
		logging.Error(err)
		return err
	}
	go ta.submitTxn()
	return err
}

func (ta *TransactionWithAuth) StorageSCCollectReward(providerId string, providerType int) error {
	pr := &scCollectReward{
		ProviderId:   providerId,
		ProviderType: providerType,
	}
	err := ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_COLLECT_REWARD, pr, "0")
	if err != nil {
		logging.Error(err)
		return err
	}
	go func() { ta.submitTxn() }()
	return err
}

func (ta *TransactionWithAuth) VestingUpdateConfig(ip InputMap) (err error) {
	err = ta.t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_UPDATE_SETTINGS, ip, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

// faucet smart contract

func (ta *TransactionWithAuth) FaucetUpdateConfig(ip InputMap) (err error) {
	err = ta.t.createSmartContractTxn(FaucetSmartContractAddress,
		transaction.FAUCETSC_UPDATE_SETTINGS, ip, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerScUpdateConfig(ip InputMap) (err error) {
	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_SETTINGS, ip, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerScUpdateGlobals(ip InputMap) (err error) {
	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_UPDATE_GLOBALS, ip, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) StorageScUpdateConfig(ip InputMap) (err error) {
	err = ta.t.createSmartContractTxn(StorageSmartContractAddress,
		transaction.STORAGESC_UPDATE_SETTINGS, ip, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) ZCNSCUpdateGlobalConfig(ip InputMap) (err error) {
	err = ta.t.createSmartContractTxn(ZCNSCSmartContractAddress,
		transaction.ZCNSC_UPDATE_GLOBAL_CONFIG, ip, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.submitTxn()
	return
}

func (ta *TransactionWithAuth) GetVerifyConfirmationStatus() int {
	return ta.t.GetVerifyConfirmationStatus()
}

func (ta *TransactionWithAuth) MinerSCMinerSettings(info MinerSCMinerInfo) (
	err error) {
	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_SETTINGS, info, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCSharderSettings(info MinerSCMinerInfo) (
	err error) {

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_SETTINGS, info, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCDeleteMiner(info MinerSCMinerInfo) (
	err error) {

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_MINER_DELETE, info, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) MinerSCDeleteSharder(info MinerSCMinerInfo) (
	err error) {

	err = ta.t.createSmartContractTxn(MinerSmartContractAddress,
		transaction.MINERSC_SHARDER_DELETE, info, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) ZCNSCUpdateAuthorizerConfig(ip AuthorizerNode) (err error) {
	err = ta.t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_UPDATE_AUTHORIZER_CONFIG, ip, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.submitTxn()
	return
}

func (ta *TransactionWithAuth) ZCNSCAddAuthorizer(ip AddAuthorizerPayload) (err error) {
	err = ta.t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_ADD_AUTHORIZER, ip, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.submitTxn()
	return
}

func (ta *TransactionWithAuth) ZCNSCAuthorizerHealthCheck(ip *AuthorizerHealthCheckPayload) (err error) {
	err = ta.t.createSmartContractTxn(ZCNSCSmartContractAddress, transaction.ZCNSC_AUTHORIZER_HEALTH_CHECK, ip, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go ta.t.setNonceAndSubmit()
	return
}

func (ta *TransactionWithAuth) VestingTrigger(poolID string) (err error) {
	err = ta.t.vestingPoolTxn(transaction.VESTING_TRIGGER, poolID, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) VestingStop(sr *VestingStopRequest) (err error) {
	err = ta.t.createSmartContractTxn(VestingSmartContractAddress,
		transaction.VESTING_STOP, sr, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) VestingUnlock(poolID string) (err error) {

	err = ta.t.vestingPoolTxn(transaction.VESTING_UNLOCK, poolID, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}

func (ta *TransactionWithAuth) VestingDelete(poolID string) (err error) {
	err = ta.t.vestingPoolTxn(transaction.VESTING_DELETE, poolID, "0")
	if err != nil {
		logging.Error(err)
		return
	}
	go func() { ta.submitTxn() }()
	return
}
