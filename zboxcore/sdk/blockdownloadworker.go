package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	zlogger "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

const (
	NotEnoughTokens = "not_enough_tokens"
	LockExists      = "lock_exists"
)

type BlockDownloadRequest struct {
	blobber            *blockchain.StorageNode
	allocationID       string
	allocationTx       string
	allocOwnerID       string
	blobberIdx         int
	remotefilepath     string
	remotefilepathhash string
	chunkSize          int
	blockNum           int64
	encryptedKey       string
	contentMode        string
	numBlocks          int64
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
	var err error
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
		rm.OwnerID = req.allocOwnerID
		rm.Timestamp = common.Now()
		rm.ReadCounter = getBlobberReadCtr(req.allocationID, req.blobber.ID) + req.numBlocks
		err = rm.Sign()
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error: Signing readmarker failed")}
			return
		}
		var rmData []byte
		rmData, err = json.Marshal(rm)
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error creating readmarker")}
			return
		}
		if len(req.remotefilepath) > 0 {
			req.remotefilepathhash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
		}

		var httpreq *http.Request
		httpreq, err = zboxutil.NewDownloadRequest(req.blobber.Baseurl, req.allocationTx)
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error creating download request")}
			return
		}

		header := &DownloadRequestHeader{}
		header.PathHash = req.remotefilepathhash

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

			var rspData downloadBlock

			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusOK {
				if err = json.Unmarshal(respBody, &rspData); err == nil && rspData.LatestRM != nil {
					if err := rm.ValidateWithOtherRM(rspData.LatestRM); err != nil {
						retry = 3
						return err
					}

					if rspData.LatestRM.ReadCounter >= getBlobberReadCtr(req.allocationID, req.blobber.ID) {
						zlogger.Logger.Info("Will be retrying download")
						setBlobberReadCtr(req.allocationID, req.blobber.ID, rspData.LatestRM.ReadCounter)
						shouldRetry = true
						return errors.New("stale_read_marker", "readmarker counter is not in sync with latest counter")
					}

					return nil

				}

				if bytes.Contains(respBody, []byte(NotEnoughTokens)) {
					shouldRetry, retry = false, 3 // don't repeat
					req.blobber.SetSkip(true)
					return errors.New(NotEnoughTokens, "")
				}

				if bytes.Contains(respBody, []byte(LockExists)) {
					zlogger.Logger.Debug("Lock exists error.")
					shouldRetry = true
					return errors.New(LockExists, string(respBody))
				}

				return errors.New("response_error", string(respBody))
			}

			rspData.idx = req.blobberIdx
			rspData.Success = true

			if req.encryptedKey != "" {
				if req.authTicket != nil {
					// ReEncryptionHeaderSize for the additional header bytes for ReEncrypt,  where chunk_size - EncryptionHeaderSize is the encrypted data size
					rspData.BlockChunks = req.splitData(respBody, req.chunkSize-EncryptionHeaderSize+ReEncryptionHeaderSize)
				} else {
					rspData.BlockChunks = req.splitData(respBody, req.chunkSize)
				}
			} else {
				rspData.BlockChunks = req.splitData(respBody, req.chunkSize)
			}

			incBlobberReadCtr(req.allocationID, req.blobber.ID, req.numBlocks)
			req.result <- &rspData
			return nil
		})

		if err != nil {
			if shouldRetry {
				retry = 0
				shouldRetry = false
				zlogger.Logger.Debug("Retrying for Error occurred: ", err)
				continue
			}
			if retry >= 3 {
				req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: err}
				return
			}
			retry++
			continue
		}
		return
	}

	req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: err}

}

func AddBlockDownloadReq(req *BlockDownloadRequest) {
	downloadBlockChan[req.blobber.ID] <- req
}
