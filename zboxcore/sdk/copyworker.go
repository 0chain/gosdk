package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path"
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
	dirOnly        bool
	ctx            context.Context
	ctxCncl        context.CancelFunc
	copyMask       zboxutil.Uint128
	maskMU         *sync.Mutex
	connectionID   string
	timestamp      int64
	Consensus
}

const objAlreadyExists = "Object Already exists"

func (req *CopyRequest) getObjectTreeFromBlobber(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	return getObjectTreeFromBlobber(req.ctx, req.allocationID, req.allocationTx, req.remotefilepath, blobber)
}

func (req *CopyRequest) getFileMetaFromBlobber(pos int) (fileRef *fileref.FileRef, err error) {
	listReq := &ListRequest{
		allocationID:   req.allocationID,
		allocationTx:   req.allocationTx,
		blobbers:       req.blobbers,
		remotefilepath: req.remotefilepath,
		ctx:            req.ctx,
	}
	respChan := make(chan *fileMetaResponse)
	go listReq.getFileMetaInfoFromBlobber(req.blobbers[pos], int(pos), respChan)
	refRes := <-respChan
	if refRes.err != nil {
		err = refRes.err
		return
	}
	fileRef = refRes.fileref
	return
}

func (req *CopyRequest) copyBlobberObject(
	blobber *blockchain.StorageNode, blobberIdx int) (err error) {

	defer func() {
		if err != nil {
			req.maskMU.Lock()
			// Removing blobber from mask
			req.copyMask = req.copyMask.And(zboxutil.NewUint128(1).Lsh(uint64(blobberIdx)).Not())
			req.maskMU.Unlock()
		}
	}()
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
	return errors.New("unknown_issue",
		fmt.Sprintf("last status code: %d, last response message: %s", latestStatusCode, latestRespMsg))
}

func (req *CopyRequest) ProcessWithBlobbers() ([]fileref.RefEntity, error) {
	var (
		pos          uint64
		consensusRef *fileref.FileRef
	)
	numList := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, numList)
	blobberErrors := make([]error, numList)
	versionMap := make(map[int64]int)

	wg := &sync.WaitGroup{}
	for i := req.copyMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		wg.Add(1)
		go func(blobberIdx int) {
			defer wg.Done()
			// refEntity, err := req.copyBlobberObject(req.blobbers[blobberIdx], blobberIdx)
			refEntity, err := req.getFileMetaFromBlobber(blobberIdx)
			if err != nil {
				blobberErrors[blobberIdx] = err
				l.Logger.Debug(err.Error())
				return
			}
			objectTreeRefs[blobberIdx] = refEntity
			req.maskMU.Lock()
			versionMap[refEntity.AllocationVersion] += 1
			if versionMap[refEntity.AllocationVersion] >= req.consensusThresh {
				consensusRef = refEntity
			}
			req.maskMU.Unlock()
		}(int(pos))
	}
	wg.Wait()
	if consensusRef == nil {
		return nil, zboxutil.MajorError(blobberErrors)
	}

	if consensusRef.Type == fileref.DIRECTORY && !consensusRef.IsEmpty {
		for ind, refEntity := range objectTreeRefs {
			if refEntity.GetAllocationVersion() != consensusRef.AllocationVersion {
				req.copyMask = req.copyMask.And(zboxutil.NewUint128(1).Lsh(uint64(ind)).Not())
			}
		}
		err := req.copySubDirectoriees(req.dirOnly)
		if err != nil {
			return nil, err
		}
		req.consensus = req.copyMask.CountOnes()
		return objectTreeRefs, errNoChange
	}

	for i := req.copyMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		wg.Add(1)
		go func(blobberIdx int) {
			defer wg.Done()
			err := req.copyBlobberObject(req.blobbers[blobberIdx], blobberIdx)
			if err != nil {
				blobberErrors[blobberIdx] = err
				l.Logger.Debug(err.Error())
				return
			}
		}(int(pos))
	}
	wg.Wait()
	err := zboxutil.MajorError(blobberErrors)
	if err != nil && strings.Contains(err.Error(), objAlreadyExists) && consensusRef.Type == fileref.DIRECTORY {
		return nil, errNoChange
	}

	return objectTreeRefs, err
}

func (req *CopyRequest) ProcessCopy() error {
	defer req.ctxCncl()

	wg := &sync.WaitGroup{}
	var pos uint64

	objectTreeRefs, err := req.ProcessWithBlobbers()

	if !req.isConsensusOk() {
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
	status, _, err := req.allocationObj.CheckAllocStatus()
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
	dirOnly        bool
	ctx            context.Context
	ctxCncl        context.CancelFunc
	copyMask       zboxutil.Uint128
	maskMU         *sync.Mutex

	Consensus
}

func (co *CopyOperation) Process(allocObj *Allocation, connectionID string) ([]fileref.RefEntity, zboxutil.Uint128, error) {
	if co.remotefilepath == "/" {
		return nil, co.copyMask, errors.New("invalid_path", "Invalid path for copy cannot copy root directory")
	}
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
		Consensus:      Consensus{RWMutex: &sync.RWMutex{}},
		dirOnly:        co.dirOnly,
	}
	cR.consensusThresh = co.consensusThresh
	cR.fullconsensus = co.fullconsensus

	objectTreeRefs, err := cR.ProcessWithBlobbers()

	if !cR.isConsensusOk() {
		if err != nil {
			if err == errNoChange {
				return nil, cR.copyMask, err
			}
			l.Logger.Error("copy failed: ", cR.remotefilepath, cR.destPath)
			return nil, cR.copyMask, errors.New("copy_failed", fmt.Sprintf("Copy failed. %s", err.Error()))
		}

		return nil, cR.copyMask, errors.New("consensus_not_met",
			fmt.Sprintf("Copy failed. Required consensus %d, got %d",
				cR.Consensus.consensusThresh, cR.Consensus.consensus))
	}
	return objectTreeRefs, cR.copyMask, err

}

func (co *CopyOperation) buildChange(refs []fileref.RefEntity, uid uuid.UUID) []allocationchange.AllocationChange {

	changes := make([]allocationchange.AllocationChange, len(refs))

	for idx, ref := range refs {
		if ref == nil {
			change := &allocationchange.EmptyFileChange{}
			changes[idx] = change
			continue
		}
		newChange := &allocationchange.CopyFileChange{
			DestPath:   co.destPath,
			Uuid:       uid,
			ObjectTree: ref,
		}
		newChange.Operation = constants.FileOperationCopy
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

func NewCopyOperation(remotePath string, destPath string, copyMask zboxutil.Uint128, maskMU *sync.Mutex, consensusTh int, fullConsensus int, copyDirOnly bool, ctx context.Context) *CopyOperation {
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

func (req *CopyRequest) copySubDirectoriees(dirOnly bool) error {
	var (
		offsetPath string
		pathLevel  int
	)

	for {
		if !dirOnly {
			oResult, err := req.allocationObj.GetRefs(req.remotefilepath, offsetPath, "", "", fileref.FILE, fileref.REGULAR, 0, getRefPageLimit, WithObjectContext(req.ctx), WithObjectConsensusThresh(req.consensusThresh), WithSingleBlobber(true))
			if err != nil {
				return err
			}
			if len(oResult.Refs) == 0 {
				break
			}
			ops := make([]OperationRequest, 0, len(oResult.Refs))
			for _, ref := range oResult.Refs {
				opMask := req.copyMask
				if ref.Type == fileref.DIRECTORY {
					continue
				}
				if ref.PathLevel > pathLevel {
					pathLevel = ref.PathLevel
				}
				basePath := strings.TrimPrefix(path.Dir(ref.Path), path.Dir(req.remotefilepath))
				destPath := path.Join(req.destPath, basePath)
				op := OperationRequest{
					OperationType: constants.FileOperationCopy,
					RemotePath:    ref.Path,
					DestPath:      destPath,
					Mask:          &opMask,
				}
				ops = append(ops, op)
			}
			err = req.allocationObj.DoMultiOperation(ops)
			if err != nil {
				return err
			}
			offsetPath = oResult.Refs[len(oResult.Refs)-1].Path
			if len(oResult.Refs) < getRefPageLimit {
				break
			}
		}
	}

	offsetPath = ""
	level := len(strings.Split(strings.TrimSuffix(req.remotefilepath, "/"), "/"))
	if pathLevel == 0 {
		pathLevel = level + 1
	}

	for pathLevel > level {
		oResult, err := req.allocationObj.GetRefs(req.remotefilepath, offsetPath, "", "", fileref.DIRECTORY, fileref.REGULAR, pathLevel, getRefPageLimit, WithObjectContext(req.ctx), WithObjectMask(req.copyMask), WithObjectConsensusThresh(req.consensusThresh), WithSingleBlobber(true))
		if err != nil {
			return err
		}
		if len(oResult.Refs) == 0 {
			pathLevel--
		} else {
			ops := make([]OperationRequest, 0, len(oResult.Refs))
			for _, ref := range oResult.Refs {
				opMask := req.copyMask
				if ref.Type == fileref.FILE {
					continue
				}
				basePath := strings.TrimPrefix(path.Dir(ref.Path), path.Dir(req.remotefilepath))
				destPath := path.Join(req.destPath, basePath)
				op := OperationRequest{
					OperationType: constants.FileOperationCopy,
					RemotePath:    ref.Path,
					DestPath:      destPath,
					Mask:          &opMask,
					CopyDirOnly:   true,
				}
				ops = append(ops, op)
			}
			err = req.allocationObj.DoMultiOperation(ops)
			if err != nil {
				return err
			}
			offsetPath = oResult.Refs[len(oResult.Refs)-1].Path
			if len(oResult.Refs) < getRefPageLimit {
				pathLevel--
			}
		}
	}

	return nil
}
