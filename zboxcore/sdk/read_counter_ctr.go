package sdk

import (
	"sync"
)

type blobberReadCounter struct {
	m  map[string]int64
	mu *sync.Mutex
}

var brc blobberReadCounter

func InitReadCounter() {
	brc = blobberReadCounter{
		m:  make(map[string]int64),
		mu: &sync.Mutex{},
	}

}

func setBlobberReadCtr(blobberID string, ctr int64) {
	brc.mu.Lock()
	defer brc.mu.Unlock()
	brc.m[blobberID] = ctr
}

func getBlobberReadCtr(blobberID string) int64 {
	return brc.m[blobberID]
}

func incBlobberReadCtr(blobberID string, numBlocks int64) {
	brc.mu.Lock()
	defer brc.mu.Unlock()
	brc.m[blobberID] += numBlocks
}
