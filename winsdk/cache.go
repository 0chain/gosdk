package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/sdk"
	lru "github.com/hashicorp/golang-lru/v2"
)

type cachedAllocation struct {
	CacheExpiresAt time.Time
	Allocation     *sdk.Allocation
}

type cachedFileMeta struct {
	CacheExpiresAt time.Time
	FileMeta       *sdk.ConsolidatedFileMeta
}

var (
	cachedAllocations, _ = lru.New[string, *cachedAllocation](100)
	cachedFileMetas, _   = lru.New[string, *cachedFileMeta](1000)
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
		CacheExpiresAt: time.Now().Add(30 * time.Minute),
	}

	cachedAllocations.Add(allocationID, it)

	return it.Allocation, nil
}

func getAllocationWith(authTicket string) (*sdk.Allocation, *marker.AuthTicket, error) {

	sEnc, err := base64.StdEncoding.DecodeString(authTicket)
	if err != nil {
		return nil, nil, errors.New("Error decoding the auth ticket." + err.Error())
	}
	at := &marker.AuthTicket{}
	err = json.Unmarshal(sEnc, at)
	if err != nil {
		return nil, nil, errors.New("Error unmarshaling the auth ticket." + err.Error())
	}
	alloc, err := getAllocation(at.AllocationID)
	return alloc, at, err
}

func getFileMeta(allocationID, remotePath string) (*sdk.ConsolidatedFileMeta, error) {

	var it *cachedFileMeta
	var ok bool

	it, ok = cachedFileMetas.Get(allocationID + ":" + remotePath)

	if ok && it.CacheExpiresAt.After(time.Now()) {
		return it.FileMeta, nil
	}

	a, err := getAllocation(allocationID)
	if err != nil {
		return nil, err
	}

	f, err := a.GetFileMeta(remotePath)
	if err != nil {
		return nil, err
	}

	it = &cachedFileMeta{
		FileMeta:       f,
		CacheExpiresAt: time.Now().Add(30 * time.Minute),
	}

	cachedFileMetas.Add(allocationID, it)

	return it.FileMeta, nil
}
