package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/0chain/errors"
	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type ReferencePathResult struct {
	*fileref.ReferencePath
	LatestWM *marker.WriteMarker `json:"latest_write_marker"`
	Version  string              `json:"version"`
}

type CommitResult struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_msg,omitempty"`
}

func ErrorCommitResult(errMsg string) *CommitResult {
	result := &CommitResult{Success: false, ErrorMessage: errMsg}
	return result
}

func SuccessCommitResult() *CommitResult {
	result := &CommitResult{Success: true}
	return result
}

const MARKER_VERSION = "v2"

type CommitRequest struct {
	changes       []allocationchange.AllocationChange
	blobber       *blockchain.StorageNode
	allocationID  string
	allocationTx  string
	connectionID  string
	sig           string
	wg            *sync.WaitGroup
	result        *CommitResult
	timestamp     int64
	blobberInd    uint64
	version       int64
	isRepair      bool
	repairVersion int64
	repairOffset  string
}

var commitChan map[string]chan *CommitRequest
var initCommitMutex sync.Mutex

func InitCommitWorker(blobbers []*blockchain.StorageNode) {
	initCommitMutex.Lock()
	defer initCommitMutex.Unlock()
	if commitChan == nil {
		commitChan = make(map[string]chan *CommitRequest)
	}

	for _, blobber := range blobbers {
		if _, ok := commitChan[blobber.ID]; !ok {
			commitChan[blobber.ID] = make(chan *CommitRequest, 1)
			blobberChan := commitChan[blobber.ID]
			go startCommitWorker(blobberChan, blobber.ID)
		}
	}

}

func startCommitWorker(blobberChan chan *CommitRequest, blobberID string) {
	for {
		commitreq, open := <-blobberChan
		if !open {
			break
		}
		commitreq.processCommit()
	}
	initCommitMutex.Lock()
	defer initCommitMutex.Unlock()
	delete(commitChan, blobberID)
}

func (commitreq *CommitRequest) processCommit() {
	defer commitreq.wg.Done()
	start := time.Now()
	l.Logger.Debug("received a commit request")
	err := commitreq.commitBlobber()
	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		return
	}
	l.Logger.Debug("[commitBlobber]", time.Since(start).Milliseconds())
	commitreq.result = SuccessCommitResult()
}

func (req *CommitRequest) commitBlobber() (err error) {
	vm := &marker.VersionMarker{
		Version:       req.version,
		Timestamp:     req.timestamp,
		ClientID:      client.GetClientID(),
		AllocationID:  req.allocationID,
		BlobberID:     req.blobber.ID,
		IsRepair:      req.isRepair,
		RepairVersion: req.repairVersion,
		RepairOffset:  req.repairOffset,
	}
	err = vm.Sign()
	if err != nil {
		l.Logger.Error("Signing writemarker failed: ", err)
		return err
	}
	vmData, err := json.Marshal(vm)
	if err != nil {
		l.Logger.Error("Creating writemarker failed: ", err)
		return err
	}

	l.Logger.Debug("Committing to blobber." + req.blobber.Baseurl)
	var (
		resp           *http.Response
		shouldContinue bool
	)
	for retries := 0; retries < 6; retries++ {
		err, shouldContinue = func() (err error, shouldContinue bool) {
			body := new(bytes.Buffer)
			formWriter, err := getFormWritter(req.connectionID, vmData, body)
			if err != nil {
				l.Logger.Error("Creating form writer failed: ", err)
				return
			}
			httpreq, err := zboxutil.NewCommitRequest(req.blobber.Baseurl, req.allocationID, req.allocationTx, body)
			if err != nil {
				l.Logger.Error("Error creating commit req: ", err)
				return
			}
			httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
			reqCtx, ctxCncl := context.WithTimeout(context.Background(), time.Second*60)
			resp, err = zboxutil.Client.Do(httpreq.WithContext(reqCtx))
			defer ctxCncl()

			if err != nil {
				logger.Logger.Error("Commit: ", err)
				return
			}

			if resp.Body != nil {
				defer resp.Body.Close()
			}

			var respBody []byte
			respBody, err = io.ReadAll(resp.Body)
			if err != nil {
				logger.Logger.Error("Response read: ", err)
				return
			}
			if resp.StatusCode == http.StatusOK {
				logger.Logger.Debug(req.blobber.Baseurl, " committed")
				return
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				logger.Logger.Debug(req.blobber.Baseurl,
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

			if strings.Contains(string(respBody), "pending_markers:") {
				logger.Logger.Debug("Commit pending for blobber ",
					req.blobber.Baseurl, " Retrying")
				time.Sleep(5 * time.Second)
				shouldContinue = true
				return
			}

			if strings.Contains(string(respBody), "chain_length_exceeded") {
				l.Logger.Error("Chain length exceeded for blobber ",
					req.blobber.Baseurl, " Retrying")
				time.Sleep(5 * time.Second)
				shouldContinue = true
				return
			}

			err = thrown.New("commit_error",
				fmt.Sprintf("Got error response %s with status %d", respBody, resp.StatusCode))
			return
		}()
		if shouldContinue {
			continue
		}
		return
	}
	return thrown.New("commit_error", fmt.Sprintf("Commit failed with response status %d", resp.StatusCode))
}

func AddCommitRequest(req *CommitRequest) {
	commitChan[req.blobber.ID] <- req
}

func (commitreq *CommitRequest) calculateHashRequest(ctx context.Context, paths []string) error { //nolint
	var req *http.Request
	req, err := zboxutil.NewCalculateHashRequest(commitreq.blobber.Baseurl, commitreq.allocationID, commitreq.allocationTx, paths)
	if err != nil || len(paths) == 0 {
		l.Logger.Error("Creating calculate hash req", err)
		return err
	}
	ctx, cncl := context.WithTimeout(ctx, (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("Calculate hash error:", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			l.Logger.Error("Calculate hash response : ", resp.StatusCode)
		}
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Logger.Error("Calculate hash: Resp", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New(strconv.Itoa(resp.StatusCode), fmt.Sprintf("Calculate hash error response: Body: %s ", string(resp_body)))
		}
		return nil
	})
	return err
}

func getFormWritter(connectionID string, vmData []byte, body *bytes.Buffer) (*multipart.Writer, error) {
	formWriter := multipart.NewWriter(body)
	err := formWriter.WriteField("connection_id", connectionID)
	if err != nil {
		return nil, err
	}

	err = formWriter.WriteField("version_marker", string(vmData))
	if err != nil {
		return nil, err
	}
	formWriter.Close()
	return formWriter, nil
}
