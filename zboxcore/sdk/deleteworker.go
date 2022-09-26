package sdk

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/0chain/errors"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type DeleteRequest struct {
	allocationObj  *Allocation
	allocationID   string
	allocationTx   string
	blobbers       []*blockchain.StorageNode
	remotefilepath string
	ctx            context.Context
	wg             *sync.WaitGroup
	deleteMask     zboxutil.Uint128
	maskMu         *sync.Mutex
	connectionID   string
	consensus      Consensus
}

func (req *DeleteRequest) deleteBlobberFile(
	blobber *blockchain.StorageNode, blobberIdx int) {

	defer req.wg.Done()

	var err error

	defer func() {
		if err != nil {
			logger.Logger.Error(err)
			req.maskMu.Lock()
			req.deleteMask = req.deleteMask.And(zboxutil.NewUint128(1).Lsh(uint64(blobberIdx)).Not())
			req.maskMu.Unlock()
		}
	}()

	query := &url.Values{}

	query.Add("connection_id", req.connectionID)
	query.Add("path", req.remotefilepath)

	httpreq, err := zboxutil.NewDeleteRequest(blobber.Baseurl, req.allocationTx, query)
	if err != nil {
		l.Logger.Error(blobber.Baseurl, "Error creating delete request", err)
		return
	}

	var resp *http.Response
	for i := 0; i < 3; i++ {
		if resp != nil {
			resp.Body.Close()
		}
		ctx, cncl := context.WithTimeout(req.ctx, time.Minute)
		resp, err = zboxutil.Client.Do(httpreq.WithContext(ctx))
		cncl()

		if err != nil {
			logger.Logger.Error(blobber.Baseurl, "Delete: ", err)
			return
		}

		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			req.consensus.Done()
			l.Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " deleted.")
			return
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			logger.Logger.Error("Got too many request error")
			var r int
			r, err = zboxutil.GetRateLimitValue(resp)
			if err != nil {
				return
			}
			time.Sleep(time.Duration(r) * time.Second)
			continue
		}

		if resp.StatusCode == http.StatusNoContent {
			req.consensus.Done()
			l.Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " not available in blobber.")
			return
		}

		var respBody []byte
		respBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Logger.Error(blobber.Baseurl, "Response: ", string(respBody))
			return
		}
		err = errors.New("response_error", fmt.Sprintf("unexpected response with status code %d, message: %s",
			resp.StatusCode, string(respBody)))
		return
	}

}

func (req *DeleteRequest) getObjectTreeFromBlobber(pos uint64) (
	fRefEntity fileref.RefEntity, err error) {

	defer func() {
		if err != nil {
			req.maskMu.Lock()
			req.deleteMask = req.deleteMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
			req.maskMu.Unlock()
		}
	}()

	fRefEntity, err = getObjectTreeFromBlobber(
		req.ctx, req.allocationID, req.allocationTx,
		req.remotefilepath, req.blobbers[pos])
	return
}

func (req *DeleteRequest) ProcessDelete() (err error) {
	num := req.deleteMask.CountOnes()
	objectTreeRefs := make([]fileref.RefEntity, num)
	var deleteMutex sync.Mutex
	removedNum := 0
	req.wg = &sync.WaitGroup{}
	req.wg.Add(num)

	var pos uint64
	for i := req.deleteMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		go func(blobberIdx uint64) {
			defer req.wg.Done()
			refEntity, err := req.getObjectTreeFromBlobber(blobberIdx)
			if err == nil {
				req.consensus.Done()
				objectTreeRefs[blobberIdx] = refEntity
				return
			}
			//it was removed from the blobber
			if errors.Is(err, constants.ErrNotFound) {
				req.consensus.Done()
				deleteMutex.Lock()
				removedNum++
				deleteMutex.Unlock()
				return
			}

			l.Logger.Error(err.Error())
		}(pos)
	}
	req.wg.Wait()

	req.consensus.consensus = removedNum
	numDeletes := req.deleteMask.CountOnes()

	req.wg.Add(numDeletes)

	for i := req.deleteMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		go req.deleteBlobberFile(req.blobbers[pos], int(pos))
	}
	req.wg.Wait()

	if !req.consensus.isConsensusOk() {
		return errors.New("consensus_not_met",
			fmt.Sprintf("Consensus on delete failed. Required consensus %d got %d",
				req.consensus.consensusThresh, req.consensus.consensus))
	}

	writeMarkerMutex, err := CreateWriteMarkerMutex(client.GetClient(), req.allocationObj)
	if err != nil {
		return fmt.Errorf("Delete failed: %s", err.Error())
	}
	err = writeMarkerMutex.Lock(
		req.ctx, &req.deleteMask, req.maskMu,
		req.blobbers, &req.consensus, time.Minute, req.connectionID)

	defer writeMarkerMutex.Unlock(req.ctx, req.deleteMask, req.blobbers, time.Minute, req.connectionID) //nolint: errcheck
	if err != nil {
		return fmt.Errorf("Delete failed: %s", err.Error())
	}

	req.consensus.consensus = removedNum
	wg := &sync.WaitGroup{}
	activeBlobbers := req.deleteMask.CountOnes()
	wg.Add(activeBlobbers)
	commitReqs := make([]*CommitRequest, activeBlobbers)
	var c int
	for i := req.deleteMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		newChange := &allocationchange.DeleteFileChange{}
		newChange.ObjectTree = objectTreeRefs[pos]
		newChange.NumBlocks = newChange.ObjectTree.GetNumBlocks()
		newChange.Operation = constants.FileOperationDelete
		newChange.Size = newChange.ObjectTree.GetSize()

		commitReq := &CommitRequest{
			allocationID: req.allocationID,
			allocationTx: req.allocationTx,
			blobber:      req.blobbers[pos],
			connectionID: req.connectionID,
			wg:           wg,
		}
		commitReq.changes = append(commitReq.changes, newChange)
		commitReqs[c] = commitReq
		go AddCommitRequest(commitReq)
		c++
	}
	wg.Wait()

	for _, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				l.Logger.Info("Commit success", commitReq.blobber.Baseurl)
				req.consensus.Done()
			} else {
				l.Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			l.Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !req.consensus.isConsensusOk() {
		return errors.New("consensus_not_met",
			fmt.Sprintf("Consensus on commit not met. Required %d, got %d",
				req.consensus.consensusThresh, req.consensus.consensus))
	}
	return nil
}
