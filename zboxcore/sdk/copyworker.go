package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/errors"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"

	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type CopyRequest struct {
	allocationObj  *Allocation
	allocationID   string
	allocationTx   string
	blobbers       []*blockchain.StorageNode
	remotefilepath string
	destPath       string
	ctx            context.Context
	wg             *sync.WaitGroup
	copyMask       zboxutil.Uint128
	maskMU         *sync.Mutex
	connectionID   string
	Consensus
}

func (req *CopyRequest) getObjectTreeFromBlobber(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	return getObjectTreeFromBlobber(req.ctx, req.allocationID, req.allocationTx, req.remotefilepath, blobber)
}

func (req *CopyRequest) copyBlobberObject(
	blobber *blockchain.StorageNode, blobberIdx int) (refEntity fileref.RefEntity, err error) {

	defer req.wg.Done()

	defer func() {
		if err != nil {
			req.maskMU.Lock()
			// Removing blobber from mask
			req.copyMask = req.copyMask.And(zboxutil.NewUint128(1).Lsh(uint64(blobberIdx)).Not())
			req.maskMU.Unlock()
		}
	}()
	refEntity, err = req.getObjectTreeFromBlobber(req.blobbers[blobberIdx])
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	for i := 0; i < 3; i++ {
		if resp != nil {
			resp.Body.Close()
		}

		body := new(bytes.Buffer)
		formWriter := multipart.NewWriter(body)

		formWriter.WriteField("connection_id", req.connectionID)
		formWriter.WriteField("path", req.remotefilepath)
		formWriter.WriteField("dest", req.destPath)
		formWriter.Close()

		var httpreq *http.Request
		httpreq, err = zboxutil.NewCopyRequest(blobber.Baseurl, req.allocationTx, body)
		if err != nil {
			l.Logger.Error(blobber.Baseurl, "Error creating rename request", err)
			return nil, err
		}
		httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
		l.Logger.Info(httpreq.URL.Path)
		ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
		resp, err = zboxutil.Client.Do(httpreq.WithContext(ctx))
		cncl()

		if err != nil {
			logger.Logger.Error("Copy: ", err)
			return nil, err
		}
		defer resp.Body.Close()

		var respBody []byte
		respBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Logger.Error("Error: Resp ", err)
			return nil, err
		}

		if resp.StatusCode == http.StatusOK {
			l.Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " copied.")
			req.Consensus.Done()
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
		l.Logger.Error(blobber.Baseurl, "Response: ", string(respBody))
		err = errors.New("response_error", string(respBody))

	}
	return
}

func (req *CopyRequest) ProcessCopy() error {
	numList := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, numList)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)

	var pos uint64
	for i := req.copyMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		go func(blobberIdx int) {
			refEntity, err := req.copyBlobberObject(req.blobbers[blobberIdx], blobberIdx)
			if err != nil {
				l.Logger.Error(err.Error())
				return
			}
			objectTreeRefs[blobberIdx] = refEntity
		}(int(pos))
	}

	req.wg.Wait()

	if !req.isConsensusOk() {
		return errors.New("consensus_not_met",
			fmt.Sprintf("Copy failed. Required consensus %d, got %d",
				req.Consensus.consensusThresh, req.Consensus.consensus))
	}

	writeMarkerMutex, err := CreateWriteMarkerMutex(client.GetClient(), req.allocationObj)
	if err != nil {
		return fmt.Errorf("Copy failed: %s", err.Error())
	}
	err = writeMarkerMutex.Lock(req.ctx, &req.copyMask, req.maskMU, req.blobbers, &req.Consensus, time.Minute, req.connectionID)
	defer writeMarkerMutex.Unlock(req.ctx, req.copyMask, req.blobbers, time.Minute, req.connectionID) //nolint: errcheck
	if err != nil {
		return fmt.Errorf("Copy failed: %s", err.Error())
	}

	req.Consensus.Reset()
	wg := &sync.WaitGroup{}
	activeBlobbers := req.copyMask.CountOnes()
	wg.Add(activeBlobbers)
	commitReqs := make([]*CommitRequest, activeBlobbers)

	var c int
	for i := req.copyMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		newChange := &allocationchange.CopyFileChange{
			DestPath:   req.destPath,
			ObjectTree: objectTreeRefs[pos],
		}
		newChange.NumBlocks = 0
		newChange.Operation = constants.FileOperationCopy
		newChange.Size = 0
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
				req.consensus++
			} else {
				l.Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			l.Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !req.isConsensusOk() {
		return errors.New("consensus_not_met",
			fmt.Sprintf("Commit on copy failed. Required consensus %d, got %d",
				req.Consensus.consensusThresh, req.Consensus.consensus))
	}
	return nil
}
