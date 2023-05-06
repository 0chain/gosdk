package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/logger"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

const (
	DirectoryExists = "directory_exists"
)

type DirRequest struct {
	allocationObj *Allocation
	allocationID  string
	allocationTx  string
	remotePath    string
	blobbers      []*blockchain.StorageNode
	ctx           context.Context
	ctxCncl       context.CancelFunc
	wg            *sync.WaitGroup
	dirMask       zboxutil.Uint128
	mu            *sync.Mutex
	connectionID  string
	timestamp     int64
	Consensus
}

func (req *DirRequest) ProcessDir(a *Allocation) error {
	l.Logger.Info("Start creating dir for blobbers")

	defer req.ctxCncl()
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
				existingDirCount++
				countMu.Unlock()
			}
		}(pos)
	}

	req.wg.Wait()

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
	//Check if the allocation is to be repaired or rolled back
	status, err := req.allocationObj.CheckAllocStatus()
	if err != nil {
		logger.Logger.Error("Error checking allocation status: ", err)
		return fmt.Errorf("directory creation failed: %s", err.Error())
	}

	if status == Repair {
		logger.Logger.Info("Repairing allocation")
		//TODO: Need status callback to call repair allocation
		// err = req.allocationObj.RepairAlloc()
		// if err != nil {
		// 	return err
		// }
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

		newChange := &allocationchange.DirCreateChange{
			RemotePath: req.remotePath,
			Uuid:       uid,
			Timestamp:  common.Timestamp(req.timestamp),
		}

		commitReq.change = newChange
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
	formWriter.WriteField("connection_id", req.connectionID)

	formWriter.WriteField("dir_path", req.remotePath)

	formWriter.Close()
	httpreq, err := zboxutil.NewCreateDirRequest(blobber.Baseurl, req.allocationID, body)
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
			ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
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
			l.Logger.Error(blobber.Baseurl, " Response: ", msg)
			if strings.Contains(msg, DirectoryExists) {
				req.Consensus.Done()
				req.mu.Lock()
				req.dirMask = req.dirMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
				req.mu.Unlock()
				alreadyExists = true
				return
			}

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
