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
	logger.Logger.Debug("[writemarkerLocked]", time.Since(start).Milliseconds())
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
	logger.Logger.Debug("[checkAllocStatus]", time.Since(start).Milliseconds())
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
	logger.Logger.Info("[commitRequests]", time.Since(start).Milliseconds())
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
			for _, commitReq := range commitReqs {
				if commitReq.result.Success {
					mo.allocationObj.Blobbers[commitReq.blobberInd].AllocationVersion++
				}
			}
			mo.allocationObj.allocationVersion += 1
			logger.Logger.Info("Allocation version updated to ", mo.allocationObj.allocationVersion, " activeBlobbers ", activeBlobbers)
		}
	}

	return nil

}
