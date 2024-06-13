package sdk

import (
	"bytes"
	"context"
	"encoding/hex"
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
	"github.com/minio/sha256-simd"
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
	changes      []allocationchange.AllocationChange
	blobber      *blockchain.StorageNode
	allocationID string
	allocationTx string
	connectionID string
	wg           *sync.WaitGroup
	result       *CommitResult
	timestamp    int64
	blobberInd   uint64
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
	paths := make([]string, 0)
	for _, change := range commitreq.changes {
		paths = append(paths, change.GetAffectedPath()...)
	}
	if len(paths) == 0 {
		l.Logger.Debug("Nothing to commit")
		commitreq.result = SuccessCommitResult()
		return
	}
	var req *http.Request
	var lR ReferencePathResult
	req, err := zboxutil.NewReferencePathRequest(commitreq.blobber.Baseurl, commitreq.allocationID, commitreq.allocationTx, paths)
	if err != nil {
		l.Logger.Error("Creating ref path req", err)
		return
	}
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("Ref path error:", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			l.Logger.Error("Ref path response : ", resp.StatusCode)
		}
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Logger.Error("Ref path: Resp", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New(
				strconv.Itoa(resp.StatusCode),
				fmt.Sprintf("Reference path error response: Status: %d - %s ",
					resp.StatusCode, string(resp_body)))
		}
		err = json.Unmarshal(resp_body, &lR)
		if err != nil {
			l.Logger.Error("Reference path json decode error: ", err)
			return err
		}
		return nil
	})

	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		return
	}
	rootRef, err := lR.GetDirTree(commitreq.allocationID)

	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		return
	}
	hasher := sha256.New()
	if lR.LatestWM != nil {
		err = lR.LatestWM.VerifySignature(client.GetClientPublicKey())
		if err != nil {
			e := errors.New("signature_verification_failed", err.Error())
			commitreq.result = ErrorCommitResult(e.Error())
			return
		}
		if commitreq.timestamp <= lR.LatestWM.Timestamp {
			commitreq.timestamp = lR.LatestWM.Timestamp + 1
		}

		rootRef.CalculateHash()
		prevAllocationRoot := rootRef.Hash
		if prevAllocationRoot != lR.LatestWM.AllocationRoot {
			l.Logger.Error("Allocation root from latest writemarker mismatch. Expected: " + prevAllocationRoot + " got: " + lR.LatestWM.AllocationRoot)
			errMsg := fmt.Sprintf(
				"calculated allocation root mismatch from blobber %s. Expected: %s, Got: %s",
				commitreq.blobber.Baseurl, prevAllocationRoot, lR.LatestWM.AllocationRoot)
			commitreq.result = ErrorCommitResult(errMsg)
			return
		}
		if lR.LatestWM.ChainHash != "" {
			prevChainHash, err := hex.DecodeString(lR.LatestWM.ChainHash)
			if err != nil {
				commitreq.result = ErrorCommitResult(err.Error())
				return
			}
			hasher.Write(prevChainHash) //nolint:errcheck
		}
	}

	var size int64
	fileIDMeta := make(map[string]string)

	for _, change := range commitreq.changes {
		err = change.ProcessChange(rootRef, fileIDMeta)
		if err != nil {
			commitreq.result = ErrorCommitResult(err.Error())
			return
		}
		size += change.GetSize()
	}
	rootRef.CalculateHash()
	var chainHash string
	if lR.Version == MARKER_VERSION {
		decodedHash, _ := hex.DecodeString(rootRef.Hash)
		hasher.Write(decodedHash) //nolint:errcheck
		chainHash = hex.EncodeToString(hasher.Sum(nil))
	}
	err = commitreq.commitBlobber(rootRef, chainHash, lR.LatestWM, size, fileIDMeta)
	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		return
	}
	l.Logger.Debug("[commitBlobber]", time.Since(start).Milliseconds())
	commitreq.result = SuccessCommitResult()
}

func (req *CommitRequest) commitBlobber(
	rootRef *fileref.Ref, chainHash string, latestWM *marker.WriteMarker, size int64,
	fileIDMeta map[string]string) (err error) {

	fileIDMetaData, err := json.Marshal(fileIDMeta)
	if err != nil {
		l.Logger.Error("Marshalling inode metadata failed: ", err)
		return err
	}

	wm := &marker.WriteMarker{}
	wm.AllocationRoot = rootRef.Hash
	wm.ChainSize = size
	if latestWM != nil {
		wm.PreviousAllocationRoot = latestWM.AllocationRoot
		wm.ChainSize += latestWM.ChainSize
	} else {
		wm.PreviousAllocationRoot = ""
	}
	if wm.AllocationRoot == wm.PreviousAllocationRoot {
		l.Logger.Debug("Allocation root and previous allocation root are same")
		return nil
	}
	wm.ChainHash = chainHash
	wm.FileMetaRoot = rootRef.FileMetaHash
	wm.AllocationID = req.allocationID
	wm.Size = size
	wm.BlobberID = req.blobber.ID
	wm.Timestamp = req.timestamp
	wm.ClientID = client.GetClientID()
	err = wm.Sign()
	if err != nil {
		l.Logger.Error("Signing writemarker failed: ", err)
		return err
	}
	wmData, err := json.Marshal(wm)
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
			formWriter, err := getFormWritter(req.connectionID, wmData, fileIDMetaData, body)
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

func getFormWritter(connectionID string, wmData, fileIDMetaData []byte, body *bytes.Buffer) (*multipart.Writer, error) {
	formWriter := multipart.NewWriter(body)
	err := formWriter.WriteField("connection_id", connectionID)
	if err != nil {
		return nil, err
	}

	err = formWriter.WriteField("write_marker", string(wmData))
	if err != nil {
		return nil, err
	}

	err = formWriter.WriteField("file_id_meta", string(fileIDMetaData))
	if err != nil {
		return nil, err
	}
	formWriter.Close()
	return formWriter, nil
}
