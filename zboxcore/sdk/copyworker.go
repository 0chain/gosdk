package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/google/uuid"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/util"
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
	ctxCncl        context.CancelFunc
	copyMask       zboxutil.Uint128
	maskMU         *sync.Mutex
	connectionID   string
	timestamp      int64
	Consensus
}

func (req *CopyRequest) getObjectTreeFromBlobber(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	return getObjectTreeFromBlobber(req.ctx, req.allocationID, req.allocationTx, req.remotefilepath, blobber)
}

func (req *CopyRequest) copyBlobberObject(
	blobber *blockchain.StorageNode, blobberIdx int) (refEntity fileref.RefEntity, err error) {

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
	var shouldContinue bool
	var latestRespMsg string
	var latestStatusCode int
	for i := 0; i < 3; i++ {
		err, shouldContinue = func() (err error, shouldContinue bool) {
			body := new(bytes.Buffer)
			formWriter := multipart.NewWriter(body)

			err = formWriter.WriteField("connection_id", req.connectionID)
			if err != nil {
				return err, false
			}

			err = formWriter.WriteField("path", req.remotefilepath)
			if err != nil {
				return err, false
			}

			err = formWriter.WriteField("dest", req.destPath)
			if err != nil {
				return err, false
			}

			err = formWriter.Close()
			if err != nil {
				return err, false
			}

			var (
				httpreq  *http.Request
				respBody []byte
				ctx      context.Context
				cncl     context.CancelFunc
			)

			httpreq, err = zboxutil.NewCopyRequest(blobber.Baseurl, req.allocationID, req.allocationTx, body)
			if err != nil {
				l.Logger.Error(blobber.Baseurl, "Error creating rename request", err)
				return
			}

			httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
			l.Logger.Info(httpreq.URL.Path)
			ctx, cncl = context.WithTimeout(req.ctx, DefaultUploadTimeOut)
			resp, err = zboxutil.Client.Do(httpreq.WithContext(ctx))
			defer cncl()

			if err != nil {
				logger.Logger.Error("Copy: ", err)
				return
			}

			if resp.Body != nil {
				defer resp.Body.Close()
			}
			respBody, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Logger.Error("Error: Resp ", err)
				return
			}

			if resp.StatusCode == http.StatusOK {
				l.Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " copied.")
				req.Consensus.Done()
				return
			}

			latestRespMsg = string(respBody)
			latestStatusCode = resp.StatusCode

			if resp.StatusCode == http.StatusTooManyRequests {
				logger.Logger.Error("Got too many request error")
				var r int
				r, err = zboxutil.GetRateLimitValue(resp)
				if err != nil {
					logger.Logger.Error(err)
					return
				}
				time.Sleep(time.Duration(r) * time.Second)
				shouldContinue = true
				return
			}
			l.Logger.Error(blobber.Baseurl, "Response: ", string(respBody))
			err = errors.New("response_error", string(respBody))
			return
		}()

		if err != nil {
			return
		}
		if shouldContinue {
			continue
		}
		return
	}
	return nil, errors.New("unknown_issue",
		fmt.Sprintf("last status code: %d, last response message: %s", latestStatusCode, latestRespMsg))
}

func (req *CopyRequest) ProcessWithBlobbers() ([]fileref.RefEntity, []error) {
	var pos uint64
	numList := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, numList)
	blobberErrors := make([]error, numList)

	wg := &sync.WaitGroup{}
	for i := req.copyMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		wg.Add(1)
		go func(blobberIdx int) {
			defer wg.Done()
			refEntity, err := req.copyBlobberObject(req.blobbers[blobberIdx], blobberIdx)
			if err != nil {
				blobberErrors[blobberIdx] = err
				l.Logger.Error(err.Error())
				return
			}
			objectTreeRefs[blobberIdx] = refEntity
		}(int(pos))
	}
	wg.Wait()
	return objectTreeRefs, blobberErrors
}

func (req *CopyRequest) ProcessCopy() error {
	defer req.ctxCncl()

	wg := &sync.WaitGroup{}
	var pos uint64

	objectTreeRefs, blobberErrors := req.ProcessWithBlobbers()

	if !req.isConsensusOk() {
		err := zboxutil.MajorError(blobberErrors)
		if err != nil {
			return errors.New("copy_failed", fmt.Sprintf("Copy failed. %s", err.Error()))
		}

		return errors.New("consensus_not_met",
			fmt.Sprintf("Copy failed. Required consensus %d, got %d",
				req.Consensus.consensusThresh, req.Consensus.consensus))
	}

	writeMarkerMutex, err := CreateWriteMarkerMutex(client.GetClient(), req.allocationObj)
	if err != nil {
		return fmt.Errorf("Copy failed: %s", err.Error())
	}
	err = writeMarkerMutex.Lock(req.ctx, &req.copyMask, req.maskMU,
		req.blobbers, &req.Consensus, 0, time.Minute, req.connectionID)
	if err != nil {
		return fmt.Errorf("Copy failed: %s", err.Error())
	}
	defer writeMarkerMutex.Unlock(req.ctx, req.copyMask, req.blobbers, time.Minute, req.connectionID) //nolint: errcheck

	//Check if the allocation is to be repaired or rolled back
	status, err := req.allocationObj.CheckAllocStatus()
	if err != nil {
		logger.Logger.Error("Error checking allocation status: ", err)
		return fmt.Errorf("Copy failed: %s", err.Error())
	}

	if status == Repair {
		logger.Logger.Info("Repairing allocation")
		// // TODO: Need status callback to call repair allocation
		// err = req.allocationObj.RepairAlloc()
		// if err != nil {
		// 	return err
		// }
	}
	if status != Commit {
		return ErrRetryOperation
	}

	req.Consensus.Reset()
	activeBlobbers := req.copyMask.CountOnes()
	wg.Add(activeBlobbers)
	commitReqs := make([]*CommitRequest, activeBlobbers)
	req.timestamp = int64(common.Now())
	uid := util.GetNewUUID()
	var c int
	for i := req.copyMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		newChange := &allocationchange.CopyFileChange{
			DestPath:   req.destPath,
			Uuid:       uid,
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
			timestamp:    req.timestamp,
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

type CopyOperation struct {
	remotefilepath string
	destPath       string
	ctx            context.Context
	ctxCncl        context.CancelFunc
	copyMask       zboxutil.Uint128
	maskMU         *sync.Mutex

	Consensus
}

func (co *CopyOperation) Process(allocObj *Allocation, connectionID string) ([]fileref.RefEntity, zboxutil.Uint128, error) {
	// make copyRequest object
	cR := &CopyRequest{
		allocationObj:  allocObj,
		allocationID:   allocObj.ID,
		allocationTx:   allocObj.Tx,
		connectionID:   connectionID,
		blobbers:       allocObj.Blobbers,
		remotefilepath: co.remotefilepath,
		destPath:       co.destPath,
		ctx:            co.ctx,
		ctxCncl:        co.ctxCncl,
		copyMask:       co.copyMask,
		maskMU:         co.maskMU,
	}
	cR.consensusThresh = co.consensusThresh
	cR.fullconsensus = co.fullconsensus

	objectTreeRefs, blobberErrors := cR.ProcessWithBlobbers()

	if !cR.isConsensusOk() {
		err := zboxutil.MajorError(blobberErrors)
		if err != nil {
			return nil, cR.copyMask, errors.New("copy_failed", fmt.Sprintf("Copy failed. %s", err.Error()))
		}

		return nil, cR.copyMask, errors.New("consensus_not_met",
			fmt.Sprintf("Copy failed. Required consensus %d, got %d",
				cR.Consensus.consensusThresh, cR.Consensus.consensus))
	}
	return objectTreeRefs, cR.copyMask, nil

}

func (co *CopyOperation) buildChange(refs []fileref.RefEntity, uid uuid.UUID) []allocationchange.AllocationChange {

	changes := make([]allocationchange.AllocationChange, len(refs))

	for idx, ref := range refs {
		newChange := &allocationchange.CopyFileChange{
			DestPath:   co.destPath,
			Uuid:       uid,
			ObjectTree: ref,
		}
		changes[idx] = newChange
	}
	return changes
}

func (co *CopyOperation) Verify(a *Allocation) error {

	if !a.CanCopy() {
		return constants.ErrFileOptionNotPermitted
	}

	if co.remotefilepath == "" || co.destPath == "" {
		return errors.New("invalid_path", "Invalid path for copy")
	}
	isabs := zboxutil.IsRemoteAbs(co.remotefilepath)
	if !isabs {
		return errors.New("invalid_path", "Path should be valid and absolute")
	}

	err := ValidateRemoteFileName(co.destPath)
	if err != nil {
		return err
	}
	return nil
}

func (co *CopyOperation) Completed(allocObj *Allocation) {

}

func (co *CopyOperation) Error(allocObj *Allocation, consensus int, err error) {

}

func NewCopyOperation(remotePath string, destPath string, copyMask zboxutil.Uint128, maskMU *sync.Mutex, consensusTh int, fullConsensus int, ctx context.Context) *CopyOperation {
	co := &CopyOperation{}
	co.remotefilepath = zboxutil.RemoteClean(remotePath)
	co.copyMask = copyMask
	co.maskMU = maskMU
	co.consensusThresh = consensusTh
	co.fullconsensus = fullConsensus
	if destPath != "/" {
		destPath = strings.TrimSuffix(destPath, "/")
	}
	co.destPath = destPath
	co.ctx, co.ctxCncl = context.WithCancel(ctx)
	return co

}
