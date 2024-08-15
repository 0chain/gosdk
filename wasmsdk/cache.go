//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"errors"
	"time"

	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zboxcore/client"
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

// getAllocation get allocation from cache
// if not found in cache, fetch from blockchain
// and store in cache
//   - allocationId is the allocation id
func getAllocation(allocationId string) (*sdk.Allocation, error) {

	it, ok := cachedAllocations.Get(allocationId)

	if ok {
		if ok && it.Expiration.After(time.Now()) {
			return it.Allocation, nil
		}
	}
	sdk.SetWasm()
	a, err := sdk.GetAllocation(allocationId)
	if err != nil {
		return nil, err
	}
	sdk.SetShouldVerifyHash(false)
	it = &cachedAllocation{
		Allocation: a,
		Expiration: time.Now().Add(120 * time.Minute),
	}

	cachedAllocations.Add(allocationId, it)
	return it.Allocation, nil
}

// clearAllocation remove allocation from caching
func clearAllocation(allocationID string) {
	cachedAllocations.Remove(allocationID)
}

// reloadAllocation reload allocation from blockchain and update cache
//   - allocationID is the allocation id
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

func addWebWorkers(alloc *sdk.Allocation) (err error) {
	c := client.GetClient()
	if c == nil || len(c.Keys) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	respChan := make(chan error, len(alloc.Blobbers))
	respRequired := 0
	for _, blober := range alloc.Blobbers {
		weborker, workerCreated, _ := jsbridge.NewWasmWebWorker(blober.ID,
			blober.Baseurl,
			c.ClientID,
			c.ClientKey,
			c.PeerPublicKey,
			c.Keys[0].PublicKey,
			c.Keys[0].PrivateKey,
			c.Mnemonic,
			c.IsSplit) //nolint:errcheck
		if workerCreated {
			respRequired++
			go func() {
				eventChan, err := weborker.Listen(ctx)
				if err != nil {
					respChan <- err
					return
				}
				_, ok := <-eventChan
				if !ok {
					respChan <- errors.New("worker chan closed")
					return
				}
				respChan <- nil
			}()
		}
	}
	if respRequired == 0 {
		return
	}
	for {
		select {
		case <-ctx.Done():
			PrintError(ctx.Err())
			return ctx.Err()
		case err = <-respChan:
			if err != nil {
				PrintError(err)
				return
			}
			respRequired--
			if respRequired == 0 {
				close(respChan)
				return
			}
		}
	}
}
