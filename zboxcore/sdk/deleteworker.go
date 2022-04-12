package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/bits"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"errors"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
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

	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)

	_ = formWriter.WriteField("connection_id", req.connectionID)
	_ = formWriter.WriteField("path", req.remotefilepath)
	formWriter.Close()
	httpreq, err := zboxutil.NewDeleteRequest(blobber.Baseurl, req.allocationTx, body)
	if err != nil {
		Logger.Error(blobber.Baseurl, "Error creating delete request", err)
		return
	}
	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
	_ = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Delete : ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			req.consensus.Done()
			deleteMutex.Lock()
			req.deleteMask |= (1 << uint32(blobberIdx))
			deleteMutex.Unlock()
			Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " deleted.")
		} else if resp.StatusCode == http.StatusNoContent {
			req.consensus.Done()
			deleteMutex.Lock()
			req.deleteMask |= (1 << uint32(blobberIdx))
			deleteMutex.Unlock()
			Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " not available in blobber.")
		} else {
			resp_body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				Logger.Error(blobber.Baseurl, "Response: ", string(resp_body))
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

			Logger.Error(err.Error())

		}(i)
	}
	req.wg.Wait()

	req.deleteMask = uint32(0)
	req.consensus.consensus = float32(removedNum)
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
		return fmt.Errorf("Delete failed: Success_rate:%2f, expected:%2f", req.consensus.getConsensusRate(), req.consensus.getConsensusRequiredForOk())
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

	req.consensus.consensus = float32(removedNum)
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
				Logger.Info("Commit success", commitReq.blobber.Baseurl)
				req.consensus.Done()
			} else {
				Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !req.consensus.isConsensusOk() {
		return errors.New("Delete failed: Commit consensus failed")
	}
	return nil
}
