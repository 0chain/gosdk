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
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

const (
	DirectoryExists = "directory_exists"
)

type DirRequest struct {
	allocationID string
	allocationTx string
	remotePath   string
	blobbers     []*blockchain.StorageNode
	ctx          context.Context
	wg           *sync.WaitGroup
	dirMask      zboxutil.Uint128
	mu           *sync.Mutex
	connectionID string
	Consensus
}

func (req *DirRequest) ProcessDir(a *Allocation) error {
	l.Logger.Info("Start creating dir for blobbers")

	var pos uint64
	var existingDirCount int

	for i := req.dirMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		req.wg.Add(1)
		go func(pos uint64) {
			defer req.wg.Done()

			err, alreadyExists := req.createDirInBlobber(a.Blobbers[pos], pos)
			if err != nil {
				l.Logger.Error(err.Error())
			}
			if alreadyExists {
				existingDirCount++
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
		context.TODO(), &req.dirMask, req.mu,
		req.blobbers, &req.Consensus, time.Minute, req.connectionID)
	defer writeMarkerMU.Unlock(context.TODO(), req.dirMask,
		a.Blobbers, time.Minute, req.connectionID) //nolint: errcheck
	if err != nil {
		return fmt.Errorf("directory creation failed. Err: %s", err.Error())
	}

	return req.commitRequest(existingDirCount)
}

func (req *DirRequest) commitRequest(existingDirCount int) error {
	req.Consensus.Reset()
	req.consensus = existingDirCount
	wg := &sync.WaitGroup{}
	activeBlobbersNum := req.dirMask.CountOnes()
	wg.Add(activeBlobbersNum)

	commitReqs := make([]*CommitRequest, activeBlobbersNum)
	var pos uint64
	var c int
	for i := req.dirMask; !i.Equals(zboxutil.NewUint128(0)); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		commitReq := &CommitRequest{}
		commitReq.allocationID = req.allocationID
		commitReq.allocationTx = req.allocationTx
		commitReq.blobber = req.blobbers[pos]

		newChange := &allocationchange.DirCreateChange{}
		newChange.RemotePath = req.remotePath

		commitReq.changes = append(commitReq.changes, newChange)
		commitReq.connectionID = req.connectionID
		commitReq.wg = wg
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

	var resp *http.Response
	for i := 0; i < 3; i++ {
		ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))

		resp, err = zboxutil.Client.Do(httpreq.WithContext(ctx))
		cncl()
		if err != nil {
			l.Logger.Error(err)
			return err, false
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			l.Logger.Info("Successfully created directory ", req.remotePath)
			req.Consensus.Done()
			return nil, false
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			var r int
			r, err = zboxutil.GetRateLimitValue(resp)
			if err != nil {
				return err, false
			}
			l.Logger.Debug(fmt.Sprintf("Got too many request error. Retrying after %d seconds", r))
			time.Sleep(time.Duration(r) * time.Second)
			continue
		}

		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err, false
		}

		msg := string(resp_body)
		l.Logger.Error(blobber.Baseurl, " Response: ", msg)
		if strings.Contains(msg, DirectoryExists) {
			req.Consensus.Done()
			req.mu.Lock()
			req.dirMask = req.dirMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
			req.mu.Unlock()
			return nil, true
		}

		return errors.New("response_error", msg), false

	}

	return errors.New("dir_creation_failed",
		fmt.Sprintf("Directory creation failed with response status: %d", resp.StatusCode)), false
}
