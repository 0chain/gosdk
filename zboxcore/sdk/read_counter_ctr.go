package sdk

import (
	"sync"
)

type blobberReadCounter struct {
	m  map[string]int64
	mu *sync.RWMutex
}

var brc *blobberReadCounter

func InitReadCounter() {
	if brc == nil {
		brc = &blobberReadCounter{
			m:  make(map[string]int64),
			mu: &sync.RWMutex{},
		}
	}

}

func setBlobberReadCtr(clientID, blobberID string, ctr int64) {
	key := clientID + blobberID
	brc.mu.Lock()
	brc.m[key] = ctr
	brc.mu.Unlock()
}

func getBlobberReadCtr(clientID, blobberID string) int64 {
	key := clientID + blobberID
	brc.mu.RLock()
	c := brc.m[key]
	brc.mu.RUnlock()
	return c
}

func incBlobberReadCtr(clientID, blobberID string, numBlocks int64) {
	key := clientID + blobberID
	brc.mu.Lock()
	brc.m[key] += numBlocks
	brc.mu.Unlock()
}
