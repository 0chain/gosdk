//go:build js && wasm
// +build js,wasm

package main

import (
	"time"

	"github.com/0chain/gosdk/zboxcore/sdk"
	lru "github.com/hashicorp/golang-lru/v2"
)

type cachedAllocation struct {
	Expiration time.Time
	Allocation *sdk.Allocation
}

var (
	cachedAllocations, _ = lru.New[string, *cachedAllocation](100)
)

func getAllocation(allocationId string) (*sdk.Allocation, error) {

	it, ok := cachedAllocations.Get(allocationId)

	if ok {
		if ok && it.Expiration.After(time.Now()) {
			return it.Allocation, nil
		}
	}
	sdk.IsWasm = true
	a, err := sdk.GetAllocation(allocationId)
	if err != nil {
		return nil, err
	}

	it = &cachedAllocation{
		Allocation: a,
		Expiration: time.Now().Add(5 * time.Minute),
	}

	cachedAllocations.Add(allocationId, it)
	return it.Allocation, nil
}

// clearAllocation remove allocation from caching
func clearAllocation(allocationID string) {
	cachedAllocations.Remove(allocationID)
}

func reloadAllocation(allocationID string) (*sdk.Allocation, error) {
	a, err := sdk.GetAllocation(allocationID)
	if err != nil {
		return nil, err
	}

	it := &cachedAllocation{
		Allocation: a,
		Expiration: time.Now().Add(5 * time.Minute),
	}

	cachedAllocations.Add(allocationID, it)

	return it.Allocation, nil
}
