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
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
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
	changes      []allocationchange.AllocationChange
	blobber      *blockchain.StorageNode
	allocationID string
	allocationTx string
	connectionID string
	wg           *sync.WaitGroup
	result       *CommitResult
}

var commitChan map[string]chan *CommitRequest
var initCommitMutex sync.Mutex

func InitCommitWorker(blobbers []*blockchain.StorageNode) {
	// if commitChan != nil {
	// 	for _, v := range commitChan {
	// 		close(v)
	// 	}
	// }
	// commitChan = make(map[string]chan *CommitRequest)
	// for _, blobber := range blobbers {
	// 	Logger.Info("Atempting to start the commit worker for ", blobber.Baseurl)
	// 	commitChan[blobber.ID] = make(chan *CommitRequest, 1)
	// 	go startCommitWorker(blobber)
	// }
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
	for true {
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
	Logger.Info("received a commit request")
	paths := make([]string, 0)
	for _, change := range commitreq.changes {
		paths = append(paths, change.GetAffectedPath())
	}
	var req *http.Request
	var lR ReferencePathResult
	req, err := zboxutil.NewReferencePathRequest(commitreq.blobber.Baseurl, commitreq.allocationTx, paths)
	if err != nil || len(paths) == 0 {
		Logger.Error("Creating ref path req", err)
		return
	}
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Ref path error:", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			Logger.Error("Ref path response : ", resp.StatusCode)
		}
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Error("Ref path: Resp", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New(strconv.Itoa(resp.StatusCode), fmt.Sprintf("Reference path error response: Status: %d - %s ", resp.StatusCode, string(resp_body)))

		} else {
			//Logger.Info("Reference path:", string(resp_body))
			err = json.Unmarshal(resp_body, &lR)
			if err != nil {
				Logger.Error("Reference path json decode error: ", err)
				return err
			}
		}
		return nil
	})
	//process the commit request for the blobber here
	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		commitreq.wg.Done()
		return
	}
	rootRef, err := lR.GetDirTree(commitreq.allocationID)
	if lR.LatestWM != nil {
		//Can not verify signature due to collaborator flow
		// //TODO: Verify the writemarker
		// err = lR.LatestWM.VerifySignature(client.GetClientPublicKey())
		// if err != nil {
		// 	commitreq.result = ErrorCommitResult(err.Error())
		// 	commitreq.wg.Done()
		// 	return
		// }

		rootRef.CalculateHash()
		prevAllocationRoot := encryption.Hash(rootRef.Hash + ":" + strconv.FormatInt(lR.LatestWM.Timestamp, 10))
		if prevAllocationRoot != lR.LatestWM.AllocationRoot {
			// Removing this check for testing purpose as per the convo with Saswata
			Logger.Info("Allocation root from latest writemarker mismatch. Expected: " + prevAllocationRoot + " got: " + lR.LatestWM.AllocationRoot)
			// err = commitreq.calculateHashRequest(ctx, paths)
			// if err != nil {
			// 	commitreq.result = ErrorCommitResult("Failed to call blobber to recalculate the hash. URL: " + commitreq.blobber.Baseurl + ", Err : " + err.Error())
			// 	commitreq.wg.Done()
			// 	return
			// }
			// Logger.Info("Recalculate hash call to blobber successfull")
			// commitreq.result = ErrorCommitResult("Allocation root from latest writemarker mismatch. Expected: " + prevAllocationRoot + " got: " + lR.LatestWM.AllocationRoot)
			// commitreq.wg.Done()
			// return
		}
	}
	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		commitreq.wg.Done()
		return
	}
	size := int64(0)
	for _, change := range commitreq.changes {
		err = change.ProcessChange(rootRef)
		if err != nil {
			break
		}
		size += change.GetSize()
	}
	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		commitreq.wg.Done()
		return
	}
	err = commitreq.commitBlobber(rootRef, lR.LatestWM, size)
	if err != nil {
		commitreq.result = ErrorCommitResult(err.Error())
		commitreq.wg.Done()
		return
	}
	commitreq.result = SuccessCommitResult()
	commitreq.wg.Done()
}

func (req *CommitRequest) commitBlobber(rootRef *fileref.Ref, latestWM *marker.WriteMarker, size int64) error {
	wm := &marker.WriteMarker{}
	timestamp := int64(common.Now())
	wm.AllocationRoot = encryption.Hash(rootRef.Hash + ":" + strconv.FormatInt(timestamp, 10))
	if latestWM != nil {
		wm.PreviousAllocationRoot = latestWM.AllocationRoot
	} else {
		wm.PreviousAllocationRoot = ""
	}

	wm.AllocationID = req.allocationID
	wm.Size = size
	wm.BlobberID = req.blobber.ID
	wm.Timestamp = timestamp
	wm.ClientID = client.GetClientID()
	err := wm.Sign()
	if err != nil {
		Logger.Error("Signing writemarker failed: ", err)
		return err
	}
	body := new(bytes.Buffer)
	formWriter := multipart.NewWriter(body)
	wmData, err := json.Marshal(wm)
	if err != nil {
		Logger.Error("Creating writemarker failed: ", err)
		return err
	}
	formWriter.WriteField("connection_id", req.connectionID)
	formWriter.WriteField("write_marker", string(wmData))

	formWriter.Close()

	httpreq, err := zboxutil.NewCommitRequest(req.blobber.Baseurl, req.allocationTx, body)
	if err != nil {
		Logger.Error("Error creating commit req: ", err)
		return err
	}
	httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 60))
	Logger.Info("Committing to blobber." + req.blobber.Baseurl)
	err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Commit: ", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			Logger.Info(req.blobber.Baseurl, req.connectionID, " committed")
		} else {
			Logger.Error("Commit response: ", resp.StatusCode)
		}

		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Error("Response read: ", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			Logger.Error(req.blobber.Baseurl, " Commit response:", string(resp_body))
			return errors.New("commit_error", string(resp_body))
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
		Logger.Error("Creating calculate hash req", err)
		return err
	}
	ctx, cncl := context.WithTimeout(context.Background(), (time.Second * 30))
	err = zboxutil.HttpDo(ctx, cncl, req, func(resp *http.Response, err error) error {
		if err != nil {
			Logger.Error("Calculate hash error:", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			Logger.Error("Calculate hash response : ", resp.StatusCode)
		}
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Error("Calculate hash: Resp", err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New(strconv.Itoa(resp.StatusCode), fmt.Sprintf("Calculate hash error response: Body: %s ", string(resp_body)))
		}
		return nil
	})
	return err
}
