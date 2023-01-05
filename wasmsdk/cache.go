//go:build js && wasm
// +build js,wasm

package main

import (
	"sync"
	"time"

	"github.com/0chain/gosdk/zboxcore/sdk"
)

type cachedAllocation struct {
	Expiration time.Time
	Allocation *sdk.Allocation
}

var (
	cachedAllocations      = make(map[string]*cachedAllocation)
	cachedAllocationsMutex sync.Mutex
)

func getAllocation(allocationID string) (*sdk.Allocation, error) {
	cachedAllocationsMutex.Lock()
	defer cachedAllocationsMutex.Unlock()

	it, ok := cachedAllocations[allocationID]

	if !ok || it.Expiration.Before(time.Now()) {

		a, err := sdk.GetAllocation(allocationID)
		if err != nil {
			return nil, err
		}

		it = &cachedAllocation{
			Allocation: a,
			Expiration: time.Now().Add(5 * time.Minute),
		}

		cachedAllocations[allocationID] = it

	}

	return it.Allocation, nil
}
