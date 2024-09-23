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
	"strings"
	"sync"
	"time"

	"github.com/0chain/common/core/util/wmpt"
	"github.com/0chain/errors"
	thrown "github.com/0chain/errors"
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

type MoveRequest struct {
	allocationObj  *Allocation
	allocationID   string
	allocationTx   string
	sig            string
	blobbers       []*blockchain.StorageNode
	remotefilepath string
	destPath       string
	ctx            context.Context
	ctxCncl        context.CancelFunc
	moveMask       zboxutil.Uint128
	maskMU         *sync.Mutex
	connectionID   string
	timestamp      int64
	destLookupHash string
	Consensus
}

func (req *MoveRequest) getObjectTreeFromBlobber(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	return getObjectTreeFromBlobber(req.ctx, req.allocationID, req.allocationTx, req.sig, req.remotefilepath, blobber)
}

func (req *MoveRequest) getFileMetaFromBlobber(pos int) (fileRef *fileref.FileRef, err error) {
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

func (req *MoveRequest) moveBlobberObject(
	blobber *blockchain.StorageNode, blobberIdx int, fetchObjectTree bool) (refEntity fileref.RefEntity, err error) {

	defer func() {
		if err != nil {
			req.maskMU.Lock()
			// Removing blobber from mask
			req.moveMask = req.moveMask.And(zboxutil.NewUint128(1).Lsh(uint64(blobberIdx)).Not())
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

			formWriter.Close()

			var (
				httpreq  *http.Request
				respBody []byte
				ctx      context.Context
				cncl     context.CancelFunc
			)

			httpreq, err = zboxutil.NewMoveRequest(blobber.Baseurl, req.allocationID, req.allocationTx, req.sig, body)
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
				logger.Logger.Error("Move: ", err)
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
				l.Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " moved.")
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

func (req *MoveRequest) ProcessWithBlobbers() ([]fileref.RefEntity, error) {
	var pos uint64
	numList := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, numList)
	blobberErrors := make([]error, numList)
	wg := &sync.WaitGroup{}
	for i := req.moveMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		wg.Add(1)
		go func(blobberIdx int) {
			defer wg.Done()
			refEntity, err := req.moveBlobberObject(req.blobbers[blobberIdx], blobberIdx, true)
			if err != nil {
				blobberErrors[blobberIdx] = err
				l.Logger.Error(err.Error())
				return
			}
			objectTreeRefs[blobberIdx] = refEntity
		}(int(pos))
	}
	wg.Wait()
	return objectTreeRefs, zboxutil.MajorError(blobberErrors)
}

func (req *MoveRequest) ProcessWithBlobbersV2() ([]fileref.RefEntity, error) {

	var (
		pos          uint64
		consensusRef *fileref.FileRef
	)
	numList := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, numList)
	blobberErrors := make([]error, numList)
	versionMap := make(map[string]int)
	wg := &sync.WaitGroup{}
	for i := req.moveMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		wg.Add(1)
		go func(blobberIdx int) {
			defer wg.Done()
			refEntity, err := req.getFileMetaFromBlobber(blobberIdx)
			if err != nil {
				blobberErrors[blobberIdx] = err
				l.Logger.Debug(err.Error())
				return
			}
			refEntity.Path = path.Join(req.destPath, path.Base(refEntity.Path))
			objectTreeRefs[blobberIdx] = refEntity
			req.maskMU.Lock()
			versionMap[refEntity.AllocationRoot] += 1
			if versionMap[refEntity.AllocationRoot] >= req.consensusThresh {
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
			if refEntity.GetAllocationRoot() != consensusRef.AllocationRoot {
				req.moveMask = req.moveMask.And(zboxutil.NewUint128(1).Lsh(uint64(ind)).Not())
			}
		}
		subRequest := &subDirRequest{
			allocationObj:   req.allocationObj,
			remotefilepath:  req.remotefilepath,
			destPath:        req.destPath,
			ctx:             req.ctx,
			consensusThresh: req.consensusThresh,
			opType:          constants.FileOperationMove,
			subOpType:       constants.FileOperationMove,
			mask:            req.moveMask,
		}
		err := subRequest.processSubDirectories()
		if err != nil {
			return nil, err
		}
		op := OperationRequest{
			OperationType: constants.FileOperationDelete,
			RemotePath:    req.remotefilepath,
		}
		err = req.allocationObj.DoMultiOperation([]OperationRequest{op})
		if err != nil {
			return nil, err
		}
		req.consensus = req.moveMask.CountOnes()
		return nil, errNoChange
	}

	for i := req.moveMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		wg.Add(1)
		go func(blobberIdx int) {
			defer wg.Done()
			_, err := req.moveBlobberObject(req.blobbers[blobberIdx], blobberIdx, false)
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
	req.destLookupHash = fileref.GetReferenceLookup(req.allocationID, consensusRef.Path)
	return objectTreeRefs, err
}

func (req *MoveRequest) ProcessMove() error {
	defer req.ctxCncl()

	wg := &sync.WaitGroup{}
	var pos uint64

	objectTreeRefs, err := req.ProcessWithBlobbers()

	if !req.isConsensusOk() {
		if err != nil {
			return errors.New("move_failed", fmt.Sprintf("Move failed. %s", err.Error()))
		}

		return errors.New("consensus_not_met",
			fmt.Sprintf("Move failed. Required consensus %d, got %d",
				req.Consensus.consensusThresh, req.Consensus.consensus))
	}

	writeMarkerMutex, err := CreateWriteMarkerMutex(client.GetClient(), req.allocationObj)
	if err != nil {
		return fmt.Errorf("Move failed: %s", err.Error())
	}
	err = writeMarkerMutex.Lock(req.ctx, &req.moveMask, req.maskMU,
		req.blobbers, &req.Consensus, 0, time.Minute, req.connectionID)
	if err != nil {
		return fmt.Errorf("Move failed: %s", err.Error())
	}

	//Check if the allocation is to be repaired or rolled back
	status, _, err := req.allocationObj.CheckAllocStatus()
	if err != nil {
		logger.Logger.Error("Error checking allocation status: ", err)
		return fmt.Errorf("Move failed: %s", err.Error())
	}
	defer writeMarkerMutex.Unlock(req.ctx, req.moveMask, req.blobbers, time.Minute, req.connectionID) //nolint: errcheck

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

	req.Consensus.Reset()
	req.timestamp = int64(common.Now())
	activeBlobbers := req.moveMask.CountOnes()
	wg.Add(activeBlobbers)
	commitReqs := make([]*CommitRequest, activeBlobbers)
	var c int
	for i := req.moveMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		moveChange := &allocationchange.MoveFileChange{
			DestPath:   req.destPath,
			ObjectTree: objectTreeRefs[pos],
		}
		moveChange.NumBlocks = 0
		moveChange.Operation = constants.FileOperationMove
		moveChange.Size = 0
		commitReq := &CommitRequest{
			allocationID: req.allocationID,
			allocationTx: req.allocationTx,
			sig:          req.sig,
			blobber:      req.blobbers[pos],
			connectionID: req.connectionID,
			wg:           wg,
			timestamp:    req.timestamp,
		}
		// commitReq.change = moveChange
		commitReq.changes = append(commitReq.changes, moveChange)
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
			fmt.Sprintf("Commit on move failed. Required consensus %d, got %d",
				req.Consensus.consensusThresh, req.Consensus.consensus))
	}
	return nil
}

type MoveOperation struct {
	remotefilepath string
	destPath       string
	srcLookupHash  string
	destLookupHash string
	ctx            context.Context
	ctxCncl        context.CancelFunc
	moveMask       zboxutil.Uint128
	maskMU         *sync.Mutex
	consensus      Consensus
	objectTreeRefs []fileref.RefEntity
}

func (mo *MoveOperation) Process(allocObj *Allocation, connectionID string) ([]fileref.RefEntity, zboxutil.Uint128, error) {
	mR := &MoveRequest{
		allocationObj:  allocObj,
		allocationID:   allocObj.ID,
		allocationTx:   allocObj.Tx,
		sig:            allocObj.sig,
		connectionID:   connectionID,
		blobbers:       allocObj.Blobbers,
		remotefilepath: mo.remotefilepath,
		ctx:            mo.ctx,
		ctxCncl:        mo.ctxCncl,
		moveMask:       mo.moveMask,
		maskMU:         mo.maskMU,
		destPath:       mo.destPath,
		Consensus:      Consensus{RWMutex: &sync.RWMutex{}},
	}
	mR.Consensus.fullconsensus = mo.consensus.fullconsensus
	mR.Consensus.consensusThresh = mo.consensus.consensusThresh
	var err error
	if allocObj.StorageVersion == StorageV2 {
		mo.objectTreeRefs, err = mR.ProcessWithBlobbersV2()
	} else {
		mo.objectTreeRefs, err = mR.ProcessWithBlobbers()
	}

	if !mR.Consensus.isConsensusOk() {
		if err != nil {
			if err == errNoChange {
				return nil, mR.moveMask, err
			}
			return nil, mR.moveMask, thrown.New("move_failed", fmt.Sprintf("Move failed. %s", err.Error()))
		}

		return nil, mR.moveMask, thrown.New("consensus_not_met",
			fmt.Sprintf("Move failed. Required consensus %d, got %d",
				mR.Consensus.consensusThresh, mR.Consensus.consensus))
	}
	mo.destLookupHash = mR.destLookupHash
	mo.srcLookupHash = fileref.GetReferenceLookup(mR.allocationID, mR.remotefilepath)
	return mo.objectTreeRefs, mR.moveMask, nil
}

func (mo *MoveOperation) buildChange(refs []fileref.RefEntity, uid uuid.UUID) []allocationchange.AllocationChange {

	changes := make([]allocationchange.AllocationChange, len(refs))
	for idx, ref := range refs {
		if ref == nil {
			change := &allocationchange.EmptyFileChange{}
			changes[idx] = change
			continue
		}
		moveChange := &allocationchange.MoveFileChange{
			DestPath:   mo.destPath,
			ObjectTree: ref,
		}
		moveChange.NumBlocks = 0
		moveChange.Operation = constants.FileOperationMove
		moveChange.Size = 0
		moveChange.Uuid = uid
		changes[idx] = moveChange
	}
	return changes
}

func (mo *MoveOperation) Verify(a *Allocation) error {

	if !a.CanMove() {
		return constants.ErrFileOptionNotPermitted
	}

	if mo.remotefilepath == "" || mo.destPath == "" {
		return errors.New("invalid_path", "Invalid path for move")
	}
	isabs := zboxutil.IsRemoteAbs(mo.remotefilepath)
	if !isabs {
		return errors.New("invalid_path", "Path should be valid and absolute")
	}

	err := ValidateRemoteFileName(mo.destPath)

	if err != nil {
		return err
	}
	return nil
}

func (mo *MoveOperation) Completed(allocObj *Allocation) {

}

func (mo *MoveOperation) Error(allocObj *Allocation, consensus int, err error) {

}

func NewMoveOperation(remotePath string, destPath string, moveMask zboxutil.Uint128, maskMU *sync.Mutex, consensusTh int, fullConsensus int, ctx context.Context) *MoveOperation {
	mo := &MoveOperation{}
	mo.remotefilepath = zboxutil.RemoteClean(remotePath)
	if destPath != "/" {
		destPath = strings.TrimSuffix(destPath, "/")
	}
	mo.destPath = destPath
	mo.moveMask = moveMask
	mo.maskMU = maskMU
	mo.consensus.consensusThresh = consensusTh
	mo.consensus.fullconsensus = fullConsensus
	mo.ctx, mo.ctxCncl = context.WithCancel(ctx)
	return mo
}

func (mo *MoveOperation) ProcessChangeV2(trie *wmpt.WeightedMerkleTrie, changeIndex uint64) error {
	if mo.objectTreeRefs == nil || mo.objectTreeRefs[changeIndex] == nil || mo.objectTreeRefs[changeIndex].GetType() == fileref.DIRECTORY {
		return nil
	}
	decodedSrcHash, _ := hex.DecodeString(mo.srcLookupHash)
	err := trie.Update(decodedSrcHash, nil, 0)
	if err != nil {
		l.Logger.Error("Error updating trie", zap.Error(err))
		return err
	}
	decodedDestHash, _ := hex.DecodeString(mo.destLookupHash)
	ref := mo.objectTreeRefs[changeIndex]
	numBlocks := uint64(ref.GetNumBlocks())
	fileMetaRawHash := ref.GetFileMetaHashV2()
	err = trie.Update(decodedDestHash, fileMetaRawHash, numBlocks)
	if err != nil {
		l.Logger.Error("Error updating trie", zap.Error(err))
		return err
	}
	return nil
}

func (mo *MoveOperation) GetLookupHash(changeIndex uint64) []string {
	if mo.objectTreeRefs == nil || mo.objectTreeRefs[changeIndex] == nil || mo.objectTreeRefs[changeIndex].GetType() == fileref.DIRECTORY {
		return nil
	}
	return []string{mo.destLookupHash, mo.srcLookupHash}
}
