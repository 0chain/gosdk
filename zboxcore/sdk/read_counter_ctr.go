package sdk

import (
	"sync"

	"github.com/0chain/gosdk/zboxcore/blockchain"
)

type blobberReadCounter struct {
	m  map[string]int64
	mu *sync.RWMutex
}

var brc blobberReadCounter

func InitReadCounter() {
	brc = blobberReadCounter{
		m:  make(map[string]int64),
		mu: &sync.RWMutex{},
	}

}

func setBlobberReadCtr(blobber *blockchain.StorageNode, ctr int64) {
	brc.mu.Lock()
	brc.m[blobber.ID] = ctr
	brc.mu.Unlock()
}

func getBlobberReadCtr(blobber *blockchain.StorageNode) int64 {
	brc.mu.RLock()
	c := brc.m[blobber.ID]
	brc.mu.RUnlock()
	return c
}

func incBlobberReadCtr(blobber *blockchain.StorageNode, numBlocks int64) {
	brc.mu.Lock()
	brc.m[blobber.ID] += numBlocks
	brc.mu.Unlock()
}
