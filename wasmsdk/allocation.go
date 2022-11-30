//go:build js && wasm
// +build js,wasm

package main

import (
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/client"
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

func createAllocation(name string, datashards, parityshards int, size, expiry int64,
	minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64, lock int64, blobberIds []string) (
	*transaction.Transaction, error) {

	options := sdk.CreateAllocationOptions{
		Name:         name,
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

func transferAllocation(allocationId, newOwnerId, newOwnerPublicKey string) error {
	if allocationId == "" {
		return RequiredArg("allocationId")
	}

	if newOwnerId == "" {
		return RequiredArg("newOwnerId")
	}

	if newOwnerPublicKey == "" {
		return RequiredArg("newOwnerPublicKey")
	}

	_, _, err := sdk.CuratorTransferAllocation(allocationId, newOwnerId, newOwnerPublicKey)

	return err
}

func freezeAllocation(allocationId string) error {

	_, _, err := sdk.UpdateAllocation(
		"",           //allocationName,
		0,            //size,
		0,            //int64(expiry/time.Second),
		allocationId, // allocID,
		0,            //lock,
		true,         // setImmutable,
		false,        //updateTerms,
		"",           //addBlobberId,
		"",           //removeBlobberId,
	)

	return err

}
