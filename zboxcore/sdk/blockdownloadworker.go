package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"0chain.net/clientsdk/core/common"
	"0chain.net/clientsdk/zboxcore/blockchain"
	"0chain.net/clientsdk/zboxcore/client"
	"0chain.net/clientsdk/zboxcore/fileref"
	. "0chain.net/clientsdk/zboxcore/logger"
	"0chain.net/clientsdk/zboxcore/marker"
	"0chain.net/clientsdk/zboxcore/zboxutil"
)

type BlockDownloadRequest struct {
	blobber            *blockchain.StorageNode
	allocationID       string
	blobberIdx         int
	remotefilepath     string
	remotefilepathhash string
	blockNum           int64
	authTicket         *marker.AuthTicket
	wg                 *sync.WaitGroup
	ctx                context.Context
	result             chan *downloadBlock
}

type downloadBlock struct {
	Data     []byte             `json:"data"`
	Success  bool               `json:"success"`
	LatestRM *marker.ReadMarker `json:"latest_rm"`
	idx      int
	err      error
}

var blobberReadCounter *sync.Map

func getBlobberReadCtr(blobber *blockchain.StorageNode) int64 {
	rctr, ok := blobberReadCounter.Load(blobber.ID)
	if ok {
		return rctr.(int64)
	}
	return int64(0)
}

func incBlobberReadCtr(blobber *blockchain.StorageNode) {
	rctr, ok := blobberReadCounter.Load(blobber.ID)
	if !ok {
		rctr = int64(0)
	}
	blobberReadCounter.Store(blobber.ID, (rctr.(int64))+1)
}

func setBlobberReadCtr(blobber *blockchain.StorageNode, ctr int64) {
	blobberReadCounter.Store(blobber.ID, ctr)
}

var downloadBlockChan map[string]chan *BlockDownloadRequest
var initDownloadMutex sync.Mutex

func InitBlockDownloader(blobbers []*blockchain.StorageNode) {
	initDownloadMutex.Lock()
	defer initDownloadMutex.Unlock()
	if downloadBlockChan == nil {
		// for _, v := range downloadBlockChan {
		// 	close(v)
		// }
		downloadBlockChan = make(map[string]chan *BlockDownloadRequest)
	}
	blobberReadCounter = &sync.Map{}

	for _, blobber := range blobbers {
		if _, ok := downloadBlockChan[blobber.ID]; !ok {
			downloadBlockChan[blobber.ID] = make(chan *BlockDownloadRequest, 1)
			blobberChan := downloadBlockChan[blobber.ID]
			go startBlockDownloadWorker(blobberChan)
		}
	}
}

func startBlockDownloadWorker(blobberChan chan *BlockDownloadRequest) {
	for true {
		blockDownloadReq, open := <-blobberChan
		if !open {
			break
		}
		blockDownloadReq.downloadBlobberBlock()
	}
}

func (req *BlockDownloadRequest) downloadBlobberBlock() {
	defer req.wg.Done()
	retry := 0
	for retry < 3 {
		rm := &marker.ReadMarker{}
		rm.ClientID = client.GetClientID()
		rm.ClientPublicKey = client.GetClientPublicKey()
		rm.BlobberID = req.blobber.ID
		rm.AllocationID = req.allocationID
		rm.OwnerID = client.GetClientID()
		rm.Timestamp = common.Now()
		rm.ReadCounter = getBlobberReadCtr(req.blobber) + 1
		err := rm.Sign()
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: fmt.Errorf("Error: Signing readmarker failed: %s", err.Error())}
			return
		}
		body := new(bytes.Buffer)
		formWriter := multipart.NewWriter(body)
		rmData, err := json.Marshal(rm)
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: fmt.Errorf("Error creating readmarker: %s", err.Error())}
			return
		}
		if len(req.remotefilepath) > 0 {
			req.remotefilepathhash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
		}
		formWriter.WriteField("path_hash", req.remotefilepathhash)

		formWriter.WriteField("block_num", fmt.Sprintf("%d", req.blockNum))
		formWriter.WriteField("read_marker", string(rmData))
		if req.authTicket != nil {
			authTicketBytes, _ := json.Marshal(req.authTicket)
			formWriter.WriteField("auth_token", string(authTicketBytes))
		}
		formWriter.Close()
		httpreq, err := zboxutil.NewDownloadRequest(req.blobber.Baseurl, req.allocationID, body)
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: fmt.Errorf("Error creating download request: %s", err.Error())}
			return
		}
		httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())
		// TODO: Fix the timeout
		ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
		shouldRetry := false
		err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
			if err != nil {
				return err
			}
			if resp.Body != nil {
				defer resp.Body.Close()
			}
			if resp.StatusCode == http.StatusOK {
				//req.consensus++
				response, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return fmt.Errorf("[%d] Read error:%s\n", req.blobberIdx, err.Error())
				}
				var rspData downloadBlock
				rspData.idx = req.blobberIdx
				err = json.Unmarshal(response, &rspData)
				if err != nil {
					return fmt.Errorf("[%d] Json decode error:%s\n", req.blobberIdx, err.Error())
				}
				if rspData.Success {
					incBlobberReadCtr(req.blobber)
					req.result <- &rspData
					return nil
				}
				if rspData.LatestRM != nil && rspData.LatestRM.ReadCounter >= getBlobberReadCtr(req.blobber) {
					Logger.Info("Will be retrying download")
					setBlobberReadCtr(req.blobber, rspData.LatestRM.ReadCounter)
					shouldRetry = true
					return fmt.Errorf("Need to retry the download")
				}

			} else {
				resp_body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				err = fmt.Errorf("Response Error: %s", string(resp_body))
				return err
			}
			return nil
		})
		if err != nil && (!shouldRetry || retry >= 3) {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: err}
		}
		if shouldRetry {
			retry++
		} else {
			break
		}
	}
}

func AddBlockDownloadReq(req *BlockDownloadRequest) {
	downloadBlockChan[req.blobber.ID] <- req
}
