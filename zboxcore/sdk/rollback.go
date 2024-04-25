package sdk

import (
	"bytes"
	"context"
	"encoding/hex"
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
	"github.com/minio/sha256-simd"
	"go.uber.org/zap"
)

type LatestPrevWriteMarker struct {
	LatestWM *marker.WriteMarker `json:"latest_write_marker"`
	PrevWM   *marker.WriteMarker `json:"prev_write_marker"`
	Version  string              `json:"version"`
}

type AllocStatus byte

const (
	Commit AllocStatus = iota
	Repair
	Broken
	Rollback
)

var ErrRetryOperation = errors.New("retry_operation")

type RollbackBlobber struct {
	blobber      *blockchain.StorageNode
	commitResult *CommitResult
	lpm          *LatestPrevWriteMarker
}

func GetWritemarker(allocID, allocTx, sig, id, baseUrl string) (*LatestPrevWriteMarker, error) {

	var lpm LatestPrevWriteMarker

	req, err := zboxutil.NewWritemarkerRequest(baseUrl, allocID, allocTx, sig)
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
		err = json.Unmarshal(body, &lpm)
		if err != nil {
			return nil, err
		}
		if lpm.LatestWM != nil {
			err = lpm.LatestWM.VerifySignature(client.GetClientPublicKey())
			if err != nil {
				return nil, fmt.Errorf("signature verification failed for latest writemarker: %s", err.Error())
			}
			if lpm.PrevWM != nil {
				err = lpm.PrevWM.VerifySignature(client.GetClientPublicKey())
				if err != nil {
					return nil, fmt.Errorf("signature verification failed for latest writemarker: %s", err.Error())
				}
			}
		}
		return &lpm, nil
	}

	return nil, fmt.Errorf("writemarker error response %d", http.StatusTooManyRequests)
}

func (rb *RollbackBlobber) processRollback(ctx context.Context, tx string) error {

	wm := &marker.WriteMarker{}
	wm.AllocationID = rb.lpm.LatestWM.AllocationID
	wm.Timestamp = rb.lpm.LatestWM.Timestamp
	wm.BlobberID = rb.lpm.LatestWM.BlobberID
	wm.ClientID = client.GetClientID()
	wm.Size = -rb.lpm.LatestWM.Size
	wm.ChainSize = wm.Size + rb.lpm.LatestWM.ChainSize

	if rb.lpm.PrevWM != nil {
		wm.AllocationRoot = rb.lpm.PrevWM.AllocationRoot
		wm.PreviousAllocationRoot = rb.lpm.PrevWM.AllocationRoot
		wm.FileMetaRoot = rb.lpm.PrevWM.FileMetaRoot
		if wm.AllocationRoot == rb.lpm.LatestWM.AllocationRoot {
			return nil
		}
	}
	if rb.lpm.Version == MARKER_VERSION {
		decodedHash, _ := hex.DecodeString(wm.AllocationRoot)
		prevChainHash, _ := hex.DecodeString(rb.lpm.LatestWM.ChainHash)
		hasher := sha256.New()
		hasher.Write(prevChainHash) //nolint:errcheck
		hasher.Write(decodedHash)   //nolint:errcheck
		wm.ChainHash = hex.EncodeToString(hasher.Sum(nil))
	} else if rb.lpm.Version == "" {
		wm.Size = 0
	}

	err := wm.Sign()
	if err != nil {
		l.Logger.Error("Signing writemarker failed: ", err)
		return err
	}
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	wmData, err := json.Marshal(wm)
	if err != nil {
		l.Logger.Error("Creating writemarker failed: ", err)
		return err
	}
	connID := zboxutil.NewConnectionId()
	formWriter.WriteField("write_marker", string(wmData))
	formWriter.WriteField("connection_id", connID)
	formWriter.Close()

	req, err := zboxutil.NewRollbackRequest(rb.blobber.Baseurl, wm.AllocationID, tx, body)
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

func (a *Allocation) CheckAllocStatus() (AllocStatus, error) {

	wg := &sync.WaitGroup{}
	markerChan := make(chan *RollbackBlobber, len(a.Blobbers))
	var errCnt int32
	var markerError error
	for _, blobber := range a.Blobbers {

		wg.Add(1)
		go func(blobber *blockchain.StorageNode) {

			defer wg.Done()
			wr, err := GetWritemarker(a.ID, a.Tx, a.sig, blobber.ID, blobber.Baseurl)
			if err != nil {
				atomic.AddInt32(&errCnt, 1)
				markerError = err
				l.Logger.Error("error during getWritemarker", zap.Error(err))
			}
			if wr == nil {
				markerChan <- nil
			} else {
				markerChan <- &RollbackBlobber{
					blobber:      blobber,
					lpm:          wr,
					commitResult: &CommitResult{},
				}
			}
		}(blobber)

	}
	wg.Wait()
	close(markerChan)
	if a.ParityShards > 0 && errCnt > int32(a.ParityShards) {
		return Broken, common.NewError("check_alloc_status_failed", markerError.Error())
	}

	versionMap := make(map[string][]*RollbackBlobber)

	var (
		prevVersion   string
		latestVersion string
		highestTS     int64
	)

	for rb := range markerChan {

		if rb == nil || rb.lpm.LatestWM == nil {
			continue
		}

		version := rb.lpm.LatestWM.FileMetaRoot

		if highestTS < rb.lpm.LatestWM.Timestamp {
			prevVersion = latestVersion
			highestTS = rb.lpm.LatestWM.Timestamp
			latestVersion = version
		}

		if prevVersion == "" && version != latestVersion {
			prevVersion = version
		}

		if _, ok := versionMap[version]; !ok {
			versionMap[version] = make([]*RollbackBlobber, 0)
		}

		versionMap[version] = append(versionMap[version], rb)
	}

	if len(versionMap) < 2 {
		return Commit, nil
	}

	req := a.DataShards

	if len(versionMap[latestVersion]) > req {
		return Commit, nil
	}

	if len(versionMap[latestVersion]) >= req || len(versionMap[prevVersion]) >= req || len(versionMap) > 2 {
		// TODO: Return Repair after refactoring the repair function
		return Repair, nil
	} else {
		l.Logger.Info("versionMapLen", zap.Int("versionMapLen", len(versionMap)), zap.Int("latestLen", len(versionMap[latestVersion])), zap.Int("prevLen", len(versionMap[prevVersion])))
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
		return Broken, common.NewError("rollback_failed", "Rollback failed")
	}

	if errCnt == int32(fullConsensus) {
		return Repair, nil
	}

	return Rollback, nil
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
			wr, err := GetWritemarker(a.ID, a.Tx, a.sig, blobber.ID, blobber.Baseurl)
			if err != nil {
				l.Logger.Error("error during getWritemarker", zap.Error(err))
			}
			if wr == nil {
				markerChan <- nil
			} else {
				markerChan <- &RollbackBlobber{
					blobber:      blobber,
					lpm:          wr,
					commitResult: &CommitResult{},
				}
			}
		}(blobber)

	}
	wg.Wait()
	close(markerChan)

	for rb := range markerChan {
		if rb == nil || rb.lpm.LatestWM == nil {
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
