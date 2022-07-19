//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"

	"github.com/0chain/gosdk/wasmsdk/zbox"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

var (
	zboxHost   string
	zboxClient *zbox.Client
)

func SetZBoxHost(host string) {
	zboxHost = host

	c := client.GetClient()

	if len(c.ClientID) > 0 {
		zboxClient = zbox.NewClient(zboxHost, c.ClientID, c.ClientKey)
	}
}

func CreateFreeAllocation() (string, error) {
	return zboxClient.GetFreeMarker()
}

func CreateAllocation(name string, datashards, parityshards int, size, expiry int64,
	minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64, lock int64) (
	string, error) {

	fmt.Println("createAllocation:", name, datashards, parityshards, size, expiry, minReadPrice, maxReadPrice, minWritePrice, maxWritePrice, lock)

	hash, _, err := sdk.CreateAllocation(name, datashards, parityshards, size, expiry, sdk.PriceRange{
		Min: uint64(minReadPrice),
		Max: uint64(maxReadPrice),
	}, sdk.PriceRange{
		Min: uint64(minWritePrice),
		Max: uint64(maxWritePrice),
	}, uint64(lock))

	fmt.Println("createAllocation:", hash, err)

	return hash, err
}
