package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/bits"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"errors"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type DeleteRequest struct {
	allocationID   string
	allocationTx   string
	blobbers       []*blockchain.StorageNode
	remotefilepath string
	ctx            context.Context
	wg             *sync.WaitGroup
	listMask       uint32
	deleteMask     uint32
	connectionID   string
	Consensus
}

func (req *DeleteRequest) deleteBlobberFile(blobber *blockchain.StorageNode, blobberIdx int) {
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
			req.consensus++
			req.deleteMask |= (1 << uint32(blobberIdx))
			Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " deleted.")
		} else if resp.StatusCode == http.StatusNoContent {
			req.consensus++
			req.deleteMask |= (1 << uint32(blobberIdx))
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
	numList := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, numList)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	totalRefFound := 0
	for i := 0; i < numList; i++ {
		go func(blobberIdx int) {
			defer req.wg.Done()
			refEntity, err := req.getObjectTreeFromBlobber(req.blobbers[blobberIdx])
			if err != nil && !strings.Contains(err.Error(), "Invalid path") {
				Logger.Error(err.Error())
				return
			}
			req.consensus++
			req.listMask |= (1 << uint32(blobberIdx))
			if refEntity != nil {
				totalRefFound++
			}
			objectTreeRefs[blobberIdx] = refEntity
		}(i)
	}
	req.wg.Wait()

	if !req.isConsensusOk() {
		return fmt.Errorf("Delete failed: Success_rate:%2f, expected:%2f", req.getConsensusRate(), req.getConsensusRequiredForOk())
	}
	if totalRefFound == 0 {
		return fmt.Errorf("Delete failed: Invalid reference %s", req.remotefilepath)
	}

	initConsensus := numList - totalRefFound
	req.deleteMask = uint32(0)
	req.consensus = float32(initConsensus)
	req.wg = &sync.WaitGroup{}

	var pos int
	for i := req.listMask; i != 0; i &= ^(1 << uint32(pos)) {
		pos = bits.TrailingZeros32(i)
		if objectTreeRefs[pos] != nil {
			req.wg.Add(1)
			go req.deleteBlobberFile(req.blobbers[pos], pos)
		} else {
			req.consensus++
		}
	}
	req.wg.Wait()

	if !req.isConsensusOk() {
		return fmt.Errorf("Delete failed: Success_rate:%2f, expected:%2f", req.getConsensusRate(), req.getConsensusRequiredForOk())
	}

	req.consensus = float32(initConsensus)
	wg := &sync.WaitGroup{}
	commitReqs := make([]*CommitRequest, bits.OnesCount32(req.deleteMask))
	var c int
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
		wg.Add(1)
		go AddCommitRequest(commitReq)
		c++
	}
	wg.Wait()

	for _, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				Logger.Info("Commit success", commitReq.blobber.Baseurl)
				req.consensus++
			} else {
				Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !req.isConsensusOk() {
		return errors.New("Delete failed: Commit consensus failed")
	}
	return nil
}
