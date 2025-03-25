package sdk

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/0chain/common/core/util/wmpt"
	"github.com/0chain/errors"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/util"
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
	sig            string
	blobbers       []*blockchain.StorageNode
	remotefilepath string
	destPath       string
	ctx            context.Context
	ctxCncl        context.CancelFunc
	copyMask       zboxutil.Uint128
	maskMU         *sync.Mutex
	connectionID   string
	timestamp      int64
	dirOnly        bool
	destLookupHash string
	Consensus
}

var errNoChange = errors.New("no_change", "No change in the operation")

const objAlreadyExists = "Object Already exists"

func (req *CopyRequest) getObjectTreeFromBlobber(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	return getObjectTreeFromBlobber(req.ctx, req.allocationID, req.allocationTx, req.sig, req.remotefilepath, blobber, req.allocationObj.Owner)
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
	blobber *blockchain.StorageNode, blobberIdx int, fetchObjectTree bool) (refEntity fileref.RefEntity, err error) {

	defer func() {
		if err != nil {
			req.maskMU.Lock()
			// Removing blobber from mask
			req.copyMask = req.copyMask.And(zboxutil.NewUint128(1).Lsh(uint64(blobberIdx)).Not())
			req.maskMU.Unlock()
		}
	}()
	if fetchObjectTree {
		refEntity, err = req.getObjectTreeFromBlobber(req.blobbers[blobberIdx])
		if err != nil {
			return nil, err
		}
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

			httpreq, err = zboxutil.NewCopyRequest(blobber.Baseurl, req.allocationID, req.allocationTx, req.sig, body, req.allocationObj.Owner)
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
			respBody, err = io.ReadAll(resp.Body)
			if err != nil {
				logger.Logger.Error("Error: Resp ", err)
				return
			}

			if resp.StatusCode == http.StatusOK {
				l.Logger.Debug(blobber.Baseurl, " "+req.remotefilepath, " copied.")
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

func (req *CopyRequest) ProcessWithBlobbers() ([]fileref.RefEntity, error) {
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
			refEntity, err := req.copyBlobberObject(req.blobbers[blobberIdx], blobberIdx, true)
			if err != nil {
				blobberErrors[blobberIdx] = err
				l.Logger.Debug(err.Error())
				return
			}
			objectTreeRefs[blobberIdx] = refEntity
		}(int(pos))
	}
	wg.Wait()
	var err error
	if !req.isConsensusOk() {
		err = zboxutil.MajorError(blobberErrors)
	}
	return objectTreeRefs, err
}

func (req *CopyRequest) ProcessWithBlobbersV2() ([]fileref.RefEntity, error) {
	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 4096)
			length := runtime.Stack(stack, false)
			l.Logger.Error("PANIC in ProcessWithBlobbersV2",
				"error", r,
				"stack", string(stack[:length]),
				"remotefilepath", req.remotefilepath,
				"destPath", req.destPath)
		}
	}()

	if req.maskMU == nil {
		return nil, errors.New("copy_failed", "maskMU is nil")
	}

	if req.blobbers == nil || len(req.blobbers) == 0 {
		return nil, errors.New("copy_failed", "blobbers list is empty or nil")
	}

	var (
		pos          uint64
		consensusRef *fileref.FileRef
	)

	id := time.Now().UnixNano()
	l.Logger.Debug("Starting ProcessWithBlobbersV2, ",
		"copyMask:", req.copyMask,
		" blobbers_count:", len(req.blobbers),
		" remotefilepath:", req.remotefilepath,
		" id:", id)

	defer func() {
		l.Logger.Debug("process with blobbersV2 end", " id:", id)
	}()

	numList := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, numList)
	blobberErrors := make([]error, numList)
	versionMap := make(map[string]int)
	versionMapMutex := &sync.Mutex{} // Add mutex for concurrent map access

	wg := &sync.WaitGroup{}
	for i := req.copyMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		if int(pos) >= numList {
			l.Logger.Error("Position out of range", "pos", pos, "numList", numList)
			continue
		}

		wg.Add(1)
		go func(blobberIdx int) {
			defer func() {
				if r := recover(); r != nil {
					stack := make([]byte, 4096)
					length := runtime.Stack(stack, false)
					l.Logger.Error("PANIC in goroutine",
						"error", r,
						"stack", string(stack[:length]),
						"blobberIdx", blobberIdx)
					blobberErrors[blobberIdx] = errors.New("internal_error", fmt.Sprintf("Panic: %v", r))
				}
				l.Logger.Debug("Finishing first goroutine", "blobberIdx", blobberIdx, "id", id)
				wg.Done()
			}()

			l.Logger.Debug("Starting getFileMetaFromBlobber", "blobberIdx", blobberIdx, "id", id)
			// refEntity, err := req.copyBlobberObject(req.blobbers[blobberIdx], blobberIdx)
			refEntity, err := req.getFileMetaFromBlobber(blobberIdx)
			if err != nil {
				blobberErrors[blobberIdx] = err
				l.Logger.Debug("Error getting file meta", "error", err.Error(), "blobberIdx", blobberIdx)
				return
			}

			if refEntity == nil {
				blobberErrors[blobberIdx] = errors.New("internal_error", "refEntity is nil")
				l.Logger.Debug("refEntity is nil", "blobberIdx", blobberIdx)
				return
			}

			refEntity.Path = path.Join(req.destPath, path.Base(refEntity.Path))
			objectTreeRefs[blobberIdx] = refEntity

			req.maskMU.Lock()
			versionMapMutex.Lock() // Lock for concurrent map access
			versionMap[refEntity.AllocationRoot] += 1
			if versionMap[refEntity.AllocationRoot] >= req.consensusThresh {
				consensusRef = refEntity
			}
			versionMapMutex.Unlock() // Unlock after map access
			req.maskMU.Unlock()
		}(int(pos))
	}

	// Add timeout to prevent indefinite waiting
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		l.Logger.Debug("First stage completed normally", "id", id)
	case <-time.After(5 * time.Minute): // 5 minute timeout
		l.Logger.Error("Timeout waiting for first stage goroutines", "id", id)
		return nil, errors.New("timeout_error", "Timeout waiting for file metadata")
	}

	l.Logger.Debug("All goroutines completed", "consensusRef", consensusRef != nil, "id", id)

	if consensusRef == nil {
		return nil, zboxutil.MajorError(blobberErrors)
	}

	if consensusRef.Type == fileref.DIRECTORY && !consensusRef.IsEmpty {
		for ind, refEntity := range objectTreeRefs {
			if refEntity == nil {
				continue
			}
			if refEntity.GetAllocationRoot() != consensusRef.AllocationRoot {
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

	pos = 0 // Reset position counter for second loop
	var procWg sync.WaitGroup
	for i := req.copyMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		if int(pos) >= numList {
			l.Logger.Error("Position out of range", "pos", pos, "numList", numList)
			continue
		}

		procWg.Add(1)
		go func(blobberIdx int) {
			defer func() {
				if r := recover(); r != nil {
					stack := make([]byte, 4096)
					length := runtime.Stack(stack, false)
					l.Logger.Error("PANIC in second goroutine",
						"error", r,
						"stack", string(stack[:length]),
						"blobberIdx", blobberIdx)
					blobberErrors[blobberIdx] = errors.New("internal_error", fmt.Sprintf("Panic: %v", r))
				}
				l.Logger.Debug("Finishing second goroutine", "blobberIdx", blobberIdx, "id", id)
				procWg.Done()
			}()

			l.Logger.Debug("Starting copyBlobberObject", "blobberIdx", blobberIdx, "id", id)
			_, err := req.copyBlobberObject(req.blobbers[blobberIdx], blobberIdx, false)
			if err != nil {
				blobberErrors[blobberIdx] = err
				l.Logger.Debug("Error copying blobber object", "error", err.Error(), "blobberIdx", blobberIdx)
				return
			}
		}(int(pos))
	}

	// Add timeout for second stage
	procWaitChan := make(chan struct{})
	go func() {
		procWg.Wait()
		close(procWaitChan)
	}()

	select {
	case <-procWaitChan:
		l.Logger.Debug("Second stage completed normally", "id", id)
	case <-time.After(5 * time.Minute): // 5 minute timeout
		l.Logger.Error("Timeout waiting for second stage goroutines", "id", id)
		return nil, errors.New("timeout_error", "Timeout waiting for copy operations")
	}

	var err error
	if !req.isConsensusOk() {
		err = zboxutil.MajorError(blobberErrors)
		if err != nil && strings.Contains(err.Error(), objAlreadyExists) && consensusRef.Type == fileref.DIRECTORY {
			return nil, errNoChange
		}
	}
	req.destLookupHash = fileref.GetReferenceLookup(req.allocationID, consensusRef.Path)

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

	writeMarkerMutex, err := CreateWriteMarkerMutex(req.allocationObj)
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
			ClientId:     req.allocationObj.Owner,
			allocationID: req.allocationID,
			allocationTx: req.allocationTx,
			sig:          req.sig,
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
	destLookupHash string
	dirOnly        bool
	ctx            context.Context
	ctxCncl        context.CancelFunc
	copyMask       zboxutil.Uint128
	maskMU         *sync.Mutex
	objectTreeRefs []fileref.RefEntity

	Consensus
}

func (co *CopyOperation) Process(allocObj *Allocation, connectionID string) ([]fileref.RefEntity, zboxutil.Uint128, error) {
	// make copyRequest object
	cR := &CopyRequest{
		allocationObj:  allocObj,
		allocationID:   allocObj.ID,
		allocationTx:   allocObj.Tx,
		sig:            allocObj.sig,
		connectionID:   connectionID,
		blobbers:       allocObj.Blobbers,
		remotefilepath: co.remotefilepath,
		destPath:       co.destPath,
		ctx:            co.ctx,
		ctxCncl:        co.ctxCncl,
		copyMask:       co.copyMask,
		maskMU:         co.maskMU,
		dirOnly:        co.dirOnly,
		Consensus:      Consensus{RWMutex: &sync.RWMutex{}},
	}

	cR.consensusThresh = co.consensusThresh
	cR.fullconsensus = co.fullconsensus
	var err error
	if allocObj.StorageVersion == StorageV2 {
		co.objectTreeRefs, err = cR.ProcessWithBlobbersV2()
	} else {
		co.objectTreeRefs, err = cR.ProcessWithBlobbers()
	}

	if !cR.isConsensusOk() {
		l.Logger.Error("copy failed: ", cR.remotefilepath, cR.destPath)
		if err != nil {
			if err == errNoChange {
				return nil, cR.copyMask, err
			}
			return nil, cR.copyMask, errors.New("copy_failed", fmt.Sprintf("Copy failed. %s", err.Error()))
		}

		return nil, cR.copyMask, errors.New("consensus_not_met",
			fmt.Sprintf("Copy failed. Required consensus %d, got %d",
				cR.Consensus.consensusThresh, cR.Consensus.consensus))
	}
	co.destLookupHash = cR.destLookupHash
	return co.objectTreeRefs, cR.copyMask, err

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

func NewCopyOperation(ctx context.Context, remotePath string, destPath string, copyMask zboxutil.Uint128, maskMU *sync.Mutex, consensusTh, fullConsensus int, copyDirOnly bool) *CopyOperation {
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
	co.dirOnly = copyDirOnly
	return co

}

func (co *CopyOperation) ProcessChangeV2(trie *wmpt.WeightedMerkleTrie, changeIndex uint64) error {
	if co.objectTreeRefs == nil || co.objectTreeRefs[changeIndex] == nil || co.objectTreeRefs[changeIndex].GetType() == fileref.DIRECTORY {
		return nil
	}
	decodedDestHash, _ := hex.DecodeString(co.destLookupHash)
	ref := co.objectTreeRefs[changeIndex]
	numBlocks := uint64(ref.GetNumBlocks())
	fileMetaRawHash := ref.GetFileMetaHashV2()
	err := trie.Update(decodedDestHash, fileMetaRawHash, numBlocks)
	if err != nil {
		l.Logger.Error("Error updating trie", zap.Error(err))
		return err
	}
	return nil
}

func (co *CopyOperation) GetLookupHash(changeIndex uint64) []string {
	if co.objectTreeRefs == nil || co.objectTreeRefs[changeIndex] == nil || co.objectTreeRefs[changeIndex].GetType() == fileref.DIRECTORY {
		return nil
	}
	return []string{co.destLookupHash}
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
