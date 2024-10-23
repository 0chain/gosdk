package sdk

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0chain/errors"
	thrown "github.com/0chain/errors"
	"github.com/google/uuid"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/common"
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
	sig            string
	blobbers       []*blockchain.StorageNode
	remotefilepath string
	ctx            context.Context
	ctxCncl        context.CancelFunc
	wg             *sync.WaitGroup
	deleteMask     zboxutil.Uint128
	maskMu         *sync.Mutex
	connectionID   string
	consensus      Consensus
	timestamp      int64
}

var errFileDeleted = errors.New("file_deleted", "file is already deleted")

func (req *DeleteRequest) deleteBlobberFile(
	blobber *blockchain.StorageNode, blobberIdx int) error {

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

	httpreq, err := zboxutil.NewDeleteRequest(blobber.Baseurl, req.allocationID, req.allocationTx, req.sig, query)
	if err != nil {
		l.Logger.Error(blobber.Baseurl, "Error creating delete request", err)
		return err
	}

	var (
		resp           *http.Response
		shouldContinue bool
	)

	for i := 0; i < 3; i++ {
		err, shouldContinue = func() (err error, shouldContinue bool) {
			ctx, cncl := context.WithTimeout(req.ctx, 2*time.Minute)
			resp, err = zboxutil.Client.Do(httpreq.WithContext(ctx))
			defer cncl()

			if err != nil {
				if err == context.Canceled {
					logger.Logger.Error("context was cancelled")
					shouldContinue = true
					return
				}
				if err == io.EOF {
					shouldContinue = true
					return
				}
				logger.Logger.Error(blobber.Baseurl, "Delete: ", err)
				return
			}

			if resp.Body != nil {
				defer resp.Body.Close()
			}
			var respBody []byte

			if resp.StatusCode == http.StatusOK {
				req.consensus.Done()
				l.Logger.Debug(blobber.Baseurl, " "+req.remotefilepath, " deleted.")
				return
			}
			if resp.StatusCode == http.StatusBadRequest {
				body, err := ioutil.ReadAll(resp.Body)
				if err!= nil {
					logger.Logger.Error("Failed to read response body", err)
				}

				// Check for the specific content in the response body
				if string(body) == "file was deleted" {
					req.consensus.Done()
					l.Logger.Debug(blobber.Baseurl, " ", req.remotefilepath, " deleted.")
				}
			}

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

			if resp.StatusCode == http.StatusNoContent {
				req.consensus.Done()
				l.Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " not available in blobber.")
				return
			}

			respBody, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				l.Logger.Error(blobber.Baseurl, "Response: ", string(respBody))
				return
			}

			err = errors.New("response_error", fmt.Sprintf("unexpected response with status code %d, message: %s",
				resp.StatusCode, string(respBody)))
			return
		}()

		if err != nil {
			return err
		}

		if shouldContinue {
			continue
		}
		return nil
	}
	return errors.New("unknown_issue",
		fmt.Sprintf("latest response code: %d", resp.StatusCode))
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
		req.ctx, req.allocationID, req.allocationTx, req.sig,
		req.remotefilepath, req.blobbers[pos])
	return
}

func (req *DeleteRequest) getFileMetaFromBlobber(pos uint64) (fileRef *fileref.FileRef, err error) {
	defer func() {
		if err != nil {
			req.maskMu.Lock()
			req.deleteMask = req.deleteMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
			req.maskMu.Unlock()
		}
	}()
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

func (req *DeleteRequest) ProcessDelete() (err error) {
	defer req.ctxCncl()

	objectTreeRefs := make([]fileref.RefEntity, len(req.blobbers))
	var deleteMutex sync.Mutex
	removedNum := 0
	req.wg = &sync.WaitGroup{}

	var pos uint64
	for i := req.deleteMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		req.wg.Add(1)
		pos = uint64(i.TrailingZeros())
		go func(blobberIdx uint64) {
			defer req.wg.Done()
			refEntity, err := req.getFileMetaFromBlobber(blobberIdx)
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

	var errCount int32
	wgErrors := make(chan error)
	wgDone := make(chan bool)

	for i := req.deleteMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		req.wg.Add(1)
		pos = uint64(i.TrailingZeros())
		go func(blobberIdx uint64) {
			defer req.wg.Done()
			err = req.deleteBlobberFile(req.blobbers[blobberIdx], int(blobberIdx))
			if err != nil {
				logger.Logger.Error("error during deleteBlobberFile", err)
				errC := atomic.AddInt32(&errCount, 1)
				if errC > int32(req.consensus.fullconsensus-req.consensus.consensusThresh) {
					wgErrors <- err
				}
			}
		}(pos)
	}

	go func() {
		req.wg.Wait()
		close(wgDone)
	}()

	select {
	case <-wgDone:
		break
	case err := <-wgErrors:
		return thrown.New("delete_failed", fmt.Sprintf("Delete failed. %s", err.Error()))
	}

	if !req.consensus.isConsensusOk() {
		return errors.New("consensus_not_met",
			fmt.Sprintf("Consensus on delete failed. Required consensus %d got %d",
				req.consensus.consensusThresh, req.consensus.getConsensus()))
	}

	writeMarkerMutex, err := CreateWriteMarkerMutex(client.GetClient(), req.allocationObj)
	if err != nil {
		return fmt.Errorf("Delete failed: %s", err.Error())
	}
	err = writeMarkerMutex.Lock(
		req.ctx, &req.deleteMask, req.maskMu,
		req.blobbers, &req.consensus, removedNum, time.Minute, req.connectionID)

	if err != nil {
		return fmt.Errorf("Delete failed: %s", err.Error())
	}
	defer writeMarkerMutex.Unlock(req.ctx, req.deleteMask, req.blobbers, time.Minute, req.connectionID) //nolint: errcheck

	req.consensus.consensus = removedNum
	req.timestamp = int64(common.Now())
	wg := &sync.WaitGroup{}
	activeBlobbers := req.deleteMask.CountOnes()
	wg.Add(activeBlobbers)
	commitReqs := make([]*CommitRequest, activeBlobbers)
	var c int
	for i := req.deleteMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		newChange := &allocationchange.DeleteFileChange{}
		newChange.FileMetaRef = objectTreeRefs[pos]
		newChange.NumBlocks = newChange.FileMetaRef.GetNumBlocks()
		newChange.Operation = constants.FileOperationDelete
		newChange.Size = newChange.FileMetaRef.GetSize()
		commitReq := &CommitRequest{
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
				req.consensus.consensusThresh, req.consensus.getConsensus()))
	}
	return nil
}

type DeleteOperation struct {
	remotefilepath string
	ctx            context.Context
	ctxCncl        context.CancelFunc
	deleteMask     zboxutil.Uint128
	maskMu         *sync.Mutex
	consensus      Consensus
}

func (dop *DeleteOperation) Process(allocObj *Allocation, connectionID string) ([]fileref.RefEntity, zboxutil.Uint128, error) {
	l.Logger.Info("Started Delete Process with Connection Id", connectionID)
	deleteReq := &DeleteRequest{
		allocationObj:  allocObj,
		allocationID:   allocObj.ID,
		allocationTx:   allocObj.Tx,
		sig:            allocObj.sig,
		connectionID:   connectionID,
		blobbers:       allocObj.Blobbers,
		remotefilepath: dop.remotefilepath,
		ctx:            dop.ctx,
		ctxCncl:        dop.ctxCncl,
		deleteMask:     dop.deleteMask,
		maskMu:         dop.maskMu,
		wg:             &sync.WaitGroup{},
		consensus:      Consensus{RWMutex: &sync.RWMutex{}},
	}
	deleteReq.consensus.fullconsensus = dop.consensus.fullconsensus
	deleteReq.consensus.consensusThresh = dop.consensus.consensusThresh

	numList := len(deleteReq.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, numList)
	blobberErrors := make([]error, numList)
	versionMap := make(map[int64]int)
	var (
		pos          uint64
		consensusRef *fileref.FileRef
	)

	for i := deleteReq.deleteMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		deleteReq.wg.Add(1)
		go func(blobberIdx int) {
			defer deleteReq.wg.Done()
			refEntity, err := deleteReq.getFileMetaFromBlobber(uint64(blobberIdx))
			if errors.Is(err, constants.ErrNotFound) {
				deleteReq.consensus.Done()
				return
			} else if err != nil {
				blobberErrors[blobberIdx] = err
				l.Logger.Error(err.Error())
				return
			}
			deleteReq.consensus.Done()
			objectTreeRefs[blobberIdx] = refEntity
			deleteReq.maskMu.Lock()
			versionMap[refEntity.AllocationVersion] += 1
			if versionMap[refEntity.AllocationVersion] >= deleteReq.consensus.consensusThresh {
				consensusRef = refEntity
			}
			deleteReq.maskMu.Unlock()
		}(int(pos))
	}
	deleteReq.wg.Wait()
	if !deleteReq.consensus.isConsensusOk() {
		err := zboxutil.MajorError(blobberErrors)
		if err != nil {
			return nil, deleteReq.deleteMask, thrown.New("delete_failed", fmt.Sprintf("Delete failed. %s", err.Error()))
		}

		return nil, deleteReq.deleteMask, thrown.New("consensus_not_met",
			fmt.Sprintf("Delete failed. Required consensus %d, got %d",
				deleteReq.consensus.consensusThresh, deleteReq.consensus.consensus))
	}
	if consensusRef == nil {
		//Already deleted
		return nil, dop.deleteMask, errFileDeleted
	}
	if consensusRef.Type == fileref.DIRECTORY && !consensusRef.IsEmpty {
		for ind, refEntity := range objectTreeRefs {
			if refEntity == nil {
				continue
			}
			if refEntity.GetAllocationVersion() != consensusRef.AllocationVersion {
				deleteReq.deleteMask = deleteReq.deleteMask.And(zboxutil.NewUint128(1).Lsh(uint64(ind)).Not())
			}
		}
		err := deleteReq.deleteSubDirectories()
		if err != nil {
			return nil, deleteReq.deleteMask, err
		}
	}
	if dop.remotefilepath == "/" {
		return objectTreeRefs, deleteReq.deleteMask, nil
	}
	pos = 0
	deleteReq.consensus.Reset()
	for i := deleteReq.deleteMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		deleteReq.wg.Add(1)
		go func(blobberIdx int) {
			defer deleteReq.wg.Done()
			err := deleteReq.deleteBlobberFile(deleteReq.blobbers[blobberIdx], blobberIdx)
			if err != nil {
				logger.Logger.Error("error during deleteBlobberFile", err)
				blobberErrors[blobberIdx] = err
			}
			deleteReq.consensus.Done()
			if singleClientMode {
				lookuphash := fileref.GetReferenceLookup(deleteReq.allocationID, deleteReq.remotefilepath)
				cacheKey := fileref.GetCacheKey(lookuphash, deleteReq.blobbers[blobberIdx].ID)
				fileref.DeleteFileRef(cacheKey)
			}
		}(int(pos))
	}
	deleteReq.wg.Wait()

	if !deleteReq.consensus.isConsensusOk() {
		err := zboxutil.MajorError(blobberErrors)
		if err != nil {
			return nil, deleteReq.deleteMask, thrown.New("delete_failed", fmt.Sprintf("Delete failed. %s", err.Error()))
		}

		return nil, deleteReq.deleteMask, thrown.New("consensus_not_met",
			fmt.Sprintf("Delete failed. Required consensus %d, got %d",
				deleteReq.consensus.consensusThresh, deleteReq.consensus.consensus))
	}

	l.Logger.Debug("Delete Process Ended ")
	return objectTreeRefs, deleteReq.deleteMask, nil
}

func (do *DeleteOperation) buildChange(refs []fileref.RefEntity, uid uuid.UUID) []allocationchange.AllocationChange {

	changes := make([]allocationchange.AllocationChange, len(refs))
	for idx, ref := range refs {
		if ref == nil {
			newChange := &allocationchange.EmptyFileChange{}
			changes[idx] = newChange
		} else {
			newChange := &allocationchange.DeleteFileChange{}
			newChange.FileMetaRef = ref
			newChange.NumBlocks = newChange.FileMetaRef.GetNumBlocks()
			newChange.Operation = constants.FileOperationDelete
			newChange.Size = newChange.FileMetaRef.GetSize()
			changes[idx] = newChange
		}
	}
	return changes
}

func (dop *DeleteOperation) Verify(a *Allocation) error {

	if !a.CanDelete() {
		return constants.ErrFileOptionNotPermitted
	}

	if dop.remotefilepath == "" {
		return errors.New("invalid_path", "Invalid path for the list")
	}
	isabs := zboxutil.IsRemoteAbs(dop.remotefilepath)
	if !isabs {
		return errors.New("invalid_path", "Path should be valid and absolute")
	}
	return nil
}

func (dop *DeleteOperation) Completed(allocObj *Allocation) {

}

func (dop *DeleteOperation) Error(allocObj *Allocation, consensus int, err error) {

}

func NewDeleteOperation(remotePath string, deleteMask zboxutil.Uint128, maskMu *sync.Mutex, consensusTh int, fullConsensus int, ctx context.Context) *DeleteOperation {
	dop := &DeleteOperation{}
	dop.remotefilepath = zboxutil.RemoteClean(remotePath)
	dop.deleteMask = deleteMask
	dop.maskMu = maskMu
	dop.consensus.consensusThresh = consensusTh
	dop.consensus.fullconsensus = fullConsensus
	dop.ctx, dop.ctxCncl = context.WithCancel(ctx)
	return dop
}

func (req *DeleteRequest) deleteSubDirectories() error {
	// list all files
	var (
		offsetPath string
		pathLevel  int
	)
	for {
		oResult, err := req.allocationObj.GetRefs(req.remotefilepath, offsetPath, "", "", fileref.FILE, fileref.REGULAR, 0, getRefPageLimit, WithObjectContext(req.ctx), WithObjectMask(req.deleteMask), WithObjectConsensusThresh(req.consensus.consensusThresh), WithSingleBlobber(true))
		if err != nil {
			return err
		}
		if len(oResult.Refs) == 0 {
			break
		}
		ops := make([]OperationRequest, 0, len(oResult.Refs))
		for _, ref := range oResult.Refs {
			opMask := req.deleteMask
			if ref.Type == fileref.DIRECTORY {
				continue
			}
			if ref.PathLevel > pathLevel {
				pathLevel = ref.PathLevel
			}
			op := OperationRequest{
				OperationType: constants.FileOperationDelete,
				RemotePath:    ref.Path,
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
	// reset offsetPath
	offsetPath = ""
	level := len(strings.Split(strings.TrimSuffix(req.remotefilepath, "/"), "/"))
	if pathLevel == 0 {
		pathLevel = level + 1
	}
	// list all directories by descending order of path level
	for pathLevel > level {
		oResult, err := req.allocationObj.GetRefs(req.remotefilepath, offsetPath, "", "", fileref.DIRECTORY, fileref.REGULAR, pathLevel, getRefPageLimit, WithObjectContext(req.ctx), WithObjectMask(req.deleteMask), WithObjectConsensusThresh(req.consensus.consensusThresh), WithSingleBlobber(true))
		if err != nil {
			return err
		}
		if len(oResult.Refs) == 0 {
			pathLevel--
		} else {
			ops := make([]OperationRequest, 0, len(oResult.Refs))
			for _, ref := range oResult.Refs {
				opMask := req.deleteMask
				op := OperationRequest{
					OperationType: constants.FileOperationDelete,
					RemotePath:    ref.Path,
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
				pathLevel--
			}
		}
	}

	return nil
}
