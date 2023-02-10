//go:build js && wasm
// +build js,wasm

package main

import (
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

func getBlobberIds(blobberUrls []string) ([]string, error) {
	return sdk.GetBlobberIds(blobberUrls)
}

func getAllocationBlobbers(preferredBlobberURLs []string,
	dataShards, parityShards int, size, expiry int64,
	minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64) ([]string, error) {

	if len(preferredBlobberURLs) > 0 {
		return sdk.GetBlobberIds(preferredBlobberURLs)
	}

	return sdk.GetAllocationBlobbers(dataShards, parityShards, size, expiry, sdk.PriceRange{
		Min: uint64(minReadPrice),
		Max: uint64(maxReadPrice),
	}, sdk.PriceRange{
		Min: uint64(minWritePrice),
		Max: uint64(maxWritePrice),
	})
}

func createAllocation(datashards, parityshards int, size, expiry int64,
	minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64, lock int64, blobberIds []string) (
	*transaction.Transaction, error) {

	options := sdk.CreateAllocationOptions{
		DataShards:   datashards,
		ParityShards: parityshards,
		Size:         size,
		Expiry:       expiry,
		ReadPrice: sdk.PriceRange{
			Min: uint64(minReadPrice),
			Max: uint64(maxReadPrice),
		},
		WritePrice: sdk.PriceRange{
			Min: uint64(minWritePrice),
			Max: uint64(maxWritePrice),
		},
		Lock:       uint64(lock),
		BlobberIds: blobberIds,
	}

	_, _, txn, err := sdk.CreateAllocationWith(options)

	return txn, err
}

func listAllocations() ([]*sdk.Allocation, error) {
	return sdk.GetAllocations()
}

func transferAllocation(allocationID, newOwnerId, newOwnerPublicKey string) error {
	if allocationID == "" {
		return RequiredArg("allocationID")
	}

	if newOwnerId == "" {
		return RequiredArg("newOwnerId")
	}

	if newOwnerPublicKey == "" {
		return RequiredArg("newOwnerPublicKey")
	}

	_, _, err := sdk.CuratorTransferAllocation(allocationID, newOwnerId, newOwnerPublicKey)

	if err == nil {
		clearAllocation(allocationID)
	}

	return err
}

func freezeAllocation(allocationID string) (string, error) {

	hash, _, err := sdk.UpdateAllocation(
		0,            //size,
		0,            //int64(expiry/time.Second),
		allocationID, // allocID,
		0,            //lock,
		false,        //updateTerms,
		"",           //addBlobberId,
		"",           //removeBlobberId,
		false,        //thirdPartyExtendable,
		&sdk.FileOptionsParameters{
			ForbidUpload: sdk.FileOptionParam{true, true},
			ForbidDelete: sdk.FileOptionParam{true, true},
			ForbidUpdate: sdk.FileOptionParam{true, true},
			ForbidMove:   sdk.FileOptionParam{true, true},
			ForbidCopy:   sdk.FileOptionParam{true, true},
			ForbidRename: sdk.FileOptionParam{true, true},
		},
	)

	if err == nil {
		clearAllocation(allocationID)
	}

	return hash, err

}

func cancelAllocation(allocationID string) (string, error) {
	hash, _, err := sdk.CancelAllocation(allocationID)

	if err == nil {
		clearAllocation(allocationID)
	}

	return hash, err
}

func updateAllocation(allocationID string,
	size, expiry int64,
	lock int64,
	updateTerms bool,
	addBlobberId, removeBlobberId string) (string, error) {
	hash, _, err := sdk.UpdateAllocation(size, expiry, allocationID, uint64(lock), updateTerms, addBlobberId, removeBlobberId, false, &sdk.FileOptionsParameters{})

	if err == nil {
		clearAllocation(allocationID)
	}

	return hash, err
}
