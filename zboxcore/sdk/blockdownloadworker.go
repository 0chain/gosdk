package sdk

import (
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
	l "github.com/0chain/gosdk/zboxcore/logger"
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
		start := time.Now()
		var httpreq *http.Request
		httpreq, err = zboxutil.NewDownloadRequest(req.blobber.Baseurl, req.allocationID, req.allocationTx)
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
			elapsedDownloadReqBlobber := time.Since(start).Milliseconds()
			var rspData downloadBlock
			respBody := make([]byte, int(req.numBlocks+10)*req.chunkSize)
			n, err := resp.Body.Read(respBody)
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			respBody = respBody[:n]
			elapsedReadBody := time.Since(start).Milliseconds() - elapsedDownloadReqBlobber
			if resp.StatusCode != http.StatusOK {
				zlogger.Logger.Debug(fmt.Sprintf("downloadBlobberBlock FAIL - blobberID: %v, clientID: %v, blockNum: %d, retry: %d, response: %v", req.blobber.ID, client.GetClientID(), header.BlockNum, retry, string(respBody)))
				if err = json.Unmarshal(respBody, &rspData); err == nil {
					return errors.New("download_error", fmt.Sprintf("Response status: %d, Error: %v,", resp.StatusCode, rspData.err))
				}
				return errors.New("response_error", string(respBody))
			}

			dR := downloadResponse{}
			err = json.Unmarshal(respBody, &dR)
			if err != nil {
				return err
			}
			elapsedUnmarshal := time.Since(start).Milliseconds() - elapsedReadBody - elapsedDownloadReqBlobber
			if req.contentMode == DOWNLOAD_CONTENT_FULL && req.shouldVerify {
				now := time.Now()
				vmp := util.MerklePathForMultiLeafVerification{
					Nodes:    dR.Nodes,
					Index:    dR.Indexes,
					RootHash: req.blobberFile.validationRoot,
					DataSize: req.blobberFile.size,
				}
				err = vmp.VerifyMultipleBlocks(dR.Data)
				if err != nil {
					return errors.New("merkle_path_verification_error", err.Error())
				}
				l.Logger.Info("[verifyMultiBlock]", time.Since(now).Milliseconds())
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

			zlogger.Logger.Debug(fmt.Sprintf("downloadBlobberBlock 200 OK: blobberID: %v, clientID: %v, blockNum: %d, downloadReqBlobber: %v,readBody: %v,unmarshalJSON: %v, totalTime: %v", req.blobber.ID, client.GetClientID(), header.BlockNum, elapsedDownloadReqBlobber, elapsedReadBody, elapsedUnmarshal, time.Since(start).Milliseconds()))

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
