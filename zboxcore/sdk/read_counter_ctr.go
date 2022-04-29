package sdk

import (
	"sync"

	"github.com/0chain/gosdk/zboxcore/blockchain"
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

func setBlobberReadCtr(blobber *blockchain.StorageNode, ctr int64) {
	brc.mu.Lock()
	defer brc.mu.Unlock()
	brc.m[blobber.ID] = ctr
}

func getBlobberReadCtr(blobber *blockchain.StorageNode) int64 {
	return brc.m[blobber.ID]
}

func incBlobberReadCtr(blobber *blockchain.StorageNode, numBlocks int64) {
	brc.mu.Lock()
	defer brc.mu.Unlock()
	brc.m[blobber.ID] += numBlocks
}
