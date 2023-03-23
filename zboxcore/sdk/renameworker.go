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

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"

	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type RenameRequest struct {
	allocationObj  *Allocation
	allocationID   string
	allocationTx   string
	blobbers       []*blockchain.StorageNode
	remotefilepath string
	newName        string
	ctx            context.Context
	ctxCncl        context.CancelFunc
	wg             *sync.WaitGroup
	renameMask     zboxutil.Uint128
	maskMU         *sync.Mutex
	connectionID   string
	consensus      Consensus
}

func (req *RenameRequest) getObjectTreeFromBlobber(blobber *blockchain.StorageNode) (fileref.RefEntity, error) {
	return getObjectTreeFromBlobber(req.ctx, req.allocationID, req.allocationTx, req.remotefilepath, blobber)
}

func (req *RenameRequest) renameBlobberObject(
	blobber *blockchain.StorageNode, blobberIdx int) (refEntity fileref.RefEntity, err error) {

	defer func() {
		if err != nil {
			req.maskMU.Lock()
			req.renameMask = req.renameMask.And(zboxutil.NewUint128(1).Lsh(uint64(blobberIdx)).Not())
			req.maskMU.Unlock()
		}
	}()

	refEntity, err = req.getObjectTreeFromBlobber(req.blobbers[blobberIdx])
	if err != nil {
		return nil, err
	}

	var (
		resp             *http.Response
		shouldContinue   bool
		latestRespMsg    string
		latestStatusCode int
	)

	for i := 0; i < 3; i++ {
		err, shouldContinue = func() (err error, shouldContinue bool) {
			body := new(bytes.Buffer)
			formWriter := multipart.NewWriter(body)

			formWriter.WriteField("connection_id", req.connectionID)
			formWriter.WriteField("path", req.remotefilepath)
			formWriter.WriteField("new_name", req.newName)
			formWriter.Close()

			var httpreq *http.Request
			httpreq, err = zboxutil.NewRenameRequest(blobber.Baseurl, req.allocationTx, body)
			if err != nil {
				l.Logger.Error(blobber.Baseurl, "Error creating rename request", err)
				return
			}

			httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
			ctx, cncl := context.WithTimeout(req.ctx, DefaultUploadTimeOut)
			resp, err = zboxutil.Client.Do(httpreq.WithContext(ctx))
			defer cncl()

			if err != nil {
				logger.Logger.Error("Rename: ", err)
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
				req.consensus.Done()
				l.Logger.Info(blobber.Baseurl, " "+req.remotefilepath, " renamed.")
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

func (req *RenameRequest) ProcessRename() error {
	defer req.ctxCncl()

	numList := len(req.blobbers)
	objectTreeRefs := make([]fileref.RefEntity, numList)

	wgErrors := make(chan error)
	wgDone := make(chan bool)

	req.wg = &sync.WaitGroup{}
	req.wg.Add(numList)
	for i := 0; i < numList; i++ {
		go func(blobberIdx int) {
			defer req.wg.Done()
			refEntity, err := req.renameBlobberObject(req.blobbers[blobberIdx], blobberIdx)
			if err == nil {
				req.consensus.Done()
				req.maskMU.Lock()
				objectTreeRefs[blobberIdx] = refEntity
				req.maskMU.Unlock()
				return
			}
			select {
			case wgErrors <- err:
			default:
			}
			l.Logger.Error(err.Error())
		}(i)
	}

	go func() {
		req.wg.Wait()
		close(wgDone)
	}()

	wgErrorsList := []error{}

	select {
	case <-wgDone:
		break
	case err := <-wgErrors:
		wgErrorsList = append(wgErrorsList, err)
	}

	if !req.consensus.isConsensusOk() && req.consensus.getConsensus() == 0 && len(wgErrorsList) > 1 {
		return errors.New("rename_failed", fmt.Sprintf("Rename failed. %s", wgErrorsList[0]))
	}

	if !req.consensus.isConsensusOk() {
		return errors.New("consensus_not_met",
			fmt.Sprintf("Rename failed. Required consensus %d got %d",
				req.consensus.consensusThresh, req.consensus.getConsensus()))
	}

	writeMarkerMutex, err := CreateWriteMarkerMutex(client.GetClient(), req.allocationObj)
	if err != nil {
		return fmt.Errorf("rename failed: %s", err.Error())
	}

	err = writeMarkerMutex.Lock(req.ctx, &req.renameMask,
		req.maskMU, req.blobbers, &req.consensus, 0, time.Minute, req.connectionID)
	if err != nil {
		return fmt.Errorf("rename failed: %s", err.Error())
	}
	defer writeMarkerMutex.Unlock(req.ctx, req.renameMask, req.blobbers, time.Minute, req.connectionID) //nolint: errcheck

	req.consensus.Reset()
	activeBlobbers := req.renameMask.CountOnes()
	wg := &sync.WaitGroup{}
	wg.Add(activeBlobbers)
	commitReqs := make([]*CommitRequest, activeBlobbers)
	var pos uint64
	var c int
	for i := req.renameMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		newChange := &allocationchange.RenameFileChange{
			NewName:    req.newName,
			ObjectTree: objectTreeRefs[pos],
		}
		newChange.Operation = constants.FileOperationRename
		newChange.Size = 0

		commitReq := &CommitRequest{
			allocationID: req.allocationID,
			allocationTx: req.allocationTx,
			blobber:      req.blobbers[pos],
			connectionID: req.connectionID,
			wg:           wg,
		}
		commitReq.change = newChange
		commitReqs[c] = commitReq

		go AddCommitRequest(commitReq)

		c++
	}

	wg.Wait()

	var errMessages string
	for _, commitReq := range commitReqs {
		if commitReq.result != nil {
			if commitReq.result.Success {
				l.Logger.Info("Commit success", commitReq.blobber.Baseurl)
				req.consensus.Done()
			} else {
				errMessages += commitReq.result.ErrorMessage + "\t"
				l.Logger.Info("Commit failed", commitReq.blobber.Baseurl, commitReq.result.ErrorMessage)
			}
		} else {
			l.Logger.Info("Commit result not set", commitReq.blobber.Baseurl)
		}
	}

	if !req.consensus.isConsensusOk() {
		return errors.New("consensus_not_met",
			fmt.Sprintf("Required consensus %d got %d. Error: %s",
				req.consensus.consensusThresh, req.consensus.consensus, errMessages))
	}
	return nil
}
