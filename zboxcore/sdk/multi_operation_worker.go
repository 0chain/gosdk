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

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	l "github.com/0chain/gosdk/zboxcore/logger"

	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/google/uuid"
)

const (
	DefaultCreateConnectionTimeOut = 2 * time.Minute
)

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
	ctxCncl       context.CancelFunc
	operationMask zboxutil.Uint128
	maskMU        *sync.Mutex
	Consensus

	changes [][]allocationchange.AllocationChange
}

func (mo *MultiOperation) createConnectionObj(blobberIdx int) (err error) {

	defer func() {
		if err != nil {
			mo.maskMU.Lock()
			mo.operationMask = mo.operationMask.And(zboxutil.NewUint128(1).Lsh(uint64(blobberIdx)).Not())
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

			formWriter.WriteField("connection_id", mo.connectionID)
			formWriter.Close()

			var httpreq *http.Request
			httpreq, err = zboxutil.NewConnectionRequest(blobber.Baseurl, mo.allocationObj.Tx, body)
			if err != nil {
				l.Logger.Error(blobber.Baseurl, "Error creating rename request", err)
				return
			}

			httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
			ctx, cncl := context.WithTimeout(mo.ctx, DefaultCreateConnectionTimeOut)
			defer cncl()
			resp, err = zboxutil.Client.Do(httpreq.WithContext(ctx))

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
				l.Logger.Info(blobber.Baseurl, " connection obj created.")
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
	l.Logger.Info("MultiOperation Process start")
	wg := &sync.WaitGroup{}
	mo.changes = make([][]allocationchange.AllocationChange, len(mo.operations))
	ctx := mo.ctx
	ctxCncl := mo.ctxCncl
	defer ctxCncl()
	// Create connection obj in each blobber
	for blobberIdx := range mo.allocationObj.Blobbers {
		wg.Add(1)
		go func(pos int) {
			defer wg.Done()
			err := mo.createConnectionObj(pos)
			if err != nil {
				l.Logger.Error(err.Error())
			}
		}(blobberIdx)
	}
	wg.Wait()
	// Check consensus
	if mo.operationMask.CountOnes() < mo.consensusThresh {
		return errors.New("consensus_not_met",
			fmt.Sprintf("Multioperation failed. Required consensus %d got %d",
				mo.consensusThresh, mo.operationMask.CountOnes()))
	}

	errs := make(chan error, 1)

	for idx, op := range mo.operations {
		uid := util.GetNewUUID()
		wg.Add(1)
		go func(op Operationer, idx int) {
			defer wg.Done()

			// Check for other goroutines signal
			select {
			case <-ctx.Done():
				return
			default:
			}

			refs, mask, err := op.Process(mo.allocationObj, mo.connectionID) // Process with each blobber
			if err != nil {
				l.Logger.Error(err)

				select {
				case errs <- errors.New("", err.Error()):
				default:
				}
				ctxCncl()

				return
			}
			mo.maskMU.Lock()
			mo.operationMask.And(mask)
			mo.maskMU.Unlock()
			changes := op.buildChange(refs, uid)

			mo.changes[idx] = changes
		}(op, idx)
	}
	wg.Wait()
	if ctx.Err() != nil {
		return <-errs
	}
	// Check consensus
	if mo.operationMask.CountOnes() < mo.consensusThresh {
		return errors.New("consensus_not_met",
			fmt.Sprintf("Multioperation failed. Required consensus %d got %d",
				mo.consensusThresh, mo.operationMask.CountOnes()))
	}

	// Take transpose of mo.change because it will be easier to iterate mo if it contains blobber changes
	// in row instead of column. Currently mo.change[0] contains allocationChange for operation 1 and so on.
	// But we want mo.changes[0] to have allocationChange for blobber 1 and mo.changes[1] to have allocationChange for
	// blobber 2 and so on.
	mo.changes = zboxutil.Transpose(mo.changes)

	writeMarkerMutex, err := CreateWriteMarkerMutex(client.GetClient(), mo.allocationObj)
	if err != nil {
		return fmt.Errorf("Operation failed: %s", err.Error())
	}

	l.Logger.Info("Trying to lock write marker.....")
	err = writeMarkerMutex.Lock(mo.ctx, &mo.operationMask, mo.maskMU,
		mo.allocationObj.Blobbers, &mo.Consensus, 0, time.Minute, mo.connectionID)
	if err != nil {
		return fmt.Errorf("Operation failed: %s", err.Error())
	}
	l.Logger.Info("WriteMarker locked")
	defer writeMarkerMutex.Unlock(mo.ctx, mo.operationMask, mo.allocationObj.Blobbers, time.Minute, mo.connectionID) //nolint: errcheck

	mo.Consensus.Reset()
	activeBlobbers := mo.operationMask.CountOnes()
	commitReqs := make([]*CommitRequest, activeBlobbers)

	wg.Add(activeBlobbers)
	var pos uint64 = 0
	var counter = 0
	for i := mo.operationMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		commitReq := &CommitRequest{
			allocationID: mo.allocationObj.ID,
			allocationTx: mo.allocationObj.Tx,
			blobber:      mo.allocationObj.Blobbers[pos],
			connectionID: mo.connectionID,
			wg:           wg,
		}

		for _, change := range mo.changes[pos] {
			commitReq.changes = append(commitReq.changes, change)
		}
		commitReqs[counter] = commitReq
		l.Logger.Info("Commit request sending to blobber ", commitReq.blobber.Baseurl)
		go AddCommitRequest(commitReq)
		counter++
	}
	wg.Wait()

	for _, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				l.Logger.Info("Commit success", commitReq.blobber.Baseurl)
				mo.consensus++
			} else {
				l.Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			l.Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !mo.isConsensusOk() {
		err := errors.New("consensus_not_met",
			fmt.Sprintf("Commit failed. Required consensus %d, got %d",
				mo.Consensus.consensusThresh, mo.Consensus.consensus))
		if mo.getConsensus() != 0 {
			for _, op := range mo.operations {
				op.Error(mo.allocationObj, mo.getConsensus(), err)
			}
		}
		return err
	}
	for _, op := range mo.operations {
		op.Completed(mo.allocationObj)
	}

	return nil

}
