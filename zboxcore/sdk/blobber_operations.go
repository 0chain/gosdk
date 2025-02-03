package sdk

import (
	"encoding/json"
	"math"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk_common/core/client"
	"github.com/0chain/gosdk_common/core/transaction"
)

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
	size, authRoundExpiry int64,
	extend bool,
	allocationID string,
	lock uint64,
	addBlobberId, addBlobberAuthTicket, removeBlobberId, ownerID, ownerSigninPublicKey string,
	setThirdPartyExtendable bool, fileOptionsParams *FileOptionsParameters, ticket string,
) (hash string, nonce int64, err error) {
	if ownerID == "" {
		ownerID = client.Id()
	}

	if lock > math.MaxInt64 {
		return "", 0, errors.New("invalid_lock", "int64 overflow on lock value")
	}

	if !client.IsSDKInitialized() {
		return "", 0, sdkNotInitialized
	}

	alloc, err := GetAllocationForUpdate(allocationID)
	if err != nil {
		return "", 0, allocationNotFound
	}

	updateAllocationRequest := make(map[string]interface{})
	updateAllocationRequest["owner_id"] = ownerID
	updateAllocationRequest["owner_public_key"] = ""
	updateAllocationRequest["id"] = allocationID
	updateAllocationRequest["size"] = size
	updateAllocationRequest["extend"] = extend
	updateAllocationRequest["add_blobber_id"] = addBlobberId
	updateAllocationRequest["add_blobber_auth_ticket"] = addBlobberAuthTicket
	updateAllocationRequest["remove_blobber_id"] = removeBlobberId
	updateAllocationRequest["set_third_party_extendable"] = setThirdPartyExtendable
	updateAllocationRequest["owner_signing_public_key"] = ownerSigninPublicKey
	updateAllocationRequest["file_options_changed"], updateAllocationRequest["file_options"] = calculateAllocationFileOptions(alloc.FileOptions, fileOptionsParams)
	updateAllocationRequest["auth_round_expiry"] = authRoundExpiry

	if ticket != "" {

		type Ticket struct {
			AllocationID  string `json:"allocation_id"`
			UserID        string `json:"user_id"`
			RoundExpiry   int64  `json:"round_expiry"`
			OperationType string `json:"operation_type"`
			Signature     string `json:"signature"`
		}

		ticketData := &Ticket{}
		err := json.Unmarshal([]byte(ticket), ticketData)
		if err != nil {
			return "", 0, errors.New("invalid_ticket", "invalid ticket")
		}
		updateAllocationRequest["update_ticket"] = ticketData
	}

	sn := transaction.SmartContractTxnData{
		Name:      transaction.STORAGESC_UPDATE_ALLOCATION,
		InputArgs: updateAllocationRequest,
	}
	hash, _, nonce, _, err = storageSmartContractTxnValue(sn, lock)
	return
}
