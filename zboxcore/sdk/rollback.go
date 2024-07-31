package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"net/http"

	"errors"

	"github.com/0chain/common/core/common"
	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"go.uber.org/zap"
)

type LatestPrevWriteMarker struct {
	LatestWM *marker.WriteMarker `json:"latest_write_marker"`
	PrevWM   *marker.WriteMarker `json:"prev_write_marker"`
	Version  string              `json:"version"`
}

type LatestVersionMarker struct {
	VersionMarker *marker.VersionMarker `json:"version_marker"`
}

type AllocStatus byte

const (
	Commit AllocStatus = iota
	Repair
	Broken
	Rollback
)

var (
	ErrRetryOperation = errors.New("retry_operation")
	ErrRepairRequired = errors.New("repair_required")
)

type RollbackBlobber struct {
	blobber      *blockchain.StorageNode
	commitResult *CommitResult
	lvm          *LatestVersionMarker
	blobIndex    int
}

type BlobberStatus struct {
	ID     string
	Status string
}

func GetWritemarker(allocID, allocTx, id, baseUrl string) (*LatestVersionMarker, error) {

	var lvm LatestVersionMarker

	req, err := zboxutil.NewWritemarkerRequest(baseUrl, allocID, allocTx)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for retries := 0; retries < 3; retries++ {

		resp, err := zboxutil.Client.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			l.Logger.Info(baseUrl, "got too many requests, retrying")
			var r int
			r, err = zboxutil.GetRateLimitValue(resp)
			if err != nil {
				l.Logger.Error(err)
				return nil, err
			}
			time.Sleep(time.Duration(r) * time.Second)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("writemarker error response %s with status %d", body, resp.StatusCode)
		}

		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, &lvm)
		if err != nil {
			return nil, err
		}
		if lvm.VersionMarker != nil && lvm.VersionMarker.Version != 0 {
			err = lvm.VersionMarker.VerifySignature(client.GetClientPublicKey())
			if err != nil {
				return nil, fmt.Errorf("signature verification failed for latest writemarker: %s", err.Error())
			}
		}
		return &lvm, nil
	}

	return nil, fmt.Errorf("writemarker error response %d", http.StatusTooManyRequests)
}

func (rb *RollbackBlobber) processRollback(ctx context.Context, tx string) error {
	// don't rollback if the blobber is already in repair mode otherwise it will lead to inconsistent state
	if rb.lvm == nil || rb.lvm.VersionMarker.IsRepair {
		return nil
	}
	vm := &marker.VersionMarker{
		ClientID:     client.GetClientID(),
		BlobberID:    rb.lvm.VersionMarker.BlobberID,
		AllocationID: rb.lvm.VersionMarker.AllocationID,
		Version:      rb.lvm.VersionMarker.Version - 1,
		Timestamp:    rb.lvm.VersionMarker.Timestamp,
	}

	err := vm.Sign()
	if err != nil {
		l.Logger.Error("Signing writemarker failed: ", err)
		return err
	}
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	vmData, err := json.Marshal(vm)
	if err != nil {
		l.Logger.Error("Creating writemarker failed: ", err)
		return err
	}
	connID := zboxutil.NewConnectionId()
	formWriter.WriteField("version_marker", string(vmData))
	formWriter.WriteField("connection_id", connID)
	formWriter.Close()

	req, err := zboxutil.NewRollbackRequest(rb.blobber.Baseurl, vm.AllocationID, tx, body)
	if err != nil {
		l.Logger.Error("Creating rollback request failed: ", err)
		return err
	}
	req.Header.Add("Content-Type", formWriter.FormDataContentType())

	l.Logger.Info("Sending Rollback request to blobber: ", rb.blobber.Baseurl)

	var (
		shouldContinue bool
	)

	for retries := 0; retries < 3; retries++ {
		err, shouldContinue = func() (err error, shouldContinue bool) {
			reqCtx, ctxCncl := context.WithTimeout(ctx, DefaultUploadTimeOut)
			resp, err := zboxutil.Client.Do(req.WithContext(reqCtx))
			defer ctxCncl()
			if err != nil {
				l.Logger.Error("Rollback request failed: ", err)
				return
			}

			if resp.Body != nil {
				defer resp.Body.Close()
			}

			var respBody []byte
			respBody, err = io.ReadAll(resp.Body)
			if err != nil {
				l.Logger.Error("Response read: ", err)
				return
			}
			if resp.StatusCode == http.StatusOK {
				l.Logger.Info(rb.blobber.Baseurl, connID, "rollbacked")
				return
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				l.Logger.Info(rb.blobber.Baseurl, connID, "got too many request error. Retrying")
				var r int
				r, err = zboxutil.GetRateLimitValue(resp)
				if err != nil {
					l.Logger.Error(err)
					return
				}

				time.Sleep(time.Duration(r) * time.Second)
				shouldContinue = true
				return
			}

			if strings.Contains(string(respBody), "pending_markers:") {
				l.Logger.Info("Commit pending for blobber ",
					rb.blobber.Baseurl, " Retrying")
				time.Sleep(5 * time.Second)
				shouldContinue = true
				return
			}

			if strings.Contains(string(respBody), "chain_length_exceeded") {
				l.Logger.Info("Chain length exceeded for blobber ",
					rb.blobber.Baseurl, " Retrying")
				time.Sleep(5 * time.Second)
				shouldContinue = true
				return
			}

			err = thrown.New("commit_error",
				fmt.Sprintf("Got error response %s with status %d", respBody, resp.StatusCode))

			return
		}()
		if err != nil {
			l.Logger.Error(err)
			return err
		}
		if shouldContinue {
			continue
		}
		return nil

	}

	return thrown.New("rolback_error", fmt.Sprint("Rollback failed"))
}

func (a *Allocation) CheckAllocStatus() (AllocStatus, []BlobberStatus, error) {

	wg := &sync.WaitGroup{}
	markerChan := make(chan *RollbackBlobber, len(a.Blobbers))
	var errCnt int32
	var markerError error
	blobberRes := make([]BlobberStatus, len(a.Blobbers))
	for ind, blobber := range a.Blobbers {

		wg.Add(1)
		go func(blobber *blockchain.StorageNode, ind int) {

			defer wg.Done()
			blobStatus := BlobberStatus{
				ID:     blobber.ID,
				Status: "available",
			}
			lvm, err := GetWritemarker(a.ID, a.Tx, blobber.ID, blobber.Baseurl)
			if err != nil {
				atomic.AddInt32(&errCnt, 1)
				markerError = err
				l.Logger.Error("error during getWritemarker", zap.Error(err))
				blobStatus.Status = "unavailable"
			}
			if lvm == nil {
				markerChan <- nil
			} else {
				markerChan <- &RollbackBlobber{
					blobber:      blobber,
					lvm:          lvm,
					commitResult: &CommitResult{},
					blobIndex:    ind,
				}
				blobber.AllocationVersion = lvm.VersionMarker.Version
			}
			blobberRes[ind] = blobStatus
		}(blobber, ind)

	}
	wg.Wait()
	close(markerChan)
	if (a.ParityShards > 0 && errCnt > int32(a.ParityShards)) || (a.ParityShards == 0 && errCnt > 0) {
		return Broken, blobberRes, common.NewError("check_alloc_status_failed", markerError.Error())
	}

	versionMap := make(map[int64][]*RollbackBlobber)

	var (
		consensusReached bool
		latestVersion    int64
		prevVersion      int64
	)

	for rb := range markerChan {

		if rb == nil || rb.lvm == nil {
			continue
		}

		version := rb.lvm.VersionMarker.Version
		if version > latestVersion {
			latestVersion = version
			prevVersion = latestVersion
		}

		if _, ok := versionMap[version]; !ok {
			versionMap[version] = make([]*RollbackBlobber, 0)
		}

		versionMap[version] = append(versionMap[version], rb)
		if len(versionMap[version]) >= a.DataShards && version == latestVersion {
			consensusReached = true
		}
	}

	req := a.DataShards

	if len(versionMap) == 0 {
		return Commit, blobberRes, nil
	}

	if consensusReached {
		a.allocationVersion = latestVersion
		return Commit, blobberRes, nil
	}

	if len(versionMap[latestVersion]) >= req {
		for _, rb := range versionMap[prevVersion] {
			blobberRes[rb.blobIndex].Status = "repair"
		}
		return Repair, blobberRes, nil
	}

	// rollback to previous version
	l.Logger.Info("Rolling back to previous version")
	fullConsensus := len(versionMap[latestVersion]) - (req - len(versionMap[prevVersion]))
	errCnt = 0
	l.Logger.Info("fullConsensus", zap.Int32("fullConsensus", int32(fullConsensus)), zap.Int("latestLen", len(versionMap[latestVersion])), zap.Int("prevLen", len(versionMap[prevVersion])))
	for _, rb := range versionMap[latestVersion] {

		wg.Add(1)
		go func(rb *RollbackBlobber) {
			defer wg.Done()
			err := rb.processRollback(context.TODO(), a.Tx)
			if err != nil {
				atomic.AddInt32(&errCnt, 1)
				rb.commitResult = ErrorCommitResult(err.Error())
				l.Logger.Error("error during rollback", zap.Error(err))
			} else {
				rb.commitResult = SuccessCommitResult()
			}
		}(rb)
	}

	wg.Wait()
	if errCnt > int32(fullConsensus) {
		return Broken, blobberRes, common.NewError("rollback_failed", "Rollback failed")
	}

	if errCnt == int32(fullConsensus) {
		return Repair, blobberRes, nil
	}

	return Rollback, blobberRes, nil
}

func (a *Allocation) RollbackWithMask(mask zboxutil.Uint128) {

	wg := &sync.WaitGroup{}
	markerChan := make(chan *RollbackBlobber, mask.CountOnes())
	var pos uint64
	for i := mask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		blobber := a.Blobbers[pos]
		wg.Add(1)
		go func(blobber *blockchain.StorageNode) {

			defer wg.Done()
			wr, err := GetWritemarker(a.ID, a.Tx, blobber.ID, blobber.Baseurl)
			if err != nil {
				l.Logger.Error("error during getWritemarker", zap.Error(err))
			}
			if wr == nil {
				markerChan <- nil
			} else {
				markerChan <- &RollbackBlobber{
					blobber:      blobber,
					lvm:          wr,
					commitResult: &CommitResult{},
				}
			}
		}(blobber)

	}
	wg.Wait()
	close(markerChan)

	for rb := range markerChan {
		if rb == nil || rb.lvm == nil {
			continue
		}
		wg.Add(1)
		go func(rb *RollbackBlobber) {
			defer wg.Done()
			err := rb.processRollback(context.TODO(), a.Tx)
			if err != nil {
				rb.commitResult = ErrorCommitResult(err.Error())
				l.Logger.Error("error during rollback", zap.Error(err))
			} else {
				rb.commitResult = SuccessCommitResult()
			}
		}(rb)
	}

	wg.Wait()
}
