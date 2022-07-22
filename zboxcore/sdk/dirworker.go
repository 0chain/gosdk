package sdk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/bits"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

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
	dirMask      uint32
	mu           *sync.Mutex
	connectionID string
	Consensus
}

func (req *DirRequest) ProcessDir(a *Allocation) error {
	numList := len(a.Blobbers)
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)

	l.Logger.Info("Start creating dir for blobbers")

	for i := 0; i < numList; i++ {
		go func(blobberIdx int) {
			defer req.wg.Done()

			err := req.createDirInBlobber(a.Blobbers[blobberIdx], blobberIdx)
			if err != nil {
				l.Logger.Error(err.Error())
				return
			}
		}(i)
	}

	req.wg.Wait()

	if !req.isConsensusOk() {
		return errors.New("directory creation failed due to consensus not met")
	}

	writeMarkerMU, err := CreateWriteMarkerMutex(client.GetClient(), a)
	if err != nil {
		return fmt.Errorf("directory creation failed. Err: %s", err.Error())
	}
	err = writeMarkerMU.Lock(context.TODO(), req.connectionID)
	defer writeMarkerMU.Unlock(context.TODO(), req.connectionID) //nolint: errcheck
	if err != nil {
		return fmt.Errorf("directory creation failed. Err: %s", err.Error())
	}

	req.consensus = 0
	wg := &sync.WaitGroup{}
	okBlobbers := bits.OnesCount32(req.dirMask)
	wg.Add(okBlobbers)
	commitReqs := make([]*CommitRequest, okBlobbers)
	var c, pos int
	for i := req.dirMask; i != 0; i &= ^(1 << pos) {
		pos = bits.TrailingZeros32(i)
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
		go AddCommitRequest(commitReq)
		c++
	}
	wg.Wait()
	for _, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				l.Logger.Info("Commit success", commitReq.blobber.Baseurl)
				req.consensus++
			} else {
				l.Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			l.Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}

		if !req.isConsensusOk() {
			return errors.New("directory creation failed due consensus not met")
		}
	}
	return nil
}

func (req *DirRequest) commitRequest() error {
	req.consensus = 0
	wg := &sync.WaitGroup{}
	activeBlobbersNum := bits.OnesCount32(req.dirMask)
	wg.Add(activeBlobbersNum)

	commitReqs := make([]*CommitRequest, activeBlobbersNum)
	for i, blobber := range zboxutil.GetActiveBlobbers(req.dirMask, req.blobbers) {
		commitReq := &CommitRequest{}
		commitReq.allocationID = req.allocationID
		commitReq.allocationTx = req.allocationTx
		commitReq.blobber = blobber

		newChange := &allocationchange.DirCreateChange{}
		newChange.RemotePath = req.remotePath

		commitReq.changes = append(commitReq.changes, newChange)
		commitReq.connectionID = req.connectionID
		commitReq.wg = wg
		commitReqs[i] = commitReq
		go AddCommitRequest(commitReq)
	}
	wg.Wait()

	for _, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				l.Logger.Info("Commit success", commitReq.blobber.Baseurl)
				req.consensus++
			} else {
				l.Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			l.Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}

		if !req.isConsensusOk() {
			return errors.New("directory creation failed due consensus not met")
		}
	}
	return nil
}

func (req *DirRequest) createDirInBlobber(blobber *blockchain.StorageNode, blobberIdx int) error {
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	formWriter.WriteField("connection_id", req.connectionID)

	formWriter.WriteField("dir_path", req.remotePath)

	formWriter.Close()
	httpreq, err := zboxutil.NewCreateDirRequest(blobber.Baseurl, req.allocationID, body)
	if err != nil {
		l.Logger.Error(blobber.Baseurl, "Error creating dir request", err)
		return err
	}

	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))

	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("createdir : ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			resp_body, _ := ioutil.ReadAll(resp.Body)
			l.Logger.Info("createdir resp:", string(resp_body))
			req.mu.Lock()
			req.consensus++
			req.dirMask |= (1 << blobberIdx)
			req.mu.Unlock()
			l.Logger.Info(blobber.Baseurl, " "+req.remotePath, " created.")
		} else {
			resp_body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			msg := string(resp_body)
			l.Logger.Error(blobber.Baseurl, "Response: ", msg)
			if strings.Contains(msg, DirectoryExists) {
				req.mu.Lock()
				req.consensus++
				// should not add dirMask because there is not need to commit
				req.mu.Unlock()
			}
			return errors.New(msg)

		}
		return err
	})

	if err != nil {
		return err
	}

	return nil
}
