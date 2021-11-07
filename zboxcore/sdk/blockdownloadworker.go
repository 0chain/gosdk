package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	blobbergrpc "github.com/0chain/blobber/code/go/0chain.net/blobbercore/blobbergrpc/proto"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/clients/blobberClient"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/fileref"
	. "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"strconv"
	"strings"
	"sync"
)

type BlockDownloadRequest struct {
	blobber            *blockchain.StorageNode
	allocationID       string
	allocationTx       string
	blobberIdx         int
	remotefilepath     string
	remotefilepathhash string
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

		rmData, err := json.Marshal(rm)
		if err != nil {
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error creating readmarker")}
			return
		}

		if len(req.remotefilepath) > 0 {
			req.remotefilepathhash = fileref.GetReferenceLookup(req.allocationID, req.remotefilepath)
		}

		downloadReq := new(blobbergrpc.DownloadFileRequest)

		downloadReq.Allocation = req.allocationTx
		downloadReq.Path = req.remotefilepath
		downloadReq.PathHash = req.remotefilepathhash

		if req.rxPay {
			downloadReq.RxPay = "true" // pay oneself
		}

		downloadReq.BlockNum = strconv.FormatInt(req.blockNum, 10)
		downloadReq.NumBlocks = strconv.FormatInt(req.numBlocks, 10)
		downloadReq.ReadMarker = string(rmData)

		if req.authTicket != nil {
			authTicketBytes, _ := json.Marshal(req.authTicket)
			downloadReq.AuthToken = string(authTicketBytes)
		}
		if len(req.contentMode) > 0 {
			downloadReq.Content = req.contentMode
		}
		// TODO: Fix the timeout
		//ctx, cncl := context.WithTimeout(req.ctx, (time.Second * 30))
		shouldRetry := false
		respBytes, err := blobberClient.DownloadObject(req.blobber.Baseurl, downloadReq)
		if err != nil {
			err = fmt.Errorf("response Error: %s", err)
			if strings.Contains(err.Error(), "not_enough_tokens") {
				shouldRetry, retry = false, 3 // don't repeat
				req.blobber.SetSkip(true)
			}
		} else {
			var rspData downloadBlock
			rspData.idx = req.blobberIdx

			err = json.Unmarshal(respBytes, &rspData)
			if err != nil {
				rspData.Success = true
				rspData.RawData = respBytes
				if len(req.encryptedKey) > 0 {
					// 256 for the additional header bytes,  where chunk_size - 2 * 1024 is the encrypted data size
					chunks := req.splitData(rspData.RawData, fileref.CHUNK_SIZE-2*1024+256)
					rspData.BlockChunks = chunks
				} else {
					chunks := req.splitData(rspData.RawData, fileref.CHUNK_SIZE)
					rspData.BlockChunks = chunks
				}
				rspData.RawData = []byte{}
				incBlobberReadCtr(req.blobber, req.numBlocks)
				req.result <- &rspData
				return
			}

			if !rspData.Success && rspData.LatestRM != nil && rspData.LatestRM.ReadCounter >= getBlobberReadCtr(req.blobber) {
				Logger.Info("Will be retrying download")
				setBlobberReadCtr(req.blobber, rspData.LatestRM.ReadCounter)
				shouldRetry = true
				err = errors.New("", "Need to retry the download")
			}
		}

		if err != nil && (!shouldRetry || retry >= 3) {
			Logger.Error("could not download object-" + req.blobber.Baseurl + " - " + err.Error())
			req.result <- &downloadBlock{Success: false, idx: req.blobberIdx, err: errors.Wrap(err, "Error downloading blobber object")}
			return
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
