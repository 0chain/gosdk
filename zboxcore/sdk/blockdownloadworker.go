package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"syscall"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	zlogger "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/hitenjain14/fasthttp"
	"golang.org/x/sync/semaphore"
)

const (
	LockExists     = "lock_exists"
	RateLimitError = "rate_limit_error"
)

type BlockDownloadRequest struct {
	blobber            *blockchain.StorageNode
	allocationID       string
	allocationTx       string
	allocOwnerID       string
	blobberIdx         int
	maskIdx            int
	remotefilepath     string
	remotefilepathhash string
	chunkSize          int
	blockNum           int64
	encryptedKey       string
	contentMode        string
	numBlocks          int64
	authTicket         *marker.AuthTicket
	ctx                context.Context
	result             chan *downloadBlock
	shouldVerify       bool
	connectionID       string
	respBuf            []byte
}

type downloadResponse struct {
	Nodes   [][][]byte
	Indexes [][]int
	Data    []byte
}

type downloadBlock struct {
	BlockChunks [][]byte
	Success     bool               `json:"success"`
	LatestRM    *marker.ReadMarker `json:"latest_rm"`
	idx         int
	maskIdx     int
	err         error
	timeTaken   int64
}

var downloadBlockChan map[string]chan *BlockDownloadRequest
var initDownloadMutex sync.Mutex

func InitBlockDownloader(blobbers []*blockchain.StorageNode, workerCount int) {
	initDownloadMutex.Lock()
	defer initDownloadMutex.Unlock()
	if downloadBlockChan == nil {
		downloadBlockChan = make(map[string]chan *BlockDownloadRequest)
	}

	for _, blobber := range blobbers {
		if _, ok := downloadBlockChan[blobber.ID]; !ok {
			downloadBlockChan[blobber.ID] = make(chan *BlockDownloadRequest, workerCount)
			go startBlockDownloadWorker(downloadBlockChan[blobber.ID], workerCount)
		}
	}
}

func startBlockDownloadWorker(blobberChan chan *BlockDownloadRequest, workers int) {
	sem := semaphore.NewWeighted(int64(workers))
	fastClient := zboxutil.GetFastHTTPClient()
	for {
		blockDownloadReq, open := <-blobberChan
		if !open {
			break
		}
		if err := sem.Acquire(blockDownloadReq.ctx, 1); err != nil {
			blockDownloadReq.result <- &downloadBlock{Success: false, idx: blockDownloadReq.blobberIdx, err: err}
			continue
		}
		go func() {
			blockDownloadReq.downloadBlobberBlock(fastClient)
			sem.Release(1)
		}()
	}
}

func splitData(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, common.MustAddInt(len(buf)/lim, 1))
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:])
	}
	return chunks
}

func (req *BlockDownloadRequest) downloadBlobberBlock(fastClient *fasthttp.Client) {
	if req.numBlocks <= 0 {
		req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.New("invalid_request", "Invalid number of blocks for download")}
		return
	}
	retry := 0
	var err error
	for retry < 3 {
		if len(req.remotefilepath) > 0 {
			req.remotefilepathhash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
		}

		httpreq, err := zboxutil.NewFastDownloadRequest(req.blobber.Baseurl, req.allocationID, req.allocationTx)
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error creating download request")}
			return
		}

		header := &DownloadRequestHeader{}
		header.PathHash = req.remotefilepathhash
		header.BlockNum = req.blockNum
		header.NumBlocks = req.numBlocks
		header.VerifyDownload = req.shouldVerify
		header.ConnectionID = req.connectionID
		header.Version = "v2"

		if req.authTicket != nil {
			header.AuthToken, _ = json.Marshal(req.authTicket) //nolint: errcheck
		}
		if len(req.contentMode) > 0 {
			header.DownloadMode = req.contentMode
		}
		if req.chunkSize == 0 {
			req.chunkSize = CHUNK_SIZE
		}
		shouldRetry := false

		header.ToFastHeader(httpreq)

		err = func() error {
			now := time.Now()
			statuscode, respBuf, err := fastClient.GetWithRequest(httpreq, req.respBuf)
			fasthttp.ReleaseRequest(httpreq)
			timeTaken := time.Since(now).Milliseconds()
			if err != nil {
				zlogger.Logger.Error("Error downloading block: ", err)
				if errors.Is(err, fasthttp.ErrConnectionClosed) || errors.Is(err, syscall.EPIPE) {
					shouldRetry = true
					return errors.New("connection_closed", "Connection closed")
				}
				return err
			}

			if statuscode == http.StatusTooManyRequests {
				shouldRetry = true
				time.Sleep(time.Second * 2)
				return errors.New(RateLimitError, "Rate limit error")
			}

			if statuscode == http.StatusInternalServerError {
				shouldRetry = true
				return errors.New("internal_server_error", "Internal server error")
			}

			var rspData downloadBlock
			if statuscode != http.StatusOK {
				zlogger.Logger.Error(fmt.Sprintf("downloadBlobberBlock FAIL - blobberID: %v, clientID: %v, blockNum: %d, retry: %d, response: %v", req.blobber.ID, client.GetClientID(), header.BlockNum, retry, string(respBuf)))
				if err = json.Unmarshal(respBuf, &rspData); err == nil {
					return errors.New("download_error", fmt.Sprintf("Response status: %d, Error: %v,", statuscode, rspData.err))
				}
				return errors.New("response_error", string(respBuf))
			}

			dR := downloadResponse{}
			if req.shouldVerify {
				err = json.Unmarshal(respBuf, &dR)
				if err != nil {
					return err
				}
			} else {
				dR.Data = respBuf
			}
			if req.contentMode == DOWNLOAD_CONTENT_FULL && req.shouldVerify {
				zlogger.Logger.Info("verifying multiple blocks")
			}

			rspData.idx = req.blobberIdx
			rspData.maskIdx = req.maskIdx
			rspData.timeTaken = timeTaken
			rspData.Success = true

			if req.encryptedKey != "" {
				if req.authTicket != nil {
					// ReEncryptionHeaderSize for the additional header bytes for ReEncrypt,  where chunk_size - EncryptionHeaderSize is the encrypted data size
					rspData.BlockChunks = splitData(dR.Data, req.chunkSize-EncryptionHeaderSize+ReEncryptionHeaderSize)
				} else {
					rspData.BlockChunks = splitData(dR.Data, req.chunkSize)
				}
			} else {
				if req.chunkSize == 0 {
					req.chunkSize = CHUNK_SIZE
				}
				rspData.BlockChunks = splitData(dR.Data, req.chunkSize)
			}

			zlogger.Logger.Debug(fmt.Sprintf("downloadBlobberBlock 200 OK: blobberID: %v, clientID: %v, blockNum: %d", req.blobber.ID, client.GetClientID(), header.BlockNum))

			req.result <- &rspData
			return nil
		}()

		if err != nil {
			if shouldRetry {
				if retry >= 3 {
					req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: err}
					return
				}
				shouldRetry = false
				zlogger.Logger.Debug("Retrying for Error occurred: ", err)
				retry++
				continue
			} else {
				req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: err, maskIdx: req.maskIdx}
			}
		}
		return
	}

	req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: err, maskIdx: req.maskIdx}

}

func AddBlockDownloadReq(ctx context.Context, req *BlockDownloadRequest, rb zboxutil.DownloadBuffer, effectiveBlockSize int) {
	if rb != nil {
		reqCtx, cncl := context.WithTimeout(ctx, (time.Second * 45))
		defer cncl()
		req.respBuf = rb.RequestChunk(reqCtx, int(req.blockNum))
		if len(req.respBuf) == 0 {
			req.respBuf = make([]byte, int(req.numBlocks)*effectiveBlockSize)
		}
	} else {
		req.respBuf = make([]byte, int(req.numBlocks)*effectiveBlockSize)
	}
	downloadBlockChan[req.blobber.ID] <- req
}
