//go:build js && wasm
// +build js,wasm

package main

import (
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/wasmsdk/zbox"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

var (
	zboxHost   string
	zboxClient *zbox.Client
)

func setZBoxHost(host string) {
	zboxHost = host

	c := client.GetClient()

	if len(c.ClientID) > 0 {
		zboxClient = zbox.NewClient(zboxHost, c.ClientID, c.ClientKey)
	}
}

func CreateFreeAllocation() (string, error) {
	return zboxClient.GetFreeMarker()
}

func getAllocationBlobbers(preferredBlobberURLs []string,
	dataShards, parityShards int, size, expiry int64,
	minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64) ([]string, error) {

	if len(preferredBlobberURLs) > 0 {
		return sdk.GetBlobberIds(preferredBlobberURLs)
	}

	c := client.GetClient()

	return sdk.GetAllocationBlobbers(c.ClientID, c.ClientKey, dataShards, parityShards, size, expiry, sdk.PriceRange{
		Min: uint64(minReadPrice),
		Max: uint64(maxReadPrice),
	}, sdk.PriceRange{
		Min: uint64(minWritePrice),
		Max: uint64(maxWritePrice),
	})
}

func createAllocation(name string, datashards, parityshards int, size, expiry int64,
	minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64, lock int64) (
	*transaction.Transaction, error) {

	_, _, txn, err := sdk.CreateAllocation(name, datashards, parityshards, size, expiry, sdk.PriceRange{
		Min: uint64(minReadPrice),
		Max: uint64(maxReadPrice),
	}, sdk.PriceRange{
		Min: uint64(minWritePrice),
		Max: uint64(maxWritePrice),
	}, uint64(lock))

	return txn, err
}
