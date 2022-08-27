package sdk

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/bits"
	"net/http"
	"net/url"
	"sync"
	"time"

	"errors"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
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
	listMask       uint32
	deleteMask     uint32
	connectionID   string
	consensus      Consensus
}

func (req *DeleteRequest) deleteBlobberFile(blobber *blockchain.StorageNode, blobberIdx int, deleteMutex *sync.Mutex) {
	defer req.wg.Done()

	query := &url.Values{}

	query.Add("connection_id", req.connectionID)
	query.Add("path", req.remotefilepath)

	httpreq, err := zboxutil.NewDeleteRequest(blobber.Baseurl, req.allocationTx, query)
	if err != nil {
		l.Logger.Error(blobber.Baseurl, "Error creating delete request", err)
		return
	}

	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
	_ = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("Delete : ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			req.consensus.Done()
			deleteMutex.Lock()
			req.deleteMask |= (1 << uint32(blobberIdx))
			deleteMutex.Unlock()
			l.Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " deleted.")
		} else if resp.StatusCode == http.StatusNoContent {
			req.consensus.Done()
			deleteMutex.Lock()
			req.deleteMask |= (1 << uint32(blobberIdx))
			deleteMutex.Unlock()
			l.Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " not available in blobber.")
		} else {
			resp_body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				l.Logger.Error(blobber.Baseurl, "Response: ", string(resp_body))
			}
		}
		return nil
	})
}

func (req *DeleteRequest) getObjectTreeFromBlobber(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	return getObjectTreeFromBlobber(req.ctx, req.allocationID, req.allocationTx, req.remotefilepath, blobber)
}

func (req *DeleteRequest) ProcessDelete() error {
	num := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, num)
	var deleteMutex sync.Mutex
	removedNum := 0
	req.wg = &sync.WaitGroup{}
	req.wg.Add(num)
	for i := 0; i < num; i++ {
		go func(blobberIdx int) {
			defer req.wg.Done()
			refEntity, err := req.getObjectTreeFromBlobber(req.blobbers[blobberIdx])
			if err == nil {
				req.consensus.Done()
				deleteMutex.Lock()
				req.listMask |= (1 << uint32(blobberIdx))
				objectTreeRefs[blobberIdx] = refEntity
				deleteMutex.Unlock()
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
		}(i)
	}
	req.wg.Wait()

	req.deleteMask = uint32(0)
	req.consensus.consensus = removedNum
	numDeletes := bits.OnesCount32(req.listMask)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numDeletes)

	var c, pos int
	for i := req.listMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		go req.deleteBlobberFile(req.blobbers[pos], pos, &deleteMutex)
		c++
	}
	req.wg.Wait()

	if !req.consensus.isConsensusOk() {
		return fmt.Errorf("Delete failed: Success_rate:%d, expected:%d", req.consensus.getConsensus(), req.consensus.consensusThresh)
	}

	writeMarkerMutex, err := CreateWriteMarkerMutex(client.GetClient(), req.allocationObj)
	if err != nil {
		return fmt.Errorf("Delete failed: %s", err.Error())
	}
	err = writeMarkerMutex.Lock(context.TODO(), req.connectionID)
	defer writeMarkerMutex.Unlock(context.TODO(), req.connectionID) //nolint: errcheck
	if err != nil {
		return fmt.Errorf("Delete failed: %s", err.Error())
	}

	req.consensus.consensus = removedNum
	wg := &sync.WaitGroup{}
	wg.Add(bits.OnesCount32(req.deleteMask))
	commitReqs := make([]*CommitRequest, bits.OnesCount32(req.deleteMask))
	c = 0
	for i := req.deleteMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		commitReq := &CommitRequest{}
		commitReq.allocationID = req.allocationID
		commitReq.allocationTx = req.allocationTx
		commitReq.blobber = req.blobbers[pos]
		newChange := &allocationchange.DeleteFileChange{}
		newChange.ObjectTree = objectTreeRefs[pos]
		newChange.NumBlocks = newChange.ObjectTree.GetNumBlocks()
		newChange.Operation = constants.FileOperationDelete
		newChange.Size = newChange.ObjectTree.GetSize()
		commitReq.changes = append(commitReq.changes, newChange)
		commitReq.connectionID = req.connectionID
		commitReq.wg = wg
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
		return errors.New("Delete failed: Commit consensus failed")
	}
	return nil
}
