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
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/google/uuid"
)

const (
	DirectoryExists = "directory_exists"
)

type DirRequest struct {
	allocationObj *Allocation
	allocationID  string
	allocationTx  string
	sig           string
	remotePath    string
	blobbers      []*blockchain.StorageNode
	ctx           context.Context
	ctxCncl       context.CancelFunc
	wg            *sync.WaitGroup
	dirMask       zboxutil.Uint128
	mu            *sync.Mutex
	connectionID  string
	timestamp     int64
	alreadyExists map[uint64]bool
	customMeta    string
	Consensus
}

func (req *DirRequest) ProcessWithBlobbers(a *Allocation) int {
	var pos uint64
	var existingDirCount int
	countMu := &sync.Mutex{}
	for i := req.dirMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		req.wg.Add(1)
		go func(pos uint64) {
			defer req.wg.Done()

			err, alreadyExists := req.createDirInBlobber(a.Blobbers[pos], pos)
			if err != nil {
				l.Logger.Error(err.Error())
				return
			}
			if alreadyExists {
				countMu.Lock()
				req.alreadyExists[pos] = true
				existingDirCount++
				countMu.Unlock()
			}
		}(pos)
	}

	req.wg.Wait()
	return existingDirCount
}

func (req *DirRequest) ProcessDir(a *Allocation) error {
	l.Logger.Info("Start creating dir for blobbers")

	defer req.ctxCncl()
	existingDirCount := req.ProcessWithBlobbers(a)
	if !req.isConsensusOk() {
		return errors.New("consensus_not_met", "directory creation failed due to consensus not met")
	}

	writeMarkerMU, err := CreateWriteMarkerMutex(client.GetClient(), a)
	if err != nil {
		return fmt.Errorf("directory creation failed. Err: %s", err.Error())
	}
	err = writeMarkerMU.Lock(
		req.ctx, &req.dirMask, req.mu,
		req.blobbers, &req.Consensus, existingDirCount, time.Minute, req.connectionID)
	if err != nil {
		return fmt.Errorf("directory creation failed. Err: %s", err.Error())
	}
	defer writeMarkerMU.Unlock(req.ctx, req.dirMask,
		a.Blobbers, time.Minute, req.connectionID) //nolint: errcheck

	return req.commitRequest(existingDirCount)
}

func (req *DirRequest) commitRequest(existingDirCount int) error {
	req.Consensus.Reset()
	req.timestamp = int64(common.Now())
	req.consensus = existingDirCount
	wg := &sync.WaitGroup{}
	activeBlobbersNum := req.dirMask.CountOnes()
	wg.Add(activeBlobbersNum)

	commitReqs := make([]*CommitRequest, activeBlobbersNum)
	var pos uint64
	var c int

	uid := util.GetNewUUID()

	for i := req.dirMask; !i.Equals(zboxutil.NewUint128(0)); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		commitReq := &CommitRequest{}
		commitReq.allocationID = req.allocationID
		commitReq.allocationTx = req.allocationTx
		commitReq.blobber = req.blobbers[pos]
		commitReq.sig = req.sig
		newChange := &allocationchange.DirCreateChange{
			RemotePath: req.remotePath,
			Uuid:       uid,
			Timestamp:  common.Timestamp(req.timestamp),
		}

		commitReq.changes = append(commitReq.changes, newChange)
		commitReq.connectionID = req.connectionID
		commitReq.wg = wg
		commitReq.timestamp = req.timestamp
		commitReqs[c] = commitReq
		c++
		go AddCommitRequest(commitReq)
	}
	wg.Wait()

	for _, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				l.Logger.Info("Commit success", commitReq.blobber.Baseurl)
				req.Consensus.Done()
			} else {
				l.Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			l.Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !req.isConsensusOk() {
		return errors.New("consensus_not_met", "directory creation failed due consensus not met")
	}
	return nil
}

func (req *DirRequest) createDirInBlobber(blobber *blockchain.StorageNode, pos uint64) (err error, alreadyExists bool) {
	defer func() {
		if err != nil {
			req.mu.Lock()
			req.dirMask = req.dirMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
			req.mu.Unlock()
		}
	}()

	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	err = formWriter.WriteField("connection_id", req.connectionID)
	if err != nil {
		return err, false
	}

	err = formWriter.WriteField("dir_path", req.remotePath)
	if err != nil {
		return err, false
	}

	if req.customMeta != "" {
		err = formWriter.WriteField("custom_meta", req.customMeta)
		if err != nil {
			return err, false
		}
	}

	formWriter.Close()
	httpreq, err := zboxutil.NewCreateDirRequest(blobber.Baseurl, req.allocationID, req.allocationTx, req.sig, body)
	if err != nil {
		l.Logger.Error(blobber.Baseurl, "Error creating dir request", err)
		return err, false
	}

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())

	var (
		resp             *http.Response
		shouldContinue   bool
		latestRespMsg    string
		latestStatusCode int
	)

	for i := 0; i < 3; i++ {
		err, shouldContinue = func() (err error, shouldContinue bool) {
			ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 10))
			resp, err = zboxutil.Client.Do(httpreq.WithContext(ctx))
			cncl()
			if err != nil {
				return err, false
			}

			if resp.Body != nil {
				defer resp.Body.Close()
			}

			var (
				respBody []byte
				msg      string
			)
			if resp.StatusCode == http.StatusOK {
				l.Logger.Info("Successfully created directory ", req.remotePath)
				req.Consensus.Done()
				return
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				var r int
				r, err = zboxutil.GetRateLimitValue(resp)
				if err != nil {
					return
				}
				l.Logger.Debug(fmt.Sprintf("Got too many request error. Retrying after %d seconds", r))
				time.Sleep(time.Duration(r) * time.Second)
				shouldContinue = true
				return
			}

			respBody, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				l.Logger.Error(err)
				return
			}

			latestRespMsg = string(respBody)
			latestStatusCode = resp.StatusCode

			msg = string(respBody)
			if strings.Contains(msg, DirectoryExists) {
				req.Consensus.Done()
				alreadyExists = true
				return
			}
			l.Logger.Error(blobber.Baseurl, " Response: ", msg)

			err = errors.New("response_error", msg)
			return
		}()

		if err != nil {
			logger.Logger.Error(err)
			return
		}
		if shouldContinue {
			continue
		}
		return

	}

	return errors.New("dir_creation_failed",
		fmt.Sprintf("Directory creation failed with latest status: %d and "+
			"latest message: %s", latestStatusCode, latestRespMsg)), false
}

type DirOperation struct {
	remotePath    string
	ctx           context.Context
	ctxCncl       context.CancelFunc
	dirMask       zboxutil.Uint128
	maskMU        *sync.Mutex
	customMeta    string
	alreadyExists map[uint64]bool

	Consensus
}

func (dirOp *DirOperation) Process(allocObj *Allocation, connectionID string) ([]fileref.RefEntity, zboxutil.Uint128, error) {
	refs := make([]fileref.RefEntity, len(allocObj.Blobbers))
	dR := &DirRequest{
		allocationID:  allocObj.ID,
		allocationTx:  allocObj.Tx,
		connectionID:  connectionID,
		sig:           allocObj.sig,
		blobbers:      allocObj.Blobbers,
		remotePath:    dirOp.remotePath,
		ctx:           dirOp.ctx,
		ctxCncl:       dirOp.ctxCncl,
		dirMask:       dirOp.dirMask,
		mu:            dirOp.maskMU,
		wg:            &sync.WaitGroup{},
		alreadyExists: make(map[uint64]bool),
		customMeta:    dirOp.customMeta,
	}
	dR.Consensus = Consensus{
		RWMutex:         &sync.RWMutex{},
		consensusThresh: dR.consensusThresh,
		fullconsensus:   dR.fullconsensus,
	}

	_ = dR.ProcessWithBlobbers(allocObj)
	dirOp.alreadyExists = dR.alreadyExists

	if !dR.isConsensusOk() {
		return nil, dR.dirMask, errors.New("consensus_not_met", "directory creation failed due to consensus not met")
	}
	return refs, dR.dirMask, nil

}

func (dirOp *DirOperation) buildChange(refs []fileref.RefEntity, uid uuid.UUID) []allocationchange.AllocationChange {

	var pos uint64
	changes := make([]allocationchange.AllocationChange, len(refs))
	for i := dirOp.dirMask; !i.Equals(zboxutil.NewUint128(0)); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		if dirOp.alreadyExists[pos] {
			newChange := &allocationchange.EmptyFileChange{}
			changes[pos] = newChange
		} else {
			newChange := &allocationchange.DirCreateChange{
				RemotePath: dirOp.remotePath,
				Uuid:       uid,
				Timestamp:  common.Now(),
			}
			changes[pos] = newChange
		}
	}
	return changes
}

func (dirOp *DirOperation) Verify(a *Allocation) error {
	if dirOp.remotePath == "" {
		return errors.New("invalid_name", "Invalid name for dir")
	}

	if !path.IsAbs(dirOp.remotePath) {
		return errors.New("invalid_path", "Path is not absolute")
	}
	return nil
}

func (dirOp *DirOperation) Completed(allocObj *Allocation) {

}

func (dirOp *DirOperation) Error(allocObj *Allocation, consensus int, err error) {

}

func NewDirOperation(remotePath, customMeta string, dirMask zboxutil.Uint128, maskMU *sync.Mutex, consensusTh int, fullConsensus int, ctx context.Context) *DirOperation {
	dirOp := &DirOperation{}
	dirOp.remotePath = zboxutil.RemoteClean(remotePath)
	dirOp.dirMask = dirMask
	dirOp.maskMU = maskMU
	dirOp.consensusThresh = consensusTh
	dirOp.fullconsensus = fullConsensus
	dirOp.customMeta = customMeta
	dirOp.ctx, dirOp.ctxCncl = context.WithCancel(ctx)
	dirOp.alreadyExists = make(map[uint64]bool)
	return dirOp
}
