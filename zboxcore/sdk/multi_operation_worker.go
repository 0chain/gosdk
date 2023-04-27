package sdk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/0chain/errors"
	// "github.com/0chain/gosdk/constants"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"

	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/google/uuid"
)

type Operationer interface {
	Process(allocObj *Allocation, connectionID string) ([]fileref.RefEntity, zboxutil.Uint128, error)
	buildChange(refs []fileref.RefEntity, uid uuid.UUID) []allocationchange.AllocationChange
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

func (mo *MultiOperation) Process() error {
	l.Logger.Info("MultiOperation Process start")
	wg := &sync.WaitGroup{}
	mo.changes = make([][]allocationchange.AllocationChange, len(mo.operations))
	ctx := mo.allocationObj.ctx
	ctxCncl := mo.allocationObj.ctxCancelF
	defer ctxCncl()
	errs := make(chan error, 1)
	uid := util.GetNewUUID()
	for idx, op := range mo.operations {
		// Don't use goroutine for the first operation because in blobber code we try to fetch the allocation
		// from the postgress and sharders, if not found blobber try to create it. This is done without lock, so if we
		// sent multiple goroutine together, blobber will try to create multiple allocations for same allocation id
		// and eventually throw error.
		if idx == 0 {
			refs, mask, err := op.Process(mo.allocationObj, mo.connectionID) // Process with each blobber
			mo.operationMask.And(mask)
			if err != nil {
				return err
			}
			changes := op.buildChange(refs, uid)
			mo.changes[idx] = changes
			continue
		}
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
			mo.operationMask.And(mask)
			changes := op.buildChange(refs, uid)

			mo.changes[idx] = changes
		}(op, idx)
	}
	wg.Wait()
	if ctx.Err() != nil {
		return <-errs
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
