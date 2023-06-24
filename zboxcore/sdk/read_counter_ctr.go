package sdk

import (
	"sync"
)

type blobberReadCounter struct {
	m              map[string]int64
	mu             *sync.RWMutex
	blobberLockMap map[string]*sync.Mutex
	muBlobberLock  *sync.Mutex
}

var brc = &blobberReadCounter{
	m:              make(map[string]int64),
	mu:             &sync.RWMutex{},
	blobberLockMap: make(map[string]*sync.Mutex),
	muBlobberLock:  &sync.Mutex{},
}

func lockBlobberReadCtr(allocID, blobberID string) {
	brc.muBlobberLock.Lock()
	key := allocID + blobberID
	if _, ok := brc.blobberLockMap[key]; !ok {
		brc.blobberLockMap[key] = &sync.Mutex{}
	}
	mut := brc.blobberLockMap[key]
	brc.muBlobberLock.Unlock()
	mut.Lock()
}

func unlockBlobberReadCtr(allocID, blobberID string) {
	brc.muBlobberLock.Lock()
	key := allocID + blobberID
	mut := brc.blobberLockMap[key]
	brc.muBlobberLock.Unlock()
	mut.Unlock()
}

func setBlobberReadCtr(allocID, blobberID string, ctr int64) {
	key := allocID + blobberID
	brc.mu.Lock()
	brc.m[key] = ctr
	brc.mu.Unlock()
}

func getBlobberReadCtr(allocID, blobberID string) int64 {
	key := allocID + blobberID
	brc.mu.RLock()
	c := brc.m[key]
	brc.mu.RUnlock()
	return c
}

func incBlobberReadCtr(allocID, blobberID string, numBlocks int64) {
	key := allocID + blobberID
	brc.mu.Lock()
	brc.m[key] += numBlocks
	brc.mu.Unlock()
}
