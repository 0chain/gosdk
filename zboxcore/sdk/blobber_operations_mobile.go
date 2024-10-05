//go:build mobile
// +build mobile

package sdk

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/client"
)

// CreateAllocationForOwner creates a new allocation with the given options (txn: `storagesc.new_allocation_request`).
//
//   - owner is the client id of the owner of the allocation.
//   - ownerpublickey is the public key of the owner of the allocation.
//   - datashards is the number of data shards for the allocation.
//   - parityshards is the number of parity shards for the allocation.
//   - size is the size of the allocation.
//   - readPrice is the read price range for the allocation (Reads in Züs are free!).
//   - writePrice is the write price range for the allocation.
//   - lock is the lock value for the transaction (how much tokens to lock to the allocation, in SAS).
//   - preferredBlobberIds is a list of preferred blobber ids for the allocation.
//   - thirdPartyExtendable is a flag indicating whether the allocation can be extended by a third party.
//   - fileOptionsParams is the file options parameters for the allocation, which control the usage permissions of the files in the allocation.
//
// returns the hash of the transaction, the nonce of the transaction, the transaction object and an error if any.
func CreateAllocationForOwner(
	owner, ownerpublickey string,
	datashards, parityshards int, size int64,
	readPrice, writePrice PriceRange,
	lock uint64, preferredBlobberIds, blobberAuthTickets []string, thirdPartyExtendable, IsEnterprise, force bool, fileOptionsParams *FileOptionsParameters,
) (hash string, nonce int64, txn *transaction.Transaction, err error) {

	if lock > math.MaxInt64 {
		return "", 0, nil, errors.New("invalid_lock", "int64 overflow on lock value")
	}

	if datashards < 1 || parityshards < 1 {
		return "", 0, nil, errors.New("allocation_validation_failed", "atleast 1 data and 1 parity shards are required")
	}

	allocationRequest, err := getNewAllocationBlobbers(
		datashards, parityshards, size, readPrice, writePrice, preferredBlobberIds, blobberAuthTickets, force)
	if err != nil {
		return "", 0, nil, errors.New("failed_get_allocation_blobbers", "failed to get blobbers for allocation: "+err.Error())
	}

	if !sdkInitialized {
		return "", 0, nil, sdkNotInitialized
	}

	allocationRequest["owner_id"] = owner
	allocationRequest["owner_public_key"] = ownerpublickey
	allocationRequest["third_party_extendable"] = thirdPartyExtendable
	allocationRequest["file_options_changed"], allocationRequest["file_options"] = calculateAllocationFileOptions(63 /*0011 1111*/, fileOptionsParams)
	allocationRequest["is_enterprise"] = IsEnterprise

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.NEW_ALLOCATION_REQUEST,
		InputArgs: allocationRequest,
	}
	hash, _, nonce, txn, err = storageSmartContractTxnValue(sn, strconv.FormatUint(lock, 10))
	return
}

// CreateFreeAllocation creates a new free allocation (txn: `storagesc.free_allocation_request`).
//   - marker is the marker for the free allocation.
//   - value is the value of the free allocation.
//
// returns the hash of the transaction, the nonce of the transaction and an error if any.
func CreateFreeAllocation(marker string, value string) (string, int64, error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	recipientPublicKey := client.GetClientPublicKey()

	var input = map[string]interface{}{
		"recipient_public_key": recipientPublicKey,
		"marker":               marker,
	}

	blobbers, err := GetFreeAllocationBlobbers(input)
	if err != nil {
		return "", 0, err
	}

	input["blobbers"] = blobbers

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.NEW_FREE_ALLOCATION,
		InputArgs: input,
	}
	hash, _, n, _, err := storageSmartContractTxnValue(sn, value)
	return hash, n, err
}

// UpdateAllocation sends an update request for an allocation (txn: `storagesc.update_allocation_request`)
//
//   - size is the size of the allocation.
//   - extend is a flag indicating whether to extend the allocation.
//   - allocationID is the id of the allocation.
//   - lock is the lock value for the transaction (how much tokens to lock to the allocation, in SAS).
//   - addBlobberId is the id of the blobber to add to the allocation.
//   - addBlobberAuthTicket is the auth ticket of the blobber to add to the allocation, in case the blobber is restricted.
//   - removeBlobberId is the id of the blobber to remove from the allocation.
//   - setThirdPartyExtendable is a flag indicating whether the allocation can be extended by a third party.
//   - fileOptionsParams is the file options parameters for the allocation, which control the usage permissions of the files in the allocation.
//
// returns the hash of the transaction, the nonce of the transaction and an error if any.
func UpdateAllocation(
	size int64,
	extend bool,
	allocationID string,
	lock uint64,
	addBlobberId, addBlobberAuthTicket, removeBlobberId string,
	setThirdPartyExtendable bool, fileOptionsParams *FileOptionsParameters,
) (hash string, nonce int64, err error) {

	if lock > math.MaxInt64 {
		return "", 0, errors.New("invalid_lock", "int64 overflow on lock value")
	}

	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	alloc, err := GetAllocation(allocationID)
	if err != nil {
		return "", 0, allocationNotFound
	}

	updateAllocationRequest := make(map[string]interface{})
	updateAllocationRequest["owner_id"] = client.GetClientID()
	updateAllocationRequest["owner_public_key"] = ""
	updateAllocationRequest["id"] = allocationID
	updateAllocationRequest["size"] = size
	updateAllocationRequest["extend"] = extend
	updateAllocationRequest["add_blobber_id"] = addBlobberId
	updateAllocationRequest["add_blobber_auth_ticket"] = addBlobberAuthTicket
	updateAllocationRequest["remove_blobber_id"] = removeBlobberId
	updateAllocationRequest["set_third_party_extendable"] = setThirdPartyExtendable
	updateAllocationRequest["file_options_changed"], updateAllocationRequest["file_options"] = calculateAllocationFileOptions(alloc.FileOptions, fileOptionsParams)

	sn := transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_ALLOCATION,
		InputArgs: updateAllocationRequest,
	}
	hash, _, nonce, _, err = storageSmartContractTxnValue(sn, strconv.FormatUint(lock, 10))
	return
}

// StakePoolLock locks tokens in a stake pool.
// This function is the entry point for the staking operation.
// Provided the provider type and provider ID, the value is locked in the stake pool between the SDK client and the provider.
// Based on the locked amount, the client will get rewards as share of the provider's rewards.
//   - providerType: provider type
//   - providerID: provider ID
//   - value: value to lock
//   - fee: transaction fee
func StakePoolLock(providerType ProviderType, providerID string, value, fee string) (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	if providerType == 0 {
		return "", 0, errors.New("stake_pool_lock", "provider is required")
	}

	if providerID == "" {
		return "", 0, errors.New("stake_pool_lock", "provider_id is required")
	}

	spr := stakePoolRequest{
		ProviderType: providerType,
		ProviderID:   providerID,
	}

	var sn = transaction.SmartContractTxnData{
		InputArgs: &spr,
	}

	var scAddress string
	switch providerType {
	case ProviderBlobber, ProviderValidator:
		scAddress = STORAGE_SCADDRESS
		sn.Name = transaction.STORAGESC_STAKE_POOL_LOCK
	case ProviderMiner, ProviderSharder:
		scAddress = MINERSC_SCADDRESS
		sn.Name = transaction.MINERSC_LOCK
	case ProviderAuthorizer:
		scAddress = ZCNSC_SCADDRESS
		sn.Name = transaction.ZCNSC_LOCK
	default:
		return "", 0, errors.Newf("stake_pool_lock", "unsupported provider type: %v", providerType)
	}

	hash, _, nonce, _, err = smartContractTxnValueFeeWithRetry(scAddress, sn, value, fee)
	return
}

// StakePoolUnlock unlocks a stake pool tokens. If tokens can't be unlocked due
// to opened offers, then it returns time where the tokens can be unlocked,
// marking the pool as 'want to unlock' to avoid its usage in offers in the
// future. The time is maximal time that can be lesser in some cases. To
// unlock tokens can't be unlocked now, wait the time and unlock them (call
// this function again).
//   - providerType: provider type
//   - providerID: provider ID
//   - fee: transaction fee
func StakePoolUnlock(providerType ProviderType, providerID string, fee string) (unstake int64, nonce int64, err error) {
	if !sdkInitialized {
		return 0, 0, sdkNotInitialized
	}

	if providerType == 0 {
		return 0, 0, errors.New("stake_pool_lock", "provider is required")
	}

	if providerID == "" {
		return 0, 0, errors.New("stake_pool_lock", "provider_id is required")
	}

	spr := stakePoolRequest{
		ProviderType: providerType,
		ProviderID:   providerID,
	}

	var sn = transaction.SmartContractTxnData{
		InputArgs: &spr,
	}

	var scAddress string
	switch providerType {
	case ProviderBlobber, ProviderValidator:
		scAddress = STORAGE_SCADDRESS
		sn.Name = transaction.STORAGESC_STAKE_POOL_UNLOCK
	case ProviderMiner, ProviderSharder:
		scAddress = MINERSC_SCADDRESS
		sn.Name = transaction.MINERSC_UNLOCK
	case ProviderAuthorizer:
		scAddress = ZCNSC_SCADDRESS
		sn.Name = transaction.ZCNSC_UNLOCK
	default:
		return 0, 0, errors.Newf("stake_pool_unlock", "unsupported provider type: %v", providerType)
	}

	var out string
	if _, out, nonce, _, err = smartContractTxnValueFeeWithRetry(scAddress, sn, "0", fee); err != nil {
		return // an error
	}

	var spuu stakePoolLock
	if err = json.Unmarshal([]byte(out), &spuu); err != nil {
		return
	}

	return spuu.Amount, nonce, nil
}

// ReadPoolLock locks given number of tokes for given duration in read pool.
//   - tokens: number of tokens to lock
//   - fee: transaction fee
func ReadPoolLock(tokens, fee string) (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_READ_POOL_LOCK,
		InputArgs: nil,
	}
	hash, _, nonce, _, err = smartContractTxnValueFeeWithRetry(STORAGE_SCADDRESS, sn, tokens, fee)
	return
}

// ReadPoolUnlock unlocks tokens in expired read pool
//   - fee: transaction fee
func ReadPoolUnlock(fee string) (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_READ_POOL_UNLOCK,
		InputArgs: nil,
	}
	hash, _, nonce, _, err = smartContractTxnValueFeeWithRetry(STORAGE_SCADDRESS, sn, "0", fee)
	return
}

//
// write pool
//

// WritePoolLock locks given number of tokes for given duration in read pool.
//   - allocID: allocation ID
//   - tokens: number of tokens to lock
//   - fee: transaction fee
func WritePoolLock(allocID string, tokens, fee string) (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	type lockRequest struct {
		AllocationID string `json:"allocation_id"`
	}

	var req lockRequest
	req.AllocationID = allocID

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_WRITE_POOL_LOCK,
		InputArgs: &req,
	}

	hash, _, nonce, _, err = smartContractTxnValueFeeWithRetry(STORAGE_SCADDRESS, sn, tokens, fee)
	return
}

// WritePoolUnlock unlocks ALL tokens of a write pool. Needs to be cancelled first.
//   - allocID: allocation ID
//   - fee: transaction fee
func WritePoolUnlock(allocID string, fee string) (hash string, nonce int64, err error) {
	if !sdkInitialized {
		return "", 0, sdkNotInitialized
	}

	type unlockRequest struct {
		AllocationID string `json:"allocation_id"`
	}

	var req unlockRequest
	req.AllocationID = allocID

	var sn = transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_WRITE_POOL_UNLOCK,
		InputArgs: &req,
	}
	hash, _, nonce, _, err = smartContractTxnValueFeeWithRetry(STORAGE_SCADDRESS, sn, "0", fee)
	return
}

func smartContractTxn(scAddress string, sn transaction.SmartContractTxnData) (
	hash, out string, nonce int64, txn *transaction.Transaction, err error) {
	return smartContractTxnValue(scAddress, sn, "0")
}

func StorageSmartContractTxn(sn transaction.SmartContractTxnData) (
	hash, out string, nonce int64, txn *transaction.Transaction, err error) {

	return storageSmartContractTxnValue(sn, "0")
}

func storageSmartContractTxn(sn transaction.SmartContractTxnData) (
	hash, out string, nonce int64, txn *transaction.Transaction, err error) {

	return storageSmartContractTxnValue(sn, "0")
}

func smartContractTxnValue(scAddress string, sn transaction.SmartContractTxnData, value string) (
	hash, out string, nonce int64, txn *transaction.Transaction, err error) {

	return smartContractTxnValueFeeWithRetry(scAddress, sn, value, strconv.FormatUint(client.TxnFee(), 10))
}

func storageSmartContractTxnValue(sn transaction.SmartContractTxnData, value string) (
	hash, out string, nonce int64, txn *transaction.Transaction, err error) {

	// Fee is set during sdk initialization.
	return smartContractTxnValueFeeWithRetry(STORAGE_SCADDRESS, sn, value, strconv.FormatUint(client.TxnFee(), 10))
}

func smartContractTxnValueFeeWithRetry(scAddress string, sn transaction.SmartContractTxnData,
	value, fee string) (hash, out string, nonce int64, t *transaction.Transaction, err error) {
	hash, out, nonce, t, err = smartContractTxnValueFee(scAddress, sn, value, fee)

	if err != nil && strings.Contains(err.Error(), "invalid transaction nonce") {
		return smartContractTxnValueFee(scAddress, sn, value, fee)
	}

	return
}

func smartContractTxnValueFee(scAddress string, sn transaction.SmartContractTxnData,
	value, fee string) (hash, out string, nonce int64, t *transaction.Transaction, err error) {
	t, err = ExecuteSmartContract(scAddress, sn, value, fee)
	if err != nil {
		if t != nil {
			return "", "", t.TransactionNonce, nil, err
		}

		return "", "", 0, nil, err
	}

	return t.Hash, t.TransactionOutput, t.TransactionNonce, t, nil
}
