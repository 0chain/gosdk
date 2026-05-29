package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/remeh/sizedwaitgroup"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	l "github.com/0chain/gosdk/zboxcore/logger"

	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/google/uuid"
)

const (
	DefaultCreateConnectionTimeOut = 45 * time.Second
)

var BatchSize = 6

type MultiOperationOption func(mo *MultiOperation)

func WithRepair(latestVersion int64, repairOffsetPath string) MultiOperationOption {
	return func(mo *MultiOperation) {
		mo.Consensus.consensusThresh = 0
		mo.isRepair = true
		mo.repairVersion = latestVersion
		mo.repairOffset = repairOffsetPath
	}
}

type Operationer interface {
	Process(allocObj *Allocation, connectionID string) ([]fileref.RefEntity, zboxutil.Uint128, error)
	buildChange(refs []fileref.RefEntity, uid uuid.UUID) []allocationchange.AllocationChange
	Verify(allocObj *Allocation) error
	Completed(allocObj *Allocation)
	Error(allocObj *Allocation, consensus int, err error)
}

type MultiOperation struct {
	connectionID  string
	operations    []Operationer
	allocationObj *Allocation
	ctx           context.Context
	ctxCncl       context.CancelCauseFunc
	operationMask zboxutil.Uint128
	maskMU        *sync.Mutex
	Consensus
	changes       [][]allocationchange.AllocationChange
	isRepair      bool
	repairVersion int64
	repairOffset  string
}

func (mo *MultiOperation) createConnectionObj(blobberIdx int) (err error) {

	defer func() {
		if err == nil {
			mo.maskMU.Lock()
			mo.operationMask = mo.operationMask.Or(zboxutil.NewUint128(1).Lsh(uint64(blobberIdx)))
			mo.maskMU.Unlock()
		}
	}()

	var (
		resp           *http.Response
		shouldContinue bool
		latestRespMsg  string

		latestStatusCode int
	)
	blobber := mo.allocationObj.Blobbers[blobberIdx]

	for i := 0; i < 3; i++ {
		err, shouldContinue = func() (err error, shouldContinue bool) {
			body := new(bytes.Buffer)
			formWriter := multipart.NewWriter(body)

			err = formWriter.WriteField("connection_id", mo.connectionID)
			if err != nil {
				return err, false
			}
			formWriter.Close()

			var httpreq *http.Request
			httpreq, err = zboxutil.NewConnectionRequest(blobber.Baseurl, mo.allocationObj.ID, mo.allocationObj.Tx, mo.allocationObj.sig, body)
			if err != nil {
				l.Logger.Error(blobber.Baseurl, "Error creating new connection request", err)
				return
			}

			httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
			ctx, cncl := context.WithTimeout(mo.ctx, DefaultCreateConnectionTimeOut)
			defer cncl()
			err = zboxutil.HttpDo(ctx, cncl, httpreq, func(r *http.Response, err error) error {
				resp = r
				return err
			})
			if err != nil {
				logger.Logger.Error("Create Connection: ", err)
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
				l.Logger.Debug(blobber.Baseurl, " connection obj created.")
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

func (mo *MultiOperation) Process() error {
	l.Logger.Debug("MultiOperation Process start")
	wg := &sync.WaitGroup{}
	mo.changes = make([][]allocationchange.AllocationChange, len(mo.operations))
	ctx := mo.ctx
	ctxCncl := mo.ctxCncl
	defer ctxCncl(nil)
	swg := sizedwaitgroup.New(BatchSize)
	errsSlice := make([]error, len(mo.operations))
	var changeCount int
	for idx, op := range mo.operations {
		swg.Add()
		go func(op Operationer, idx int) {
			defer swg.Done()

			// Check for other goroutines signal
			select {
			case <-ctx.Done():
				return
			default:
			}

			_, mask, err := op.Process(mo.allocationObj, mo.connectionID) // Process with each blobber
			if err != nil {
				if err != errFileDeleted && err != errNoChange {
					l.Logger.Error(err)
					errsSlice[idx] = errors.New("", err.Error())
					ctxCncl(err)
				}
				return
			}
			mo.maskMU.Lock()
			mo.operationMask = mo.operationMask.And(mask)
			changeCount += 1
			mo.maskMU.Unlock()
		}(op, idx)
	}
	swg.Wait()

	if ctx.Err() != nil {
		err := context.Cause(ctx)
		return err
	}

	// Check consensus
	if mo.operationMask.CountOnes() < mo.consensusThresh {
		majorErr := zboxutil.MajorError(errsSlice)
		if majorErr != nil {
			return errors.New("consensus_not_met",
				fmt.Sprintf("Multioperation failed. Required consensus %d got %d. Major error: %s",
					mo.consensusThresh, mo.operationMask.CountOnes(), majorErr.Error()))
		}
		return nil
	}

	if changeCount == 0 {
		return nil
	}

	// Take transpose of mo.change because it will be easier to iterate mo if it contains blobber changes
	// in row instead of column. Currently mo.change[0] contains allocationChange for operation 1 and so on.
	// But we want mo.changes[0] to have allocationChange for blobber 1 and mo.changes[1] to have allocationChange for
	// blobber 2 and so on.
	moOverallStart := time.Now()
	start := time.Now()

	writeMarkerMutex, err := CreateWriteMarkerMutex(client.GetClient(), mo.allocationObj)
	if err != nil {
		return fmt.Errorf("Operation failed: %s", err.Error())
	}

	l.Logger.Debug("Trying to lock write marker.....")
	if singleClientMode {
		mo.allocationObj.commitMutex.Lock()
	} else {
		err = writeMarkerMutex.Lock(mo.ctx, &mo.operationMask, mo.maskMU,
			mo.allocationObj.Blobbers, &mo.Consensus, 0, time.Minute, mo.connectionID)
		if err != nil {
			return fmt.Errorf("Operation failed: %s", err.Error())
		}
	}
	// INSTRUMENTATION (May 29): elevate writemarkerLock and checkAllocStatus
	// to Info so they appear in the gateway log. Per pprof we have ~40% CPU
	// headroom on the gateway and the per-file PUT takes ~1500 ms but
	// process() is ~650 ms and commitRequests is 11-110 ms — the missing
	// ~700 ms must be in these two network-bound phases. This tells us
	// gateway-vs-blobber where the wait actually is.
	wmLockMs := time.Since(start).Milliseconds()
	logger.Logger.Info("[writemarkerLocked]", wmLockMs)
	start = time.Now()
	status := Commit
	if !mo.isRepair && !mo.allocationObj.checkStatus {
		status, _, err = mo.allocationObj.CheckAllocStatus()
		if err != nil {
			logger.Logger.Error("Error checking allocation status", err)
			if singleClientMode {
				mo.allocationObj.commitMutex.Unlock()
			} else {
				writeMarkerMutex.Unlock(mo.ctx, mo.operationMask, mo.allocationObj.Blobbers, time.Minute, mo.connectionID) //nolint: errcheck
			}
			return fmt.Errorf("Check allocation status failed: %s", err.Error())
		}
		if status == Repair {
			if singleClientMode {
				mo.allocationObj.commitMutex.Unlock()
			} else {
				writeMarkerMutex.Unlock(mo.ctx, mo.operationMask, mo.allocationObj.Blobbers, time.Minute, mo.connectionID) //nolint: errcheck
			}
			for _, op := range mo.operations {
				op.Error(mo.allocationObj, 0, ErrRepairRequired)
			}
			return ErrRepairRequired
		}
	}
	if singleClientMode {
		mo.allocationObj.checkStatus = true
		defer mo.allocationObj.commitMutex.Unlock()
	} else {
		defer writeMarkerMutex.Unlock(mo.ctx, mo.operationMask, mo.allocationObj.Blobbers, time.Minute, mo.connectionID) //nolint: errcheck
	}
	if status != Commit {
		for _, op := range mo.operations {
			op.Error(mo.allocationObj, 0, ErrRetryOperation)
		}
		return ErrRetryOperation
	}
	checkAllocMs := time.Since(start).Milliseconds()
	logger.Logger.Info("[checkAllocStatus]", checkAllocMs)
	mo.Consensus.Reset()
	var pos uint64
	if !mo.isRepair {
		for i := mo.operationMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
			pos = uint64(i.TrailingZeros())
			if mo.allocationObj.Blobbers[pos].AllocationVersion != mo.allocationObj.allocationVersion {
				mo.operationMask = mo.operationMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
			}
		}
	}
	activeBlobbers := mo.operationMask.CountOnes()
	if activeBlobbers < mo.consensusThresh {
		return errors.New("consensus_not_met", "Active blobbers less than consensus threshold")
	}
	commitReqs := make([]*CommitRequest, activeBlobbers)
	start = time.Now()
	wg.Add(activeBlobbers)

	var counter = 0
	timestamp := int64(common.Now())
	for i := mo.operationMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		commitReq := &CommitRequest{
			allocationID: mo.allocationObj.ID,
			allocationTx: mo.allocationObj.Tx,
			sig:          mo.allocationObj.sig,
			blobber:      mo.allocationObj.Blobbers[pos],
			connectionID: mo.connectionID,
			wg:           wg,
			timestamp:    timestamp,
			blobberInd:   pos,
			version:      mo.allocationObj.Blobbers[pos].AllocationVersion + 1,
		}
		if mo.isRepair {
			commitReq.isRepair = true
			commitReq.version = mo.allocationObj.Blobbers[pos].AllocationVersion
			commitReq.repairVersion = mo.repairVersion
			commitReq.repairOffset = mo.repairOffset
		}
		commitReqs[counter] = commitReq
		l.Logger.Debug("Commit request sending to blobber ", commitReq.blobber.Baseurl)
		go AddCommitRequest(commitReq)
		counter++
	}
	wg.Wait()
	commitMs := time.Since(start).Milliseconds()
	logger.Logger.Info("[commitRequests]", commitMs)
	// One-line summary so it's easy to see relative weights per PUT:
	// process() + the three commit-phase pieces. process() itself is
	// logged separately via [batch-timing] / [upload-finalize].
	logger.Logger.Info(fmt.Sprintf("[mo-phase-summary] wmLock=%dms checkAlloc=%dms commitReqs=%dms commit_phase_total=%dms",
		wmLockMs, checkAllocMs, commitMs, time.Since(moOverallStart).Milliseconds()))
	rollbackMask := zboxutil.NewUint128(0)
	errSlice := make([]error, len(commitReqs))
	for idx, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				l.Logger.Debug("Commit success", commitReq.blobber.Baseurl)
				if !mo.isRepair {
					rollbackMask = rollbackMask.Or(zboxutil.NewUint128(1).Lsh(commitReq.blobberInd))
				}
				mo.consensus++
			} else {
				errSlice[idx] = errors.New("commit_failed", commitReq.result.ErrorMessage)
				l.Logger.Error("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			l.Logger.Debug("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !mo.isConsensusOk() {
		err = zboxutil.MajorError(errSlice)
		if mo.getConsensus() != 0 {
			l.Logger.Info("Rolling back changes on minority blobbers")
			mo.allocationObj.RollbackWithMask(rollbackMask)
		}
		for _, op := range mo.operations {
			op.Error(mo.allocationObj, mo.getConsensus(), err)
		}
		return err
	} else {
		for _, op := range mo.operations {
			op.Completed(mo.allocationObj)
		}
		if singleClientMode && !mo.isRepair {
			var anyFailed bool
			for _, commitReq := range commitReqs {
				if commitReq.result.Success {
					mo.allocationObj.Blobbers[commitReq.blobberInd].AllocationVersion++
				} else {
					anyFailed = true
				}
			}
			mo.allocationObj.allocationVersion += 1
			logger.Logger.Info("Allocation version updated to ", mo.allocationObj.allocationVersion, " activeBlobbers ", activeBlobbers)
			if anyFailed {
				// Force fresh CheckAllocStatus before next write so the
				// AllocationVersion filter uses real blobber state, not stale
				// local state where failed blobbers appear permanently behind.
				mo.allocationObj.checkStatus = false
			}
		}
	}

	// Read-after-write visibility barrier. Porcupine 2026-04-21 surfaced the
	// race where a COPY/MOVE/RENAME ack (5/5 WriteMarker commits) precedes
	// the per-blobber ref-index materialization on the slower replicas, so an
	// immediate GET of the destination returns ErrNotFound from those
	// replicas, trips the read-side staleness barrier, and returns empty.
	// Wait (bounded) until every committed blobber reports the dst ref as
	// visible before returning success to the caller.
	if !mo.isRepair {
		if dests := collectDestPathsForPoll(mo.operations); len(dests) > 0 {
			mo.verifyReadAfterWrite(dests)
		}
	}

	return nil

}

// collectDestPathsForPoll returns the full destination paths of every
// mutation op (Copy/Move/Rename) in this batch. These are the paths whose
// read-visibility we must confirm before reporting the multi-op success.
func collectDestPathsForPoll(ops []Operationer) []string {
	var dests []string
	for _, op := range ops {
		switch o := op.(type) {
		case *CopyOperation:
			if o.destPath != "" && o.remotefilepath != "" {
				dests = append(dests, path.Join(o.destPath, path.Base(o.remotefilepath)))
			}
		case *MoveOperation:
			if o.destPath != "" && o.remotefilepath != "" {
				dests = append(dests, path.Join(o.destPath, path.Base(o.remotefilepath)))
			}
		case *RenameOperation:
			if o.newName != "" && o.remotefilepath != "" {
				dests = append(dests, path.Join(path.Dir(o.remotefilepath), o.newName))
			}
		}
	}
	return dests
}

// verifyReadAfterWrite polls each committed blobber for the given dst paths
// until all blobbers in the operation mask report a non-nil fileref for
// every dst, or the deadline expires. Bounded at 5 s (250 × 20 ms retries).
// On timeout, logs a warning but returns normally — the read-side staleness
// barrier reorder in downloadworker.go should also avoid false ErrNotFound
// on the majority of legitimate consensus races.
func (mo *MultiOperation) verifyReadAfterWrite(dests []string) {
	if len(dests) == 0 {
		return
	}
	deadline := time.Now().Add(5 * time.Second)
	var positions []uint64
	{
		var pos uint64
		for i := mo.operationMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
			pos = uint64(i.TrailingZeros())
			positions = append(positions, pos)
		}
	}
	if len(positions) == 0 {
		return
	}

	for time.Now().Before(deadline) {
		allVisible := true
	checkDst:
		for _, dst := range dests {
			for _, p := range positions {
				listReq := &ListRequest{
					remotefilepath: dst,
					allocationID:   mo.allocationObj.ID,
					allocationTx:   mo.allocationObj.Tx,
					sig:            mo.allocationObj.sig,
					blobbers:       mo.allocationObj.Blobbers,
					ctx:            mo.ctx,
					Consensus: Consensus{
						RWMutex:         &sync.RWMutex{},
						fullconsensus:   mo.fullconsensus,
						consensusThresh: mo.consensusThresh,
					},
				}
				rsp := make(chan *fileMetaResponse, 1)
				go listReq.getFileMetaInfoFromBlobber(mo.allocationObj.Blobbers[p], int(p), rsp)
				r := <-rsp
				if r == nil || r.fileref == nil || r.err != nil {
					allVisible = false
					break checkDst
				}
			}
		}
		if allVisible {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	logger.Logger.Info("verifyReadAfterWrite timed out before all dests visible", "dests", dests)
}
