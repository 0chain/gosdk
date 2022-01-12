package sdk

import "sync"

var blobberReadCounter *sync.Map

func getBlobberReadCtr(blobberID string) int64 {
	rctr, ok := blobberReadCounter.Load(blobberID)
	if ok {
		return rctr.(int64)
	}
	return int64(0)
}

func incBlobberReadCtr(blobberID string, numBlocks int64) {
	rctr, ok := blobberReadCounter.Load(blobberID)
	if !ok {
		rctr = int64(0)
	}
	blobberReadCounter.Store(blobberID, (rctr.(int64))+numBlocks)
}

func setBlobberReadCtr(blobberID string, ctr int64) {
	blobberReadCounter.Store(blobberID, ctr)
}

func initReadCounter() {
	blobberReadCounter = &sync.Map{}
}
