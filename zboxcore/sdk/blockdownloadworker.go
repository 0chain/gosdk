package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

type BlockDownloadRequest struct {
	blobber            *blockchain.StorageNode
	allocationID       string
	allocationTx       string
	blobberIdx         int
	remotefilepath     string
	remotefilepathhash string
	chunkSize          int
	blockNum           int64
	encryptedKey       string
	contentMode        string
	numBlocks          int64
	rxPay              bool
	authTicket         *marker.AuthTicket
	wg                 *sync.WaitGroup
	ctx                context.Context
	result             chan *downloadBlock
}

type downloadBlock struct {
	RawData     []byte `json:"data"`
	BlockChunks [][]byte
	Success     bool               `json:"success"`
	LatestRM    *marker.ReadMarker `json:"latest_rm"`
	idx         int
	err         error
	NumBlocks   int64 `json:"num_of_blocks"`
}

var blobberReadCounter *sync.Map

func getBlobberReadCtr(blobber *blockchain.StorageNode) int64 {
	rctr, ok := blobberReadCounter.Load(blobber.ID)
	if ok {
		return rctr.(int64)
	}
	return int64(0)
}

func incBlobberReadCtr(blobber *blockchain.StorageNode, numBlocks int64) {
	rctr, ok := blobberReadCounter.Load(blobber.ID)
	if !ok {
		rctr = int64(0)
	}
	blobberReadCounter.Store(blobber.ID, (rctr.(int64))+numBlocks)
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

func (req *BlockDownloadRequest) splitData(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}

func (req *BlockDownloadRequest) downloadBlobberBlock() {
	defer req.wg.Done()
	if req.numBlocks <= 0 {
		req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.New("invalid_request", "Invalid number of blocks for download")}
		return
	}
	retry := 0
	for retry < 3 {

		if req.blobber.IsSkip() {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx,
				err: errors.New("", "skip blobber by previous errors")}
			return
		}

		rm := &marker.ReadMarker{}
		rm.ClientID = client.GetClientID()
		rm.ClientPublicKey = client.GetClientPublicKey()
		rm.BlobberID = req.blobber.ID
		rm.AllocationID = req.allocationID
		rm.OwnerID = client.GetClientID()
		rm.Timestamp = common.Now()
		rm.ReadCounter = getBlobberReadCtr(req.blobber) + req.numBlocks
		err := rm.Sign()
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error: Signing readmarker failed")}
			return
		}
		body := new(bytes.Buffer)
		formWriter := multipart.NewWriter(body)
		rmData, err := json.Marshal(rm)
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error creating readmarker")}
			return
		}
		if len(req.remotefilepath) > 0 {
			req.remotefilepathhash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
		}
		formWriter.WriteField("path_hash", req.remotefilepathhash)

		if req.rxPay {
			formWriter.WriteField("rx_pay", "true") // pay oneself
		}

		formWriter.WriteField("block_num", fmt.Sprintf("%d", req.blockNum))
		formWriter.WriteField("num_blocks", fmt.Sprintf("%d", req.numBlocks))
		formWriter.WriteField("read_marker", string(rmData))
		if req.authTicket != nil {
			authTicketBytes, _ := json.Marshal(req.authTicket)
			formWriter.WriteField("auth_token", string(authTicketBytes))
		}
		if len(req.contentMode) > 0 {
			formWriter.WriteField("content", req.contentMode)
		}

		formWriter.Close()
		httpreq, err := zboxutil.NewDownloadRequest(req.blobber.Baseurl, req.allocationTx, body)
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error creating download request")}
			return
		}
		httpreq.Header.Add("Content-Type", formWriter.FormDataContentType())

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

				response, _ := ioutil.ReadAll(resp.Body)
				var rspData downloadBlock
				rspData.idx = req.blobberIdx
				err = json.Unmarshal(response, &rspData)

				// After getting start of stream JSON message, other message chunks should not be in JSON
				if err != nil {
					rspData.Success = true
					rspData.RawData = response

					// TODO download by 'chunks' should not download full stream and then chunk it
					// It has to download already chunked data
					if len(req.encryptedKey) > 0 {

						var chunks [][]byte
						if req.authTicket == nil {
							// 256 for the additional header bytes,  where chunk_size - 2 * 1024 is the encrypted data size
							chunks = req.splitData(rspData.RawData, req.chunkSize-2*1024+256)
						} else {
							chunks = req.splitData(rspData.RawData, req.chunkSize)
						}

						rspData.BlockChunks = chunks
					} else {
						chunks := req.splitData(rspData.RawData, req.chunkSize)
						rspData.BlockChunks = chunks
					}
					rspData.RawData = []byte{}
					incBlobberReadCtr(req.blobber, req.numBlocks)
					req.result <- &rspData
					return nil
				}

				if !rspData.Success && rspData.LatestRM != nil && rspData.LatestRM.ReadCounter >= getBlobberReadCtr(req.blobber) {
					Logger.Info("Will be retrying download")
					setBlobberReadCtr(req.blobber, rspData.LatestRM.ReadCounter)
					shouldRetry = true
					return errors.New("", "Need to retry the download")
				}

			} else {
				resp_body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				err = fmt.Errorf("Response Error: %s", string(resp_body))
				if strings.Contains(err.Error(), "not_enough_tokens") {
					shouldRetry, retry = false, 3 // don't repeat
					req.blobber.SetSkip(true)
				}
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
