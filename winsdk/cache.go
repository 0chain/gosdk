package main

import (
	"time"

	"github.com/0chain/gosdk/zboxcore/sdk"
	lru "github.com/hashicorp/golang-lru/v2"
)

type cachedAllocation struct {
	CacheExpiresAt time.Time
	Allocation     *sdk.Allocation
}

var (
	cachedAllocations, _ = lru.New[string, *cachedAllocation](100)
)

func getAllocation(allocationID string) (*sdk.Allocation, error) {

	var it *cachedAllocation
	var ok bool

	it, ok = cachedAllocations.Get(allocationID)

	if ok && it.CacheExpiresAt.After(time.Now()) {
		return it.Allocation, nil
	}

	a, err := sdk.GetAllocation(allocationID)
	if err != nil {
		return nil, err
	}

	it = &cachedAllocation{
		Allocation:     a,
		CacheExpiresAt: time.Now().Add(5 * time.Minute),
	}

	cachedAllocations.Add(allocationID, it)

	return it.Allocation, nil
}
