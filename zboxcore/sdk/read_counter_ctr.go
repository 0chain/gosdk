package sdk

import (
	"sync"
)

type blobberReadCounter struct {
	m  map[string]int64
	mu *sync.RWMutex
}

var brc = &blobberReadCounter{
	m:  make(map[string]int64),
	mu: &sync.RWMutex{},
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
