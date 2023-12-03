package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
			for i := 0; i < workerCount; i++ {
				blobberChan := downloadBlockChan[blobber.ID]
				go startBlockDownloadWorker(blobberChan)
			}
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

		var httpreq *http.Request
		httpreq, err = zboxutil.NewDownloadRequest(req.blobber.Baseurl, req.allocationID, req.allocationTx)
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error creating download request")}
			return
		}

		header := &DownloadRequestHeader{}
		isISO := checkISO8859_1([]byte(req.remotefilepathhash))
		if !isISO {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: fmt.Errorf("Non ISO8859_1 characters in path hash %s", req.remotefilepathhash)}
			return
		}
		header.BlockNum = req.blockNum
		isISO = checkISO8859_1([]byte(strconv.FormatInt(req.blockNum, 10)))
		if !isISO {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: fmt.Errorf("Non ISO8859_1 characters in block number %d", req.blockNum)}
			return
		}
		header.NumBlocks = req.numBlocks
		isISO = checkISO8859_1([]byte(strconv.FormatInt(req.numBlocks, 10)))
		if !isISO {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: fmt.Errorf("Non ISO8859_1 characters in number of blocks")}
			return
		}
		header.VerifyDownload = req.shouldVerify
		header.ConnectionID = req.connectionID
		isISO = checkISO8859_1([]byte(req.connectionID))
		if !isISO {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: fmt.Errorf("Non ISO8859_1 characters in connection id %s", req.connectionID)}
			return
		}
		if req.authTicket != nil {
			header.AuthToken, _ = json.Marshal(req.authTicket) //nolint: errcheck
			isISO = checkISO8859_1(header.AuthToken)
			if !isISO {
				req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: fmt.Errorf("Non ISO8859_1 characters in auth token: %s", string(header.AuthToken))}
				return
			}
		}
		if len(req.contentMode) > 0 {
			header.DownloadMode = req.contentMode
			isISO = checkISO8859_1([]byte(req.contentMode))
			if !isISO {
				req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: fmt.Errorf("Non ISO8859_1 characters in download mode %s", req.contentMode)}
			}
			return
		}
		ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
		shouldRetry := false

		header.ToHeader(httpreq)

		zlogger.Logger.Debug(fmt.Sprintf("downloadBlobberBlock - blobberID: %v, clientID: %v, blockNum: %d", req.blobber.ID, client.GetClientID(), header.BlockNum))

		err = zboxutil.HttpDo(ctx, cncl, httpreq, func(resp *http.Response, err error) error {
			if err != nil {
				return err
			}
			if resp.Body != nil {
				defer resp.Body.Close()
			}
			if req.chunkSize == 0 {
				req.chunkSize = CHUNK_SIZE
			}
			var rspData downloadBlock
			respLen := resp.Header.Get("Content-Length")
			var respBody []byte
			if respLen != "" {
				len, err := strconv.Atoi(respLen)
				zlogger.Logger.Info("respLen", len)
				if err != nil {
					zlogger.Logger.Error("respLen convert error: ", err)
					return err
				}
				respBody, err = readBody(resp.Body, len)
				if err != nil {
					zlogger.Logger.Error("respBody read error: ", err)
					return err
				}
			} else {
				respBody, err = readBody(resp.Body, int(req.numBlocks)*req.chunkSize)
				if err != nil {
					zlogger.Logger.Error("respBody read error: ", err)
					return err
				}
			}
			if resp.StatusCode != http.StatusOK {
				zlogger.Logger.Debug(fmt.Sprintf("downloadBlobberBlock FAIL - blobberID: %v, clientID: %v, blockNum: %d, retry: %d, response: %v", req.blobber.ID, client.GetClientID(), header.BlockNum, retry, string(respBody)))
				if err = json.Unmarshal(respBody, &rspData); err == nil {
					return errors.New("download_error", fmt.Sprintf("Response status: %d, Error: %v,", resp.StatusCode, rspData.err))
				}
				return errors.New("response_error", string(respBody))
			}

			dR := downloadResponse{}
			contentType := resp.Header.Get("Content-Type")
			if contentType == "application/json" {
				err = json.Unmarshal(respBody, &dR)
				if err != nil {
					return err
				}
			} else {
				dR.Data = respBody
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

func readBody(r io.Reader, size int) ([]byte, error) {
	b := make([]byte, 0, size)
	for {
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}
	}
}

func checkISO8859_1(data []byte) bool {
	for _, b := range data {
		if b > 0x7F {
			return false
		}
	}
	return true
}
