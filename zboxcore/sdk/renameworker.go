package sdk

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/0chain/common/core/util/wmpt"
	"github.com/0chain/errors"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"

	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type RenameRequest struct {
	allocationObj  *Allocation
	allocationID   string
	allocationTx   string
	sig            string
	blobbers       []*blockchain.StorageNode
	remotefilepath string
	newName        string
	ctx            context.Context
	ctxCncl        context.CancelFunc
	wg             *sync.WaitGroup
	renameMask     zboxutil.Uint128
	maskMU         *sync.Mutex
	connectionID   string
	consensus      Consensus
	timestamp      int64
}

func (req *RenameRequest) getObjectTreeFromBlobber(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	return getObjectTreeFromBlobber(req.ctx, req.allocationID, req.allocationTx, req.sig, req.remotefilepath, blobber)
}

func (req *RenameRequest) getFileMetaFromBlobber(pos int) (fileRef *fileref.FileRef, err error) {
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

func (req *RenameRequest) renameBlobberObject(
	blobber *blockchain.StorageNode, blobberIdx int, fetchObjectTree bool) (refEntity fileref.RefEntity, err error) {

	defer func() {
		if err != nil {
			req.maskMU.Lock()
			req.renameMask = req.renameMask.And(zboxutil.NewUint128(1).Lsh(uint64(blobberIdx)).Not())
			req.maskMU.Unlock()
		}
	}()
	if fetchObjectTree {
		refEntity, err = req.getObjectTreeFromBlobber(req.blobbers[blobberIdx])
		if err != nil {
			return nil, err
		}
	}

	var (
		resp             *http.Response
		shouldContinue   bool
		latestRespMsg    string
		latestStatusCode int
	)

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

			err = formWriter.WriteField("new_name", req.newName)
			if err != nil {
				return err, false
			}

			formWriter.Close()

			var httpreq *http.Request
			httpreq, err = zboxutil.NewRenameRequest(blobber.Baseurl, req.allocationID, req.allocationTx, req.sig, body)
			if err != nil {
				l.Logger.Error(blobber.Baseurl, "Error creating rename request", err)
				return
			}

			httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
			ctx, cncl := context.WithTimeout(req.ctx, DefaultUploadTimeOut)
			resp, err = zboxutil.Client.Do(httpreq.WithContext(ctx))
			defer cncl()

			if err != nil {
				logger.Logger.Error("Rename: ", err)
				return
			}

			if resp.Body != nil {
				defer resp.Body.Close()
			}
			var respBody []byte
			respBody, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Logger.Error("Error: Resp ", err)
				return
			}

			latestRespMsg = string(respBody)
			latestStatusCode = resp.StatusCode

			if resp.StatusCode == http.StatusOK {
				req.consensus.Done()
				l.Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " renamed.")
				return
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

	err = errors.New("unknown_issue",
		fmt.Sprintf("last status code: %d, last response message: %s", latestStatusCode, latestRespMsg))
	return
}

func (req *RenameRequest) ProcessWithBlobbers() ([]fileref.RefEntity, error) {
	var pos uint64
	numList := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, numList)
	blobberErrors := make([]error, numList)
	req.wg = &sync.WaitGroup{}
	for i := req.renameMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		req.wg.Add(1)
		go func(blobberIdx int) {
			defer req.wg.Done()
			refEntity, err := req.renameBlobberObject(req.blobbers[blobberIdx], blobberIdx, true)
			if err != nil {
				blobberErrors[blobberIdx] = err
				l.Logger.Error(err.Error())
				return
			}
			objectTreeRefs[blobberIdx] = refEntity
		}(int(pos))
	}
	req.wg.Wait()
	return objectTreeRefs, zboxutil.MajorError(blobberErrors)
}

func (req *RenameRequest) ProcessWithBlobbersV2() ([]fileref.RefEntity, error) {
	var (
		pos          uint64
		consensusRef *fileref.FileRef
	)
	numList := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, numList)
	blobberErrors := make([]error, numList)
	versionMap := make(map[string]int)
	req.wg = &sync.WaitGroup{}
	for i := req.renameMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		req.wg.Add(1)
		go func(blobberIdx int) {
			defer req.wg.Done()
			refEntity, err := req.getFileMetaFromBlobber(blobberIdx)
			if err != nil {
				blobberErrors[blobberIdx] = err
				l.Logger.Error(err.Error())
				return
			}
			refEntity.Path = path.Join(path.Dir(req.remotefilepath), req.newName)
			objectTreeRefs[blobberIdx] = refEntity
			req.maskMU.Lock()
			versionMap[refEntity.AllocationRoot] += 1
			if versionMap[refEntity.AllocationRoot] >= req.consensus.consensusThresh {
				consensusRef = refEntity
			}
			req.maskMU.Unlock()
		}(int(pos))
	}
	req.wg.Wait()
	if consensusRef == nil {
		return nil, zboxutil.MajorError(blobberErrors)
	}
	if consensusRef.Type == fileref.DIRECTORY && !consensusRef.IsEmpty {
		for ind, refEntity := range objectTreeRefs {
			if refEntity.GetAllocationRoot() != consensusRef.AllocationRoot {
				req.renameMask = req.renameMask.And(zboxutil.NewUint128(1).Lsh(uint64(ind)).Not())
			}
		}
		subRequest := &subDirRequest{
			allocationObj:   req.allocationObj,
			remotefilepath:  req.remotefilepath,
			destPath:        path.Join(path.Dir(req.remotefilepath), req.newName),
			ctx:             req.ctx,
			consensusThresh: req.consensus.consensusThresh,
			opType:          constants.FileOperationMove,
			subOpType:       constants.FileOperationRename,
			mask:            req.renameMask,
		}
		err := subRequest.processSubDirectories()
		if err != nil {
			return nil, err
		}
		op := OperationRequest{
			OperationType: constants.FileOperationDelete,
			RemotePath:    req.remotefilepath,
			Mask:          &req.renameMask,
		}
		err = req.allocationObj.DoMultiOperation([]OperationRequest{op})
		if err != nil {
			return nil, err
		}
		req.consensus.consensus = req.renameMask.CountOnes()
		return nil, errNoChange
	}

	for i := req.renameMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		req.wg.Add(1)
		go func(blobberIdx int) {
			defer req.wg.Done()
			_, err := req.renameBlobberObject(req.blobbers[blobberIdx], blobberIdx, false)
			if err != nil {
				blobberErrors[blobberIdx] = err
				l.Logger.Debug(err.Error())
				return
			}
		}(int(pos))
	}
	req.wg.Wait()

	return objectTreeRefs, zboxutil.MajorError(blobberErrors)
}

func (req *RenameRequest) ProcessRename() error {
	defer req.ctxCncl()

	objectTreeRefs, err := req.ProcessWithBlobbers()

	if !req.consensus.isConsensusOk() {
		if err != nil {
			return errors.New("rename_failed",
				fmt.Sprintf("Rename failed. %s", err.Error()))
		}

		return errors.New("consensus_not_met",
			fmt.Sprintf("Rename failed. Required consensus %d got %d",
				req.consensus.consensusThresh, req.consensus.getConsensus()))
	}

	writeMarkerMutex, err := CreateWriteMarkerMutex(client.GetClient(), req.allocationObj)
	if err != nil {
		return fmt.Errorf("rename failed: %s", err.Error())
	}

	err = writeMarkerMutex.Lock(req.ctx, &req.renameMask,
		req.maskMU, req.blobbers, &req.consensus, 0, time.Minute, req.connectionID)
	if err != nil {
		return fmt.Errorf("rename failed: %s", err.Error())
	}
	defer writeMarkerMutex.Unlock(req.ctx, req.renameMask, req.blobbers, time.Minute, req.connectionID) //nolint: errcheck

	//Check if the allocation is to be repaired or rolled back
	status, _, err := req.allocationObj.CheckAllocStatus()
	if err != nil {
		logger.Logger.Error("Error checking allocation status: ", err)
		return fmt.Errorf("rename failed: %s", err.Error())
	}

	if status == Repair {
		logger.Logger.Info("Repairing allocation")
		//TODO: Need status callback to call repair allocation
		// err = req.allocationObj.RepairAlloc()
		// if err != nil {
		// 	return err
		// }
	}
	if status != Commit {
		return ErrRetryOperation
	}

	req.consensus.Reset()
	req.timestamp = int64(common.Now())
	activeBlobbers := req.renameMask.CountOnes()
	wg := &sync.WaitGroup{}
	wg.Add(activeBlobbers)
	commitReqs := make([]*CommitRequest, activeBlobbers)
	var pos uint64
	var counter int
	for i := req.renameMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		newChange := &allocationchange.RenameFileChange{
			NewName:    req.newName,
			ObjectTree: objectTreeRefs[pos],
		}
		newChange.Operation = constants.FileOperationRename
		newChange.Size = 0

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
		commitReqs[counter] = commitReq

		go AddCommitRequest(commitReq)

		counter++
	}

	wg.Wait()

	var errMessages string
	for _, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				l.Logger.Info("Commit success", commitReq.blobber.Baseurl)
				req.consensus.Done()
			} else {
				errMessages += commitReq.result.ErrorMessage + "\t"
				l.Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			l.Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !req.consensus.isConsensusOk() {
		return errors.New("consensus_not_met",
			fmt.Sprintf("Required consensus %d got %d. Error: %s",
				req.consensus.consensusThresh, req.consensus.consensus, errMessages))
	}
	return nil
}

type RenameOperation struct {
	remotefilepath string
	srcLookupHash  string
	destLookupHash string
	ctx            context.Context
	ctxCncl        context.CancelFunc
	renameMask     zboxutil.Uint128
	newName        string
	maskMU         *sync.Mutex
	objectTreeRefs []fileref.RefEntity

	consensus Consensus
}

func (ro *RenameOperation) Process(allocObj *Allocation, connectionID string) ([]fileref.RefEntity, zboxutil.Uint128, error) {
	// make renameRequest object
	rR := &RenameRequest{
		allocationObj:  allocObj,
		allocationID:   allocObj.ID,
		allocationTx:   allocObj.Tx,
		sig:            allocObj.sig,
		connectionID:   connectionID,
		blobbers:       allocObj.Blobbers,
		remotefilepath: ro.remotefilepath,
		newName:        ro.newName,
		ctx:            ro.ctx,
		ctxCncl:        ro.ctxCncl,
		renameMask:     ro.renameMask,
		maskMU:         ro.maskMU,
		wg:             &sync.WaitGroup{},
		consensus:      Consensus{RWMutex: &sync.RWMutex{}},
	}
	if filepath.Base(ro.remotefilepath) == ro.newName {
		return nil, ro.renameMask, errors.New("invalid_operation", "Cannot rename to same name")
	}
	rR.consensus.fullconsensus = ro.consensus.fullconsensus
	rR.consensus.consensusThresh = ro.consensus.consensusThresh
	var err error
	if allocObj.StorageVersion == StorageV2 {
		ro.objectTreeRefs, err = rR.ProcessWithBlobbersV2()
	} else {
		ro.objectTreeRefs, err = rR.ProcessWithBlobbers()
	}

	if !rR.consensus.isConsensusOk() {
		if err != nil {
			if err == errNoChange {
				return nil, rR.renameMask, err
			}
			return nil, rR.renameMask, errors.New("rename_failed", fmt.Sprintf("Renamed failed. %s", err.Error()))
		}

		return nil, rR.renameMask, errors.New("consensus_not_met",
			fmt.Sprintf("Rename failed. Required consensus %d, got %d",
				rR.consensus.consensusThresh, rR.consensus.consensus))
	}
	ro.srcLookupHash = fileref.GetReferenceLookup(rR.allocationID, rR.remotefilepath)
	destPath := path.Join(path.Dir(rR.remotefilepath), rR.newName)
	ro.destLookupHash = fileref.GetReferenceLookup(rR.allocationID, destPath)
	return ro.objectTreeRefs, rR.renameMask, err
}

func (ro *RenameOperation) buildChange(refs []fileref.RefEntity, uid uuid.UUID) []allocationchange.AllocationChange {
	changes := make([]allocationchange.AllocationChange, len(refs))

	for idx, ref := range refs {
		if ref == nil {
			change := &allocationchange.EmptyFileChange{}
			changes[idx] = change
			continue
		}
		newChange := &allocationchange.RenameFileChange{
			NewName:    ro.newName,
			ObjectTree: ref,
		}

		newChange.Operation = constants.FileOperationRename
		newChange.Size = 0
		changes[idx] = newChange
	}
	return changes
}

func (ro *RenameOperation) Verify(a *Allocation) error {

	if !a.CanRename() {
		return constants.ErrFileOptionNotPermitted
	}

	if ro.remotefilepath == "" {
		return errors.New("invalid_path", "Invalid path for the list")
	}

	if ro.remotefilepath == "/" {
		return errors.New("invalid_operation", "cannot rename root path")
	}

	isabs := zboxutil.IsRemoteAbs(ro.remotefilepath)
	if !isabs {
		return errors.New("invalid_path", "Path should be valid and absolute")
	}

	err := ValidateRemoteFileName(ro.newName)
	if err != nil {
		return err
	}

	return nil
}

func (ro *RenameOperation) Completed(allocObj *Allocation) {

}

func (ro *RenameOperation) Error(allocObj *Allocation, consensus int, err error) {

}

func NewRenameOperation(remotePath string, destName string, renameMask zboxutil.Uint128, maskMU *sync.Mutex, consensusTh int, fullConsensus int, ctx context.Context) *RenameOperation {
	ro := &RenameOperation{}
	ro.remotefilepath = zboxutil.RemoteClean(remotePath)
	ro.newName = path.Base(destName)
	ro.renameMask = renameMask
	ro.maskMU = maskMU
	ro.consensus.consensusThresh = consensusTh
	ro.consensus.fullconsensus = fullConsensus
	ro.ctx, ro.ctxCncl = context.WithCancel(ctx)
	return ro

}

func (ro *RenameOperation) ProcessChangeV2(trie *wmpt.WeightedMerkleTrie, changeIndex uint64) error {
	if ro.objectTreeRefs == nil || ro.objectTreeRefs[changeIndex] == nil || ro.objectTreeRefs[changeIndex].GetType() == fileref.DIRECTORY {
		return nil
	}
	decodedSrcHash, _ := hex.DecodeString(ro.srcLookupHash)
	err := trie.Update(decodedSrcHash, nil, 0)
	if err != nil {
		l.Logger.Error("Error updating trie", zap.Error(err))
		return err
	}
	decodedDestHash, _ := hex.DecodeString(ro.destLookupHash)
	ref := ro.objectTreeRefs[changeIndex]
	numBlocks := uint64(ref.GetNumBlocks())
	fileMetaRawHash := ref.GetFileMetaHashV2()
	err = trie.Update(decodedDestHash, fileMetaRawHash, numBlocks)
	if err != nil {
		l.Logger.Error("Error updating trie", zap.Error(err))
		return err
	}
	return nil
}

func (ro *RenameOperation) GetLookupHash(changeIndex uint64) []string {
	if ro.objectTreeRefs == nil || ro.objectTreeRefs[changeIndex] == nil || ro.objectTreeRefs[changeIndex].GetType() == fileref.DIRECTORY {
		return nil
	}
	return []string{ro.destLookupHash, ro.srcLookupHash}
}
