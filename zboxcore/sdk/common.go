package sdk

import (
	"encoding/json"
	blobbergrpc "github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc/proto"
	"sync"

	"github.com/0chain/gosdk/core/clients/blobberClient"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
)

func getObjectTreeFromBlobber(allocationID, allocationTx, remotefilepath string, blobber *blockchain.StorageNode) (fileref.RefEntity, error) {

	respRaw, err := blobberClient.GetObjectTree(blobber.Baseurl, &blobbergrpc.GetObjectTreeRequest{
		Path:       remotefilepath,
		Allocation: allocationTx,
	})
	if err != nil {
		return nil, err
	}

	var lR ReferencePathResult
	err = json.Unmarshal(respRaw, &lR)
	if err != nil {
		return nil, err
	}

	return lR.GetRefFromObjectTree(allocationID)
}

func getAllocationDataFromBlobber(blobber *blockchain.StorageNode, allocationTx string, respCh chan<- *BlobberAllocationStats, wg *sync.WaitGroup) {
	defer wg.Done()

	respRaw, err := blobberClient.GetAllocation(blobber.Baseurl, &blobbergrpc.GetAllocationRequest{Id: allocationTx})
	if err != nil {
		logger.Logger.Error("could not get allocation from blobber -" + blobber.Baseurl + " - " + err.Error())
		respCh <- &BlobberAllocationStats{}
		return
	}

	var result BlobberAllocationStats
	err = json.Unmarshal(respRaw, &result)
	if err != nil {
		return
	}

	result.BlobberID = blobber.ID
	result.BlobberURL = blobber.Baseurl
	respCh <- &result
	return
}
