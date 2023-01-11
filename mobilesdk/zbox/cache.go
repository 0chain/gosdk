package zbox

import (
	"sync"
	"time"

	"github.com/0chain/gosdk/zboxcore/sdk"
	lru "github.com/hashicorp/golang-lru/v2"
)

type cachedAllocation struct {
	Expiration time.Time
	Allocation *sdk.Allocation
}

var (
	cachedAllocations, _   = lru.New[string, *cachedAllocation](100)
	cachedAllocationsMutex sync.Mutex
)

func getAllocation(allocationID string) (*sdk.Allocation, error) {
	cachedAllocationsMutex.Lock()
	defer cachedAllocationsMutex.Unlock()

	it, ok := cachedAllocations.Get(allocationID)

	if ok {
		if ok && it.Expiration.After(time.Now()) {
			return it.Allocation, nil
		}
	}

	a, err := sdk.GetAllocation(allocationID)
	if err != nil {
		return nil, err
	}

	it = &cachedAllocation{
		Allocation: a,
		Expiration: time.Now().Add(5 * time.Minute),
	}

	cachedAllocations.Add(allocationID, it)

	return it.Allocation, nil
}
