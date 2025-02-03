package sdk

import (
	"math"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk_common/core/client"
	"github.com/0chain/gosdk_common/zboxcore/commonsdk"
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
	setThirdPartyExtendable bool, fileOptionsParams *commonsdk.FileOptionsParameters, ticket string,
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

	hash, nonce, err = commonsdk.UpdateAllocationWithRequest(
		commonsdk.UpdateAllocationOptions{
			Size:                    size,
			Extend:                  extend,
			AllocationID:            allocationID,
			Lock:                    lock,
			AddBlobberID:            addBlobberId,
			AddBlobberAuthTicket:    addBlobberAuthTicket,
			RemoveBlobberID:         removeBlobberId,
			SetThirdPartyExtendable: setThirdPartyExtendable,
			FileOptionsParams:       fileOptionsParams,
			OwnerID:                 ownerID,
			OwnerSigninPublicKey:    ownerSigninPublicKey,
			Ticket:                  ticket,
			AuthRoundExpiry:         authRoundExpiry,
			FileOptions:             alloc.FileOptions,
		},
	)
	return
}
