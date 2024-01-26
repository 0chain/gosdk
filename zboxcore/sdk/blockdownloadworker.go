package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	zlogger "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/valyala/fasthttp"
	"golang.org/x/sync/semaphore"
)

const (
	LockExists     = "lock_exists"
	RateLimitError = "rate_limit_error"
)

type BlockDownloadRequest struct {
	blobber            *blockchain.StorageNode
	blobberFile        *blobberFile
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
	err         error
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
			blockDownloadReq.downloadBlobberBlock()
			sem.Release(1)
		}()
	}
}

func (req *BlockDownloadRequest) splitData(buf []byte, lim int) [][]byte {
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

func (req *BlockDownloadRequest) downloadBlobberBlock() {
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

		// var httpreq *http.Request
		// httpreq, err = zboxutil.NewDownloadRequest(req.blobber.Baseurl, req.allocationID, req.allocationTx)
		// if err != nil {
		// 	req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error creating download request")}
		// 	return
		// }

		httpreq, err := zboxutil.NewFastDownloadRequest(req.blobber.Baseurl, req.allocationID, req.allocationTx)
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error creating download request")}
			return
		}
		defer fasthttp.ReleaseRequest(httpreq)

		header := &DownloadRequestHeader{}
		header.PathHash = req.remotefilepathhash
		header.BlockNum = req.blockNum
		header.NumBlocks = req.numBlocks
		header.VerifyDownload = req.shouldVerify
		header.ConnectionID = req.connectionID

		if req.authTicket != nil {
			header.AuthToken, _ = json.Marshal(req.authTicket) //nolint: errcheck
		}
		if len(req.contentMode) > 0 {
			header.DownloadMode = req.contentMode
		}

		// ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
		shouldRetry := false

		header.ToFastHeader(httpreq)

		zlogger.Logger.Debug(fmt.Sprintf("downloadBlobberBlock - blobberID: %v, clientID: %v, blockNum: %d", req.blobber.ID, client.GetClientID(), header.BlockNum))

		// err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
		err = func() error {
			// if err != nil {
			// 	return err
			// }
			resp := fasthttp.AcquireResponse()
			defer fasthttp.ReleaseResponse(resp)
			buf := bytes.NewBuffer(req.respBuf)
			resp.SetBodyStream(buf, -1)
			err = fasthttp.DoTimeout(httpreq, resp, time.Second*30)
			// if resp.Body != nil {
			// 	defer resp.Body.Close()
			// }
			if req.chunkSize == 0 {
				req.chunkSize = CHUNK_SIZE
			}

			if resp.StatusCode() == http.StatusTooManyRequests {
				shouldRetry = true
				time.Sleep(time.Second * 2)
				return errors.New(RateLimitError, "Rate limit error")
			}

			if resp.StatusCode() == http.StatusInternalServerError {
				shouldRetry = true
				return errors.New("internal_server_error", "Internal server error")
			}

			var rspData downloadBlock
			// if req.shouldVerify {
			// 	req.respBuf, err = io.ReadAll(resp.Body)
			// 	if err != nil {
			// 		zlogger.Logger.Error("respBody read error: ", err)
			// 		return err
			// 	}
			// } else {
			// 	req.respBuf, err = readBody(resp.Body, req.respBuf)
			// 	if err != nil {
			// 		zlogger.Logger.Error("respBody read error: ", err)
			// 		return err
			// 	}
			// }
			if resp.StatusCode() != http.StatusOK {
				zlogger.Logger.Debug(fmt.Sprintf("downloadBlobberBlock FAIL - blobberID: %v, clientID: %v, blockNum: %d, retry: %d, response: %v", req.blobber.ID, client.GetClientID(), header.BlockNum, retry, string(req.respBuf)))
				if err = json.Unmarshal(req.respBuf, &rspData); err == nil {
					return errors.New("download_error", fmt.Sprintf("Response status: %d, Error: %v,", resp.StatusCode, rspData.err))
				}
				return errors.New("response_error", string(req.respBuf))
			}

			dR := downloadResponse{}
			contentType := resp.Header.Peek("Content-Type")
			respLen := resp.Header.ContentLength()
			if string(contentType) == "application/json" {
				err = json.Unmarshal(req.respBuf, &dR)
				if err != nil {
					return err
				}
			} else {
				dR.Data = req.respBuf[:respLen]
			}
			if req.contentMode == DOWNLOAD_CONTENT_FULL && req.shouldVerify {

				vmp := util.MerklePathForMultiLeafVerification{
					Nodes:    dR.Nodes,
					Index:    dR.Indexes,
					RootHash: req.blobberFile.validationRoot,
					DataSize: req.blobberFile.size,
				}
				zlogger.Logger.Info("verifying multiple blocks")
				err = vmp.VerifyMultipleBlocks(dR.Data)
				if err != nil {
					return errors.New("merkle_path_verification_error", err.Error())
				}
			}

			rspData.idx = req.blobberIdx
			rspData.Success = true

			if req.encryptedKey != "" {
				if req.authTicket != nil {
					// ReEncryptionHeaderSize for the additional header bytes for ReEncrypt,  where chunk_size - EncryptionHeaderSize is the encrypted data size
					rspData.BlockChunks = req.splitData(dR.Data, req.chunkSize-EncryptionHeaderSize+ReEncryptionHeaderSize)
				} else {
					rspData.BlockChunks = req.splitData(dR.Data, req.chunkSize)
				}
			} else {
				if req.chunkSize == 0 {
					req.chunkSize = CHUNK_SIZE
				}
				rspData.BlockChunks = req.splitData(dR.Data, req.chunkSize)
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
				req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: err}
			}
		}
		return
	}

	req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: err}

}

func AddBlockDownloadReq(ctx context.Context, req *BlockDownloadRequest, rb *zboxutil.DownloadBuffer, effectiveBlockSize int) {
	if rb != nil {
		reqCtx, cncl := context.WithTimeout(ctx, (time.Second * 10))
		defer cncl()
		req.respBuf = rb.RequestChunk(reqCtx, int(req.blockNum/req.numBlocks))
		if len(req.respBuf) == 0 {
			req.respBuf = make([]byte, int(req.numBlocks)*effectiveBlockSize)
		}
	}
	downloadBlockChan[req.blobber.ID] <- req
}

func readBody(r io.Reader, b []byte) ([]byte, error) {
	start := 0
	if len(b) == 0 {
		return nil, fmt.Errorf("readBody: empty buffer")
	}
	for {
		n, err := r.Read(b[start:])
		start += n
		if err != nil {
			if err == io.EOF {
				err = nil
				b = b[:start]
			}
			return b, err
		}
	}
}
