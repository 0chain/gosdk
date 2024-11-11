package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type WMLockStatus int

const (
	WMLockStatusFailed WMLockStatus = iota
	WMLockStatusPending
	WMLockStatusOK
)
const WMLockWaitTime = 2 * time.Second

type WMLockResult struct {
	Status    WMLockStatus `json:"status,omitempty"`
	CreatedAt int64        `json:"created_at,omitempty"`
}

// WriteMarkerMutex blobber WriteMarkerMutex client
type WriteMarkerMutex struct {
	mutex            sync.Mutex
	allocationObj    *Allocation
	lockedBlobbers   map[string]chan struct{}
	leadBlobberIndex int
}

// CreateWriteMarkerMutex create WriteMarkerMutex for allocation
func CreateWriteMarkerMutex(client *client.Client, allocationObj *Allocation) (*WriteMarkerMutex, error) {
	if allocationObj == nil {
		return nil, errors.Throw(constants.ErrInvalidParameter, "allocationObj")
	}

	lockedBlobbers := make(map[string]chan struct{})
	for _, b := range allocationObj.Blobbers {
		if b.ID == "" {
			logger.Logger.Error(b.Baseurl, "blobber ID is empty string")
			return nil, errors.Throw(constants.ErrInvalidParameter, "blobber ID cannot be an empty string")
		}
		lockedBlobbers[b.ID] = make(chan struct{}, 1)
	}

	return &WriteMarkerMutex{
		allocationObj:    allocationObj,
		lockedBlobbers:   lockedBlobbers,
		leadBlobberIndex: 0,
	}, nil
}

func (wmMu *WriteMarkerMutex) Unlock(
	ctx context.Context, mask zboxutil.Uint128,
	blobbers []*blockchain.StorageNode,
	timeOut time.Duration, connID string,
) {
	wg := &sync.WaitGroup{}
	var pos uint64
	for i := mask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		if pos == uint64(wmMu.leadBlobberIndex) { // Skip lead blobber
			continue
		}
		blobber := blobbers[pos]
		wg.Add(1)
		go wmMu.UnlockBlobber(ctx, blobber, connID, timeOut, wg)
	}
	wg.Wait()

	// Now unlock lead blobber
	wg.Add(1)
	go wmMu.UnlockBlobber(ctx, blobbers[uint64(wmMu.leadBlobberIndex)], connID, timeOut, wg)
	wg.Wait()
}

// Change status code to 204
func (wmMu *WriteMarkerMutex) UnlockBlobber(
	ctx context.Context, b *blockchain.StorageNode,
	connID string, timeOut time.Duration, wg *sync.WaitGroup,
) {
	defer wg.Done()
	wmMu.lockedBlobbers[b.ID] <- struct{}{}
	var err error
	defer func() {
		if err != nil {
			logger.Logger.Error(err)
		}
	}()

	var req *http.Request
	req, err = zboxutil.NewWriteMarkerUnLockRequest(
		b.Baseurl, wmMu.allocationObj.ID, wmMu.allocationObj.Tx, wmMu.allocationObj.sig, connID, "")
	if err != nil {
		return
	}

	var resp *http.Response
	var shouldContinue bool
	for retry := 0; retry < 3; retry++ {
		err, shouldContinue = func() (err error, shouldContinue bool) {
			reqCtx, ctxCncl := context.WithTimeout(ctx, timeOut)
			resp, err = zboxutil.Client.Do(req.WithContext(reqCtx))
			ctxCncl()

			if err != nil {
				return
			}
			if resp.Body != nil {
				defer resp.Body.Close()
			}
			var (
				msg  string
				data []byte
			)
			if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
				logger.Logger.Info(b.Baseurl, connID, " unlocked")
				return
			}
			if resp.StatusCode == http.StatusTooManyRequests {
				logger.Logger.Info(b.Baseurl, connID, " got too many request error. Retrying")
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

			data, err = io.ReadAll(resp.Body)
			if err != nil {
				logger.Logger.Error(err)
				return
			}

			msg = string(data)
			if msg == "EOF" {
				logger.Logger.Debug(b.Baseurl, connID, " retrying request because "+
					"server closed connection unexpectedly")
				shouldContinue = true
				return
			}

			err = errors.New("unknown_status",
				fmt.Sprintf("Blobber %s responded with status %d and message %s",
					b.Baseurl, resp.StatusCode, string(data)))

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
}

func (wmMu *WriteMarkerMutex) Lock(
	ctx context.Context, mask *zboxutil.Uint128,
	maskMu *sync.Mutex, blobbers []*blockchain.StorageNode,
	consensus *Consensus, addConsensus int, timeOut time.Duration, connID string) error {

	wmMu.mutex.Lock()
	defer wmMu.mutex.Unlock()

	consensus.Reset()
	consensus.consensus = addConsensus

	wg := &sync.WaitGroup{}

	// Lock first responsive blobber as lead blobber
	for ; wmMu.leadBlobberIndex < len(blobbers); wmMu.leadBlobberIndex++ {
		methodCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		leadBlobber := blobbers[uint64(wmMu.leadBlobberIndex)]
		wg.Add(1)
		go wmMu.lockBlobber(methodCtx, mask, maskMu, consensus, leadBlobber, uint64(wmMu.leadBlobberIndex), connID, timeOut, wg)
		wg.Wait()
		if consensus.getConsensus()-addConsensus == 1 {
			break
		}
		select {
		case <-methodCtx.Done():
			logger.Logger.Error("Locking blobber: ", leadBlobber.Baseurl, " context timeout exceeded")
			return errors.New("lock_timeout", "Locking blobber: "+leadBlobber.Baseurl+" context timeout exceeded")
		default:
		}
	}

	if consensus.getConsensus()-addConsensus != 1 {
		return errors.New("lock_consensus_not_met", "Failed to lock the lead blobber after retries")
	}

	// Once the lead blobber is locked successfully, lock the other blobbers
	var pos uint64
	for i := *mask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		if pos == uint64(wmMu.leadBlobberIndex) {
			continue
		}
		blobber := blobbers[pos]
		wg.Add(1)
		go wmMu.lockBlobber(ctx, mask, maskMu, consensus, blobber, pos, connID, timeOut, wg)
	}
	wg.Wait()
	if !consensus.isConsensusOk() {
		wmMu.Unlock(ctx, *mask, blobbers, timeOut, connID)
		return errors.New("lock_consensus_not_met",
			fmt.Sprintf("Required consensus %d got %d",
				consensus.consensusThresh, consensus.getConsensus()))
	}

	/* This goroutine will refresh lock after 30 seconds have passed. It will only complete if context is
	   completed, that is why, the caller should make proper use of context and cancel it when work is done. */
	go func() {
		for {
			<-time.NewTimer(30 * time.Second).C
			select {
			case <-ctx.Done():
				return
			default:
			}

			wg := &sync.WaitGroup{}
			cons := &Consensus{RWMutex: &sync.RWMutex{}}
			for i := *mask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
				pos = uint64(i.TrailingZeros())
				blobber := blobbers[pos]
				wg.Add(1)
				go wmMu.lockBlobber(ctx, mask, maskMu, cons, blobber, pos, connID, timeOut, wg)
			}
			wg.Wait()
		}
	}()

	return nil
}

func (wmMu *WriteMarkerMutex) lockBlobber(
	ctx context.Context, mask *zboxutil.Uint128, maskMu *sync.Mutex,
	consensus *Consensus, b *blockchain.StorageNode, pos uint64, connID string,
	timeOut time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	select {
	case <-ctx.Done():
		return
	default:
	}

	wmMu.lockedBlobbers[b.ID] <- struct{}{}
	defer func() {
		<-wmMu.lockedBlobbers[b.ID]
	}()

	var err error
	defer func() {
		if err != nil {
			logger.Logger.Error(err)
			maskMu.Lock()
			*mask = mask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
			maskMu.Unlock()
		}
	}()

	var req *http.Request
	req, err = zboxutil.NewWriteMarkerLockRequest(
		b.Baseurl, wmMu.allocationObj.ID, wmMu.allocationObj.Tx, wmMu.allocationObj.sig, connID)
	if err != nil {
		return
	}

	var resp *http.Response
	var shouldContinue bool
	for retry := 0; retry < 3; retry++ {
		select {
		case <-ctx.Done():
			return
		default:
		}
		err, shouldContinue = func() (err error, shouldContinue bool) {
			reqCtx, ctxCncl := context.WithTimeout(ctx, timeOut)
			defer ctxCncl()
			resp, err = zboxutil.Client.Do(req.WithContext(reqCtx))
			if err != nil {
				return
			}
			if resp.Body != nil {
				defer resp.Body.Close()
			}

			var data []byte
			if resp.StatusCode == http.StatusOK {
				data, err = io.ReadAll(resp.Body)
				if err != nil {
					return
				}
				wmLockRes := &WMLockResult{}
				err = json.Unmarshal(data, wmLockRes)
				if err != nil {
					return
				}
				if wmLockRes.Status == WMLockStatusOK {
					consensus.Done()
					logger.Logger.Info(b.Baseurl, connID, " locked")
					return
				}

				if wmLockRes.Status == WMLockStatusPending {
					logger.Logger.Info("Lock pending for blobber ",
						b.Baseurl, "with connection id: ", connID, " Retrying again")
					time.Sleep(WMLockWaitTime)
					shouldContinue = true
					retry--
					return
				}
				err = fmt.Errorf("Lock acquiring failed")
				return
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				logger.Logger.Info(
					b.Baseurl, connID,
					" got too many request error. Retrying")

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

			data, err = io.ReadAll(resp.Body)
			if err != nil {
				logger.Logger.Error(err)
				return
			}

			err = errors.New("unknown_status",
				fmt.Sprintf("Blobber %s responded with status %d and message %s",
					b.Baseurl, resp.StatusCode, string(data)))
			return
		}()
		if err != nil {
			return
		}
		if !shouldContinue {
			break
		}
	}
}
