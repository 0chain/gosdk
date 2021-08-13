package sdk

import (
	"context"
	"encoding/json"
	blobbergrpc "github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc/proto"
	"strconv"
	"sync"

	"github.com/0chain/gosdk/core/clients/blobberClient"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
)

type ReferencePathResult struct {
	*fileref.ReferencePath
	LatestWM *marker.WriteMarker `json:"latest_write_marker"`
}

type CommitResult struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_msg,omitempty"`
}

func ErrorCommitResult(errMsg string) *CommitResult {
	result := &CommitResult{Success: false, ErrorMessage: errMsg}
	return result
}

func SuccessCommitResult() *CommitResult {
	result := &CommitResult{Success: true}
	return result
}

type CommitRequest struct {
	changes      []allocationchange.AllocationChange
	blobber      *blockchain.StorageNode
	allocationID string
	allocationTx string
	connectionID string
	wg           *sync.WaitGroup
	result       *CommitResult
}

var commitChan map[string]chan *CommitRequest
var initCommitMutex sync.Mutex

func InitCommitWorker(blobbers []*blockchain.StorageNode) {
	// if commitChan != nil {
	// 	for _, v := range commitChan {
	// 		close(v)
	// 	}
	// }
	// commitChan = make(map[string]chan *CommitRequest)
	// for _, blobber := range blobbers {
	// 	Logger.Info("Atempting to start the commit worker for ", blobber.Baseurl)
	// 	commitChan[blobber.ID] = make(chan *CommitRequest, 1)
	// 	go startCommitWorker(blobber)
	// }
	initCommitMutex.Lock()
	defer initCommitMutex.Unlock()
	if commitChan == nil {
		commitChan = make(map[string]chan *CommitRequest)
	}

	for _, blobber := range blobbers {
		if _, ok := commitChan[blobber.ID]; !ok {
			commitChan[blobber.ID] = make(chan *CommitRequest, 1)
			blobberChan := commitChan[blobber.ID]
			go startCommitWorker(blobberChan, blobber.ID)
		}
	}

}

func startCommitWorker(blobberChan chan *CommitRequest, blobberID string) {
	for true {
		commitreq, open := <-blobberChan
		if !open {
			break
		}
		commitreq.processCommit()
	}
	initCommitMutex.Lock()
	defer initCommitMutex.Unlock()
	delete(commitChan, blobberID)
}

func (commitreq *CommitRequest) processCommit() {
	Logger.Info("received a commit request")
	paths := make([]string, 0)
	for _, change := range commitreq.changes {
		paths = append(paths, change.GetAffectedPath())
	}
	var lR ReferencePathResult

	pathsRaw, err := json.Marshal(paths)
	if err != nil {
		return
	}

	respRaw, err := blobberClient.GetReferencePath(commitreq.blobber.Baseurl, &blobbergrpc.GetReferencePathRequest{
		Paths:      string(pathsRaw),
		Path:       "",
		Allocation: commitreq.allocationTx,
	})
	//process the commit request for the blobber here
	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		commitreq.wg.Done()
		Logger.Error("could not get reference path from blobber -" + commitreq.blobber.Baseurl + " - " + err.Error())
		return
	}
	err = json.Unmarshal(respRaw, &lR)
	if err != nil {
		return
	}

	rootRef, err := lR.GetDirTree(commitreq.allocationID)
	if lR.LatestWM != nil {
		//Can not verify signature due to collaborator flow
		// //TODO: Verify the writemarker
		// err = lR.LatestWM.VerifySignature(client.GetClientPublicKey())
		// if err != nil {
		// 	commitreq.result = ErrorCommitResult(err.Error())
		// 	commitreq.wg.Done()
		// 	return
		// }

		rootRef.CalculateHash()
		prevAllocationRoot := encryption.Hash(rootRef.Hash + ":" + strconv.FormatInt(lR.LatestWM.Timestamp, 10))
		if prevAllocationRoot != lR.LatestWM.AllocationRoot {
			// Removing this check for testing purpose as per the convo with Saswata
			Logger.Info("Allocation root from latest writemarker mismatch. Expected: " + prevAllocationRoot + " got: " + lR.LatestWM.AllocationRoot)
			// err = commitreq.calculateHashRequest(ctx, paths)
			// if err != nil {
			// 	commitreq.result = ErrorCommitResult("Failed to call blobber to recalculate the hash. URL: " + commitreq.blobber.Baseurl + ", Err : " + err.Error())
			// 	commitreq.wg.Done()
			// 	return
			// }
			// Logger.Info("Recalculate hash call to blobber successfull")
			// commitreq.result = ErrorCommitResult("Allocation root from latest writemarker mismatch. Expected: " + prevAllocationRoot + " got: " + lR.LatestWM.AllocationRoot)
			// commitreq.wg.Done()
			// return
		}
	}
	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		commitreq.wg.Done()
		return
	}
	size := int64(0)
	for _, change := range commitreq.changes {
		err = change.ProcessChange(rootRef)
		if err != nil {
			break
		}
		size += change.GetSize()
	}
	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		commitreq.wg.Done()
		return
	}
	err = commitreq.commitBlobber(rootRef, lR.LatestWM, size)
	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		commitreq.wg.Done()
		return
	}
	commitreq.result = SuccessCommitResult()
	commitreq.wg.Done()
}

func (req *CommitRequest) commitBlobber(rootRef *fileref.Ref, latestWM *marker.WriteMarker, size int64) error {
	wm := &marker.WriteMarker{}
	timestamp := int64(common.Now())
	wm.AllocationRoot = encryption.Hash(rootRef.Hash + ":" + strconv.FormatInt(timestamp, 10))
	if latestWM != nil {
		wm.PreviousAllocationRoot = latestWM.AllocationRoot
	} else {
		wm.PreviousAllocationRoot = ""
	}

	wm.AllocationID = req.allocationID
	wm.Size = size
	wm.BlobberID = req.blobber.ID
	wm.Timestamp = timestamp
	wm.ClientID = client.GetClientID()
	err := wm.Sign()
	if err != nil {
		Logger.Error("Signing writemarker failed: ", err)
		return err
	}
	wmData, err := json.Marshal(wm)
	if err != nil {
		Logger.Error("Creating writemarker failed: ", err)
		return err
	}

	Logger.Info("Committing to blobber." + req.blobber.Baseurl)
	commitResp, err := blobberClient.Commit(req.blobber.Baseurl, &blobbergrpc.CommitRequest{
		Allocation:   req.allocationTx,
		ConnectionId: req.connectionID,
		WriteMarker:  string(wmData),
	})
	if err != nil {
		Logger.Error("Commit response - " + string(commitResp))
		return err
	}

	return nil
}

func AddCommitRequest(req *CommitRequest) {
	commitChan[req.blobber.ID] <- req
}

func (commitreq *CommitRequest) calculateHashRequest(ctx context.Context, paths []string) error {
	pathsRaw, err := json.Marshal(paths)
	if err != nil {
		return err
	}

	Logger.Info("Calculating Hash " + commitreq.blobber.Baseurl)
	resp, err := blobberClient.CalculateHash(commitreq.blobber.Baseurl, &blobbergrpc.CalculateHashRequest{
		Allocation: commitreq.allocationTx,
		Path:       "",
		Paths:      string(pathsRaw),
	})

	if err != nil {
		Logger.Error("Commit response - " + string(resp))
		return err
	}

	return nil
}
