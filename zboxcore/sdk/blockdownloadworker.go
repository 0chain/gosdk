package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

var downloadBlockChan map[string]chan *BlockDownloadRequest
var initDownloadMutex sync.Mutex

func InitBlockDownloader(blobbers []*blockchain.StorageNode) {
	initDownloadMutex.Lock()
	defer initDownloadMutex.Unlock()
	if downloadBlockChan == nil {
		downloadBlockChan = make(map[string]chan *BlockDownloadRequest)
	}

	for _, blobber := range blobbers {
		if _, ok := downloadBlockChan[blobber.ID]; !ok {
			downloadBlockChan[blobber.ID] = make(chan *BlockDownloadRequest, 1)
			blobberChan := downloadBlockChan[blobber.ID]
			go startBlockDownloadWorker(blobberChan)
		}
	}
}

func startBlockDownloadWorker(blobberChan chan *BlockDownloadRequest) {
	for {
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
		chunks = append(chunks, buf[:])
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
		rm.ReadCounter = getBlobberReadCtr(req.blobber.ID) + req.numBlocks
		err := rm.Sign()
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error: Signing readmarker failed")}
			return
		}
		rmData, err := json.Marshal(rm)
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error creating readmarker")}
			return
		}
		if len(req.remotefilepath) > 0 {
			req.remotefilepathhash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
		}

		httpreq, err := zboxutil.NewDownloadRequest(req.blobber.Baseurl, req.allocationTx)
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error creating download request")}
			return
		}

		header := &DownloadRequestHeader{}
		header.PathHash = req.remotefilepathhash

		if req.rxPay {
			header.RxPay = req.rxPay // pay oneself
		}
		header.BlockNum = req.blockNum
		header.NumBlocks = req.numBlocks
		header.ReadMarker = rmData

		if req.authTicket != nil {
			header.AuthToken, _ = json.Marshal(req.authTicket) //nolint: errcheck
		}
		if len(req.contentMode) > 0 {
			header.DownloadMode = req.contentMode
		}

		ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
		shouldRetry := false

		header.ToHeader(httpreq)

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

					if len(req.encryptedKey) > 0 {
						if req.authTicket != nil {
							// ReEncryptionHeaderSize for the additional header bytes for ReEncrypt,  where chunk_size - EncryptionHeaderSize is the encrypted data size
							rspData.BlockChunks = req.splitData(response, req.chunkSize-EncryptionHeaderSize+ReEncryptionHeaderSize)
						} else {
							rspData.BlockChunks = req.splitData(response, req.chunkSize)
						}
					} else {
						rspData.BlockChunks = req.splitData(response, req.chunkSize)
					}
					rspData.RawData = []byte{}
					incBlobberReadCtr(req.blobber.ID, req.numBlocks)
					req.result <- &rspData
					return nil
				}

				if !rspData.Success && rspData.LatestRM != nil && rspData.LatestRM.ReadCounter >= getBlobberReadCtr(req.blobber.ID) {
					Logger.Info("Will be retrying download")
					setBlobberReadCtr(req.blobber.ID, rspData.LatestRM.ReadCounter)
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
