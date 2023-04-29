package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type ReferencePathResult struct {
	*fileref.ReferencePath
	LatestWM *marker.WriteMarker `json:"latest_write_marker"`
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

type CommitRequest struct {
	change       allocationchange.AllocationChange
	blobber      *blockchain.StorageNode
	allocationID string
	allocationTx string
	connectionID string
	wg           *sync.WaitGroup
	result       *CommitResult
	timestamp    int64
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

	l.Logger.Info("received a commit request")
	paths := make([]string, 0)
	paths = append(paths, commitreq.change.GetAffectedPath()...)
	var req *http.Request
	var lR ReferencePathResult
	req, err := zboxutil.NewReferencePathRequest(commitreq.blobber.Baseurl, commitreq.allocationTx, paths)
	if err != nil || len(paths) == 0 {
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

	if lR.LatestWM != nil {
		err = lR.LatestWM.VerifySignature(client.GetClientPublicKey())
		if err != nil {
			e := errors.New("signature_verification_failed", err.Error())
			commitreq.result = ErrorCommitResult(e.Error())
			return
		}

		rootRef.CalculateHash()
		prevAllocationRoot := rootRef.Hash
		if prevAllocationRoot != lR.LatestWM.AllocationRoot {
			l.Logger.Info("Allocation root from latest writemarker mismatch. Expected: " + prevAllocationRoot + " got: " + lR.LatestWM.AllocationRoot)
			errMsg := fmt.Sprintf(
				"calculated allocation root mismatch from blobber %s. Expected: %s, Got: %s",
				commitreq.blobber.Baseurl, prevAllocationRoot, lR.LatestWM.AllocationRoot)
			commitreq.result = ErrorCommitResult(errMsg)
			return
		}
	}

	var size int64
	commitParams, err := commitreq.change.ProcessChange(rootRef)

	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		return
	}
	size += commitreq.change.GetSize()
	err = commitreq.commitBlobber(rootRef, lR.LatestWM, size, &commitParams)
	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		return
	}
	commitreq.result = SuccessCommitResult()
}

func (req *CommitRequest) commitBlobber(
	rootRef *fileref.Ref, latestWM *marker.WriteMarker, size int64,
	commitParams *allocationchange.CommitParams) error {

	fileIDMetaData, err := json.Marshal(commitParams.FileIDMeta)
	if err != nil {
		l.Logger.Error("Marshalling inode metadata failed: ", err)
		return err
	}

	wm := &marker.WriteMarker{}
	wm.AllocationRoot = rootRef.Hash
	if latestWM != nil {
		wm.PreviousAllocationRoot = latestWM.AllocationRoot
	} else {
		wm.PreviousAllocationRoot = ""
	}

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
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	wmData, err := json.Marshal(wm)
	if err != nil {
		l.Logger.Error("Creating writemarker failed: ", err)
		return err
	}
	formWriter.WriteField("connection_id", req.connectionID)
	formWriter.WriteField("write_marker", string(wmData))
	formWriter.WriteField("file_id_meta", string(fileIDMetaData))

	formWriter.Close()

	httpreq, err := zboxutil.NewCommitRequest(req.blobber.Baseurl, req.allocationTx, body)
	if err != nil {
		l.Logger.Error("Error creating commit req: ", err)
		return err
	}
	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 60))
	l.Logger.Info("Committing to blobber." + req.blobber.Baseurl)
	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			l.Logger.Error("Commit: ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			l.Logger.Info(req.blobber.Baseurl, req.connectionID, " committed")
		} else {
			l.Logger.Error("Commit response: ", resp.StatusCode)
		}

		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Logger.Error("Response read: ", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			l.Logger.Error(req.blobber.Baseurl, " Commit response:", string(resp_body))
			return errors.New("commit_error", string(resp_body)+string(wmData))
		}
		return nil
	})
	return err
}

func AddCommitRequest(req *CommitRequest) {
	commitChan[req.blobber.ID] <- req
}

func (commitreq *CommitRequest) calculateHashRequest(ctx context.Context, paths []string) error {
	var req *http.Request
	req, err := zboxutil.NewCalculateHashRequest(commitreq.blobber.Baseurl, commitreq.allocationTx, paths)
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
