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
	Process(allocObj *Allocation, connectionID string) ([]fileref.RefEntity, error)
	buildChange(refs []fileref.RefEntity, uid uuid.UUID) []allocationchange.AllocationChange
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
	ctx, ctxCncl := context.WithCancel(context.Background())
	defer ctxCncl()
	errs := make(chan error, 1)
	uid := util.GetNewUUID()
	for idx, op := range mo.operations {
		wg.Add(1)
		go func(op Operationer, idx int) {
			defer wg.Done()

			// Check for other goroutines signal
			select {
			case <-ctx.Done():
				return
			default:
			}

			refs, err := op.Process(mo.allocationObj, mo.connectionID) // Process with each blobber

			if err != nil {
				l.Logger.Error(err)

				select {
				case errs <- errors.New("", err.Error()):
				default:
				}
				ctxCncl()

				return
			}
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
	var cntr = 0
	for i := mo.operationMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		commitReq := &CommitRequest{
			allocationID: mo.allocationObj.ID,
			allocationTx: mo.allocationObj.Tx,
			blobber:      mo.allocationObj.Blobbers[pos],
			connectionID: mo.connectionID,
			wg:           wg,
		}
		// Check here if mo.changes[pos] is available
		for _, change := range mo.changes[pos] {
			commitReq.changes = append(commitReq.changes, change)
		}
		commitReqs[cntr] = commitReq
		l.Logger.Info("Commit request sending to blobber ", commitReq.blobber.Baseurl)
		go AddCommitRequest(commitReq)
		cntr++
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
		return errors.New("consensus_not_met",
			fmt.Sprintf("Commit failed. Required consensus %d, got %d",
				mo.Consensus.consensusThresh, mo.Consensus.consensus))
	}

	return nil

}
