package sdk

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/klauspost/reedsolomon"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"golang.org/x/sync/errgroup"
)

const (
	DOWNLOAD_CONTENT_FULL  = "full"
	DOWNLOAD_CONTENT_THUMB = "thumbnail"
)

var (
	extraCount = 2
)

type DownloadRequestOption func(dr *DownloadRequest)

// WithDownloadProgressStorer set download progress storer of download request options.
//   - storer: download progress storer instance, used to store download progress.
func WithDownloadProgressStorer(storer DownloadProgressStorer) DownloadRequestOption {
	return func(dr *DownloadRequest) {
		dr.downloadStorer = storer
	}
}

func WithWorkDir(workdir string) DownloadRequestOption {
	return func(dr *DownloadRequest) {
		dr.workdir = workdir
	}
}

func WithFileCallback(cb func()) DownloadRequestOption {
	return func(dr *DownloadRequest) {
		dr.fileCallback = cb
	}
}

type DownloadRequest struct {
	allocationID       string
	allocationTx       string
	sig                string
	allocOwnerID       string
	allocOwnerPubKey   string
	blobbers           []*blockchain.StorageNode
	datashards         int
	parityshards       int
	remotefilepath     string
	remotefilepathhash string
	fileHandler        sys.File
	localFilePath      string
	startBlock         int64
	endBlock           int64
	chunkSize          int
	numBlocks          int64
	statusCallback     StatusCallback
	ctx                context.Context
	ctxCncl            context.CancelFunc
	authTicket         *marker.AuthTicket
	downloadMask       zboxutil.Uint128
	encryptedKey       string
	isDownloadCanceled bool
	completedCallback  func(remotepath string, remotepathhash string)
	fileCallback       func()
	contentMode        string
	Consensus
	effectiveBlockSize int // blocksize - encryptionOverHead
	ecEncoder          reedsolomon.Encoder
	maskMu             *sync.Mutex
	encScheme          encryption.EncryptionScheme
	shouldVerify       bool
	blocksPerShard     int64
	connectionID       string
	skip               bool
	freeRead           bool
	fRef               *fileref.FileRef
	chunksPerShard     int64
	size               int64
	offset             int64
	bufferMap          map[int]zboxutil.DownloadBuffer
	downloadStorer     DownloadProgressStorer
	workdir            string
	downloadQueue      downloadQueue // Always initialize this queue with max time taken
	isResume           bool
	isEnterprise       bool
}

type downloadPriority struct {
	timeTaken  int64
	blobberIdx int
}

type downloadQueue []downloadPriority

func (pq downloadQueue) Len() int { return len(pq) }

func (pq downloadQueue) Less(i, j int) bool {
	return pq[i].timeTaken < pq[j].timeTaken
}

type DownloadProgress struct {
	ID               string `json:"id"`
	LastWrittenBlock int    `json:"last_block"`
	numBlocks        int    `json:"-"`
}
type blockData struct {
	blockNum int
	data     [][][]byte
}

func (req *DownloadRequest) removeFromMask(pos uint64) {
	req.maskMu.Lock()
	req.downloadMask = req.downloadMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
	req.maskMu.Unlock()
}

func (req *DownloadRequest) getBlocksDataFromBlobbers(startBlock, totalBlock int64, timeRequest bool) ([][][]byte, error) {
	shards := make([][][]byte, totalBlock)
	for i := range shards {
		shards[i] = make([][]byte, len(req.blobbers))
	}

	mask := req.downloadMask
	requiredDownloads := req.consensusThresh
	var (
		remainingMask  zboxutil.Uint128
		failed         int
		err            error
		downloadErrors []string
	)

	curReqDownloads := requiredDownloads
	for {
		remainingMask, failed, downloadErrors, err = req.downloadBlock(
			startBlock, totalBlock, mask, curReqDownloads, shards, timeRequest)
		if err != nil {
			return nil, err
		}
		if failed == 0 || (timeRequest && mask.CountOnes()-failed >= requiredDownloads) {
			break
		}

		if failed > remainingMask.CountOnes() {
			return nil, errors.New("download_failed",
				fmt.Sprintf("%d failed blobbers exceeded %d remaining blobbers."+
					" Download errors: %s",
					failed, remainingMask.CountOnes(), strings.Join(downloadErrors, " ")))
		}

		curReqDownloads = failed
		mask = remainingMask
	}
	return shards, err
}

// getBlocksData will get data blocks for some interval from minimal blobers and aggregate them and
// return to the caller
func (req *DownloadRequest) getBlocksData(startBlock, totalBlock int64, timeRequest bool) ([][][]byte, error) {

	shards, err := req.getBlocksDataFromBlobbers(startBlock, totalBlock, timeRequest)
	if err != nil {
		return nil, err
	}

	// erasure decoding
	// Can we benefit from goroutine for erasure decoding??
	// c := req.datashards * req.effectiveBlockSize
	// data := make([]byte, req.datashards*req.effectiveBlockSize*int(totalBlock))
	for i := range shards {
		err = req.decodeEC(shards[i])
		if err != nil {
			return nil, err
		}

	}
	return shards, nil
}

// downloadBlock This function will add download requests to the download channel which picks up
// download requests and processes it.
// This function will fill up `shards` in respective position and also return failed number of
// blobbers along with remainingMask that are the blobbers that are not yet requested.
func (req *DownloadRequest) downloadBlock(
	startBlock, totalBlock int64,
	mask zboxutil.Uint128, requiredDownloads int,
	shards [][][]byte, timeRequest bool) (zboxutil.Uint128, int, []string, error) {

	var remainingMask zboxutil.Uint128
	activeBlobbers := mask.CountOnes()
	if activeBlobbers < requiredDownloads {
		return zboxutil.NewUint128(0), 0, nil, errors.New("insufficient_blobbers",
			fmt.Sprintf("Required downloads %d, remaining active blobber %d",
				req.consensusThresh, activeBlobbers))
	}
	if timeRequest {
		requiredDownloads = activeBlobbers
	}
	rspCh := make(chan *downloadBlock, requiredDownloads)

	var (
		pos          uint64
		c            int
		skipDownload bool
	)

	for i := mask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		if c == requiredDownloads {
			remainingMask = i
			break
		}

		pos = uint64(i.TrailingZeros())
		blobberIdx := req.downloadQueue[pos].blobberIdx
		blockDownloadReq := &BlockDownloadRequest{
			allocationID:       req.allocationID,
			allocationTx:       req.allocationTx,
			allocOwnerID:       req.allocOwnerID,
			authTicket:         req.authTicket,
			blobber:            req.blobbers[blobberIdx],
			blobberIdx:         blobberIdx,
			maskIdx:            int(pos),
			chunkSize:          req.chunkSize,
			blockNum:           startBlock,
			contentMode:        req.contentMode,
			result:             rspCh,
			ctx:                req.ctx,
			remotefilepath:     req.remotefilepath,
			remotefilepathhash: req.remotefilepathhash,
			numBlocks:          totalBlock,
			encryptedKey:       req.encryptedKey,
			connectionID:       req.connectionID,
		}

		if blockDownloadReq.blobber.IsSkip() {
			rspCh <- &downloadBlock{
				Success: false,
				idx:     blockDownloadReq.blobberIdx,
				err:     errors.New("", "skip blobber by previous errors")}
			skipDownload = true
		}

		if !skipDownload {
			if req.shouldVerify {
				go AddBlockDownloadReq(req.ctx, blockDownloadReq, nil, req.effectiveBlockSize)
			} else {
				go AddBlockDownloadReq(req.ctx, blockDownloadReq, req.bufferMap[blobberIdx], req.effectiveBlockSize)
			}
		}

		c++
	}

	var failed int32
	downloadErrors := make([]string, requiredDownloads)
	wg := &sync.WaitGroup{}
	for i := 0; i < requiredDownloads; i++ {
		result := <-rspCh
		wg.Add(1)
		go func(i int) {
			var err error
			defer func() {
				if err != nil {
					totalFail := atomic.AddInt32(&failed, 1)
					// if first request remove from end as we will convert the slice into heap
					if timeRequest {
						req.removeFromMask(uint64(activeBlobbers - int(totalFail)))
					} else {
						req.removeFromMask(uint64(result.maskIdx))
					}
					downloadErrors[i] = fmt.Sprintf("Error %s from %s",
						err.Error(), req.blobbers[result.idx].Baseurl)
					logger.Logger.Error(err)
					if req.bufferMap != nil && req.bufferMap[result.idx] != nil {
						req.bufferMap[result.idx].ReleaseChunk(int(req.startBlock))
					}
				} else if timeRequest {
					req.downloadQueue[result.maskIdx].timeTaken = result.timeTaken
				}
				wg.Done()
			}()
			if !result.Success {
				err = fmt.Errorf("Unsuccessful download. Error: %v", result.err)
				return
			}
			err = req.fillShards(shards, result)
		}(i)
	}

	wg.Wait()
	return remainingMask, int(failed), downloadErrors, nil
}

// decodeEC will reconstruct shards and verify it
func (req *DownloadRequest) decodeEC(shards [][]byte) (err error) {
	err = req.ecEncoder.ReconstructData(shards)
	if err != nil {
		return
	}
	// c := len(shards[0])
	// data = make([]byte, req.datashards*c)
	// for i := 0; i < req.datashards; i++ {
	// 	index := i * c
	// 	copy(data[index:index+c], shards[i])
	// }
	return nil
}

//shards -> shards[i][data]

// fillShards will fill `shards` with data from blobbers that belongs to specific
// blockNumber and blobber's position index in an allocation
func (req *DownloadRequest) fillShards(shards [][][]byte, result *downloadBlock) (err error) {

	for i := 0; i < len(result.BlockChunks); i++ {
		var data []byte
		if req.encryptedKey != "" {
			data, err = req.getDecryptedData(result, i)
			if err != nil {
				return err
			}
		} else {
			data = result.BlockChunks[i]
		}
		if i >= len(shards) || len(shards[i]) <= result.idx {
			l.Logger.Error("Invalid shard index", result.idx, len(shards))
			return errors.New("invalid_shard_index", fmt.Sprintf("Invalid shard index %d shard len: %d shard block len: %d", result.idx, len(shards), i))
		}
		shards[i][result.idx] = data
	}
	return
}

// getDecryptedData will decrypt encrypted data and return it.
func (req *DownloadRequest) getDecryptedData(result *downloadBlock, blockNum int) (data []byte, err error) {
	if req.authTicket != nil {
		return req.getDecryptedDataForAuthTicket(result, blockNum)
	}

	headerBytes := result.BlockChunks[blockNum][:EncryptionHeaderSize]
	headerBytes = bytes.Trim(headerBytes, "\x00")

	if len(headerBytes) != EncryptionHeaderSize {
		logger.Logger.Error("Block has invalid header", req.blobbers[result.idx].Baseurl)
		return nil, errors.New(
			"invalid_header",
			fmt.Sprintf("Block from %s has invalid header. Required header size: %d, got %d",
				req.blobbers[result.idx].Baseurl, EncryptionHeaderSize, len(headerBytes)))
	}

	encMsg := &encryption.EncryptedMessage{}
	encMsg.EncryptedData = result.BlockChunks[blockNum][EncryptionHeaderSize:]
	encMsg.MessageChecksum, encMsg.OverallChecksum = string(headerBytes[:128]), string(headerBytes[128:])
	encMsg.EncryptedKey = req.encScheme.GetEncryptedKey()
	decryptedBytes, err := req.encScheme.Decrypt(encMsg)
	if err != nil {
		logger.Logger.Error("Block decryption failed", req.blobbers[result.idx].Baseurl, err)
		return nil, errors.New(
			"decryption_error",
			fmt.Sprintf("Decryption error %s while decrypting data from %s blobber",
				err.Error(), req.blobbers[result.idx].Baseurl))
	}
	return decryptedBytes, nil
}

// getDecryptedDataForAuthTicket will return decrypt shared encrypted data using re-encryption/re-decryption
// mechanism
func (req *DownloadRequest) getDecryptedDataForAuthTicket(result *downloadBlock, blockNum int) (data []byte, err error) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	reEncMessage := &encryption.ReEncryptedMessage{
		D1: suite.Point(),
		D4: suite.Point(),
		D5: suite.Point(),
	}
	err = reEncMessage.Unmarshal(result.BlockChunks[blockNum])
	if err != nil {
		logger.Logger.Error("ReEncrypted Block unmarshall failed", req.blobbers[result.idx].Baseurl, err)
		return nil, err
	}
	decrypted, err := req.encScheme.ReDecrypt(reEncMessage)
	if err != nil {
		logger.Logger.Error("Block redecryption failed", req.blobbers[result.idx].Baseurl, err)
		return nil, err
	}
	return decrypted, nil
}

// processDownload will setup download parameters and downloads data with given
// start block, end block and number of blocks to download in single request.
// This will also write data to the file handler and will verify content by calculating content hash.
func (req *DownloadRequest) processDownload() {
	ctx := req.ctx
	if req.completedCallback != nil {
		defer req.completedCallback(req.remotefilepath, req.remotefilepathhash)
	}
	if req.fileCallback != nil {
		defer func() {
			if !req.skip {
				req.fileCallback()
			}
		}()
	}
	defer req.ctxCncl()
	remotePathCB := req.remotefilepath
	if remotePathCB == "" {
		remotePathCB = req.remotefilepathhash
	}

	var op = OpDownload
	if req.contentMode == DOWNLOAD_CONTENT_THUMB {
		op = opThumbnailDownload
	}
	fRef := req.fRef
	if fRef != nil && fRef.ActualFileHash == emptyFileDataHash {
		logger.Logger.Info("File is empty")
		_, err := req.fileHandler.Write([]byte(emptyFileDataHash))
		if err != nil {
			req.errorCB(errors.Wrap(err, "Write file failed"), remotePathCB)
			return
		}
		req.fileHandler.Sync() //nolint
		if req.statusCallback != nil && !req.skip {
			req.statusCallback.Completed(
				req.allocationID, remotePathCB, fRef.Name, fRef.MimeType, 32, op)
		}
		return
	}
	size, chunksPerShard, blocksPerShard := req.size, req.chunksPerShard, req.blocksPerShard

	now := time.Now()
	err := req.initEC()
	if err != nil {
		logger.Logger.Error(err)
		req.errorCB(
			fmt.Errorf("Error while initializing file ref. Error: %v",
				err), remotePathCB)
		return
	}
	elapsedInitEC := time.Since(now)
	if req.encryptedKey != "" {
		err = req.initEncryption()
		if err != nil {
			req.errorCB(
				fmt.Errorf("Error while initializing encryption"), remotePathCB,
			)
			return
		}
	}
	elapsedInitEncryption := time.Since(now) - elapsedInitEC

	var downloaded int
	startBlock, endBlock, numBlocks := req.startBlock, req.endBlock, req.numBlocks
	// remainingSize should be calculated based on startBlock number
	// otherwise end data will have null bytes.
	remainingSize := size - startBlock*int64(req.effectiveBlockSize)*int64(req.datashards)

	if endBlock*int64(req.effectiveBlockSize)*int64(req.datashards) < req.size {
		remainingSize = blocksPerShard * int64(req.effectiveBlockSize) * int64(req.datashards)
	} else if req.isResume {
		remainingSize = size
	}

	if memFile, ok := req.fileHandler.(*sys.MemFile); ok {
		memFile.InitBuffer(int(remainingSize))
	}

	if req.statusCallback != nil {
		// Started will also initialize progress bar. So without calling this function
		// other callback's call will panic
		req.statusCallback.Started(req.allocationID, remotePathCB, op, int(remainingSize))
	}

	if req.shouldVerify {
		if req.isEnterprise || (req.authTicket != nil && req.encryptedKey != "") {
			req.shouldVerify = false
		}
	}
	n := int((endBlock - startBlock + numBlocks - 1) / numBlocks)

	// Buffered channel to hold the blocks as they are downloaded
	blocks := make(chan blockData, n)

	var (
		actualFileHasher  hash.Hash
		isPREAndWholeFile bool
	)

	if !req.shouldVerify && (startBlock == 0 && endBlock == chunksPerShard) && shouldVerifyHash {
		actualFileHasher = md5.New()
		isPREAndWholeFile = true
	}

	toSync := false
	if _, ok := req.fileHandler.(*sys.MemChanFile); ok {
		toSync = true
	}
	var writerAt bool
	writeAtHandler, ok := req.fileHandler.(io.WriterAt)
	if ok {
		writerAt = true
	}
	bufBlocks := int(numBlocks)
	if n == 1 && endBlock-startBlock < numBlocks {
		bufBlocks = int(endBlock - startBlock)
	}
	if !req.shouldVerify {
		var pos uint64
		req.bufferMap = make(map[int]zboxutil.DownloadBuffer)
		defer func() {
			l.Logger.Debug("Clearing download buffers: ", len(req.bufferMap))
			for ind, rb := range req.bufferMap {
				rb.ClearBuffer()
				delete(req.bufferMap, ind)
			}
			req.bufferMap = nil
		}()
		sz := downloadWorkerCount + extraCount
		if sz > n {
			sz = n
		}

		for i := req.downloadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
			pos = uint64(i.TrailingZeros())
			blobberIdx := int(pos)
			if writerAt {
				req.bufferMap[blobberIdx] = zboxutil.NewDownloadBufferWithChan(sz, bufBlocks, req.effectiveBlockSize)
			} else {
				bufMask := zboxutil.NewDownloadBufferWithMask(sz, bufBlocks, req.effectiveBlockSize)
				bufMask.SetNumBlocks(int(numBlocks))
				req.bufferMap[blobberIdx] = bufMask
			}
		}
	}
	// reset mask to number of active blobbers, not it denotes index of download queue and not blobber index
	activeBlobbers := req.downloadMask.CountOnes()
	req.downloadMask = zboxutil.NewUint128(1).Lsh(uint64(activeBlobbers)).Sub64(1)

	logger.Logger.Info(
		fmt.Sprintf("Downloading file with size: %d from start block: %d and end block: %d. "+
			"Blocks per blobber: %d remainingSize: %d and total requests: %d", size, req.startBlock, req.endBlock, blocksPerShard, remainingSize, n),
	)

	writeCtx, writeCancel := context.WithCancel(ctx)
	defer writeCancel()
	var wg sync.WaitGroup

	if !writerAt {
		wg.Add(1)
		// Handle writing the blocks in order as soon as they are downloaded
		go func() {
			defer wg.Done()
			buffer := make(map[int][][][]byte)
			for i := 0; i < n; i++ {
				select {
				case <-writeCtx.Done():
					goto breakLoop
				default:
				}
				if data, ok := buffer[i]; ok {
					// If the block we need to write next is already in the buffer, write it
					hashWg := &sync.WaitGroup{}
					if isPREAndWholeFile {
						if i == n-1 {
							writeData(actualFileHasher, data, req.datashards, int(remainingSize)) //nolint
							if calculatedFileHash, ok := checkHash(actualFileHasher, fRef, req.contentMode); !ok {
								req.errorCB(fmt.Errorf("Expected actual file hash %s, calculated file hash %s",
									fRef.ActualFileHash, calculatedFileHash), remotePathCB)
								return
							}
						} else {
							hashWg.Add(1)
							go func() {
								writeData(actualFileHasher, data, req.datashards, int(remainingSize)) //nolint
								hashWg.Done()
							}()
						}
					}

					totalWritten, err := writeData(req.fileHandler, data, req.datashards, int(remainingSize))
					if err != nil {
						req.errorCB(errors.Wrap(err, "Write file failed"), remotePathCB)
						return
					}
					if toSync {
						req.fileHandler.Sync() //nolint
					}

					if isPREAndWholeFile {
						hashWg.Wait()
					}
					for _, rb := range req.bufferMap {
						rb.ReleaseChunk(int(startBlock + int64(i)*numBlocks))
					}
					downloaded = downloaded + totalWritten
					remainingSize -= int64(totalWritten)

					if req.statusCallback != nil {
						req.statusCallback.InProgress(req.allocationID, remotePathCB, op, downloaded, nil)
					}

					// Remove the block from the buffer
					delete(buffer, i)
				} else {
					// If the block we need to write next is not in the buffer, wait for it
					for block := range blocks {
						if block.blockNum == i {
							// Write the data
							hashWg := &sync.WaitGroup{}
							if isPREAndWholeFile {
								if i == n-1 {
									writeData(actualFileHasher, block.data, req.datashards, int(remainingSize)) //nolint
									if calculatedFileHash, ok := checkHash(actualFileHasher, fRef, req.contentMode); !ok {
										req.errorCB(fmt.Errorf("Expected actual file hash %s, calculated file hash %s",
											fRef.ActualFileHash, calculatedFileHash), remotePathCB)
										return
									}
								} else {
									hashWg.Add(1)
									go func() {
										writeData(actualFileHasher, block.data, req.datashards, int(remainingSize)) //nolint
										hashWg.Done()
									}()
								}
							}

							totalWritten, err := writeData(req.fileHandler, block.data, req.datashards, int(remainingSize))
							if err != nil {
								req.errorCB(errors.Wrap(err, "Write file failed"), remotePathCB)
								return
							}

							if toSync {
								req.fileHandler.Sync() //nolint
							}

							if isPREAndWholeFile {
								hashWg.Wait()
							}
							for _, rb := range req.bufferMap {
								rb.ReleaseChunk(int(startBlock + int64(i)*numBlocks))
							}

							downloaded = downloaded + totalWritten
							remainingSize -= int64(totalWritten)

							if req.statusCallback != nil {
								req.statusCallback.InProgress(req.allocationID, remotePathCB, op, downloaded, nil)
							}

							break
						} else {
							// If this block is not the one we're waiting for, store it in the buffer
							buffer[block.blockNum] = block.data
						}
					}
				}
			}
		breakLoop:
		}()
	}
	if req.downloadStorer != nil {
		storerCtx, storerCancel := context.WithCancel(ctx)
		defer storerCancel()
		req.downloadStorer.Start(storerCtx)
	}

	var progressLock sync.Mutex
	firstReqWG := sync.WaitGroup{}
	firstReqWG.Add(1)
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(downloadWorkerCount + extraCount)
	for i := 0; i < n; i++ {
		j := i
		if i == 1 {
			firstReqWG.Wait()
			sort.Slice(req.downloadQueue, req.downloadQueue.Less)
		}
		select {
		case <-egCtx.Done():
			goto breakDownloadLoop
		default:
		}
		eg.Go(func() error {

			if j == 0 {
				defer firstReqWG.Done()
			}
			blocksToDownload := numBlocks
			if startBlock+int64(j)*numBlocks+numBlocks > endBlock {
				blocksToDownload = endBlock - (startBlock + int64(j)*numBlocks)
			}
			data, err := req.getBlocksData(startBlock+int64(j)*numBlocks, blocksToDownload, j == 0)
			if req.isDownloadCanceled {
				return errors.New("download_abort", "Download aborted by user")
			}
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Download failed for block %d. ", startBlock+int64(j)*numBlocks))
			}
			if !writerAt {
				blocks <- blockData{blockNum: j, data: data}
			} else {
				var offset int64
				if req.downloadStorer != nil {
					offset = (startBlock + int64(j)*numBlocks) * int64(req.effectiveBlockSize) * int64(req.datashards)
				} else {
					offset = int64(j) * numBlocks * int64(req.effectiveBlockSize) * int64(req.datashards)
				}
				var total int
				if j == n-1 {
					total, err = writeAtData(writeAtHandler, data, req.datashards, offset, int(remainingSize-offset))
				} else {
					total, err = writeAtData(writeAtHandler, data, req.datashards, offset, -1)
				}
				if err != nil {
					logger.Logger.Error("downloadFailed: ", startBlock+int64(j)*numBlocks, " remainingSize: ", remainingSize, " offset: ", offset)
					return errors.Wrap(err, fmt.Sprintf("WriteAt failed for block %d. ", startBlock+int64(j)*numBlocks))
				}
				for _, rb := range req.bufferMap {
					rb.ReleaseChunk(int(startBlock + int64(j)*numBlocks))
				}
				if req.downloadStorer != nil {
					go req.downloadStorer.Update(int(startBlock + int64(j)*numBlocks + blocksToDownload))
				}
				if req.statusCallback != nil {
					progressLock.Lock()
					downloaded += total
					req.statusCallback.InProgress(req.allocationID, remotePathCB, op, int(downloaded), nil)
					progressLock.Unlock()
				}
			}
			return nil
		})
	breakDownloadLoop:
	}
	if err := eg.Wait(); err != nil {
		writeCancel()
		close(blocks)
		wg.Wait()
		req.errorCB(err, remotePathCB)
		return
	}

	close(blocks)
	wg.Wait()
	// req.fileHandler.Sync() //nolint
	elapsedGetBlocksAndWrite := time.Since(now) - elapsedInitEC - elapsedInitEncryption
	l.Logger.Debug(fmt.Sprintf("[processDownload] Timings:\n allocation_id: %s,\n remotefilepath: %s,\n initEC: %d ms,\n initEncryption: %d ms,\n getBlocks and writes: %d ms",
		req.allocationID,
		req.remotefilepath,
		elapsedInitEC.Milliseconds(),
		elapsedInitEncryption.Milliseconds(),
		elapsedGetBlocksAndWrite.Milliseconds(),
	))

	if req.statusCallback != nil && !req.skip {
		req.statusCallback.Completed(
			req.allocationID, remotePathCB, fRef.Name, fRef.MimeType, int(size), op)
	}
	if req.downloadStorer != nil {
		req.downloadStorer.Remove() //nolint:errcheck
	}
}

func checkHash(actualFileHasher hash.Hash, fref *fileref.FileRef, contentMode string) (string, bool) {
	calculatedFileHash := hex.EncodeToString(actualFileHasher.Sum(nil))
	if contentMode == DOWNLOAD_CONTENT_THUMB {
		return calculatedFileHash, calculatedFileHash == fref.ActualThumbnailHash
	} else {
		return calculatedFileHash, calculatedFileHash == fref.ActualFileHash
	}
}

func (req *DownloadRequest) submitReadMarker(blobber *blockchain.StorageNode, readCount int64) (err error) {
	var retryCount = 3
	for retryCount > 0 {
		if err = req.attemptSubmitReadMarker(blobber, readCount); err != nil {
			logger.Logger.Error(fmt.Sprintf("Error while attempting to submit readmarker %v, retry: %d", err, retryCount))
			if IsErrCode(err, NotEnoughTokens) || IsErrCode(err, InvalidAuthTicket) || IsErrCode(err, InvalidShare) {
				return err
			}
			if IsErrCode(err, LockExists) || IsErrCode(err, RateLimitError) {
				continue
			}
			retryCount--
		} else {
			return nil
		}
	}
	blobber.SetSkip(true)
	return fmt.Errorf("submit read marker failed after retries: %w", err)
}

func (req *DownloadRequest) attemptSubmitReadMarker(blobber *blockchain.StorageNode, readCount int64) error {
	lockBlobberReadCtr(req.allocationID, blobber.ID)
	defer unlockBlobberReadCtr(req.allocationID, blobber.ID)
	rm := &marker.ReadMarker{
		ClientID:        client.GetClientID(),
		ClientPublicKey: client.GetClientPublicKey(),
		BlobberID:       blobber.ID,
		AllocationID:    req.allocationID,
		OwnerID:         req.allocOwnerID,
		Timestamp:       common.Now(),
		ReadCounter:     getBlobberReadCtr(req.allocationID, blobber.ID) + readCount,
		SessionRC:       readCount,
	}
	err := rm.Sign()
	if err != nil {
		return fmt.Errorf("error signing read marker: %w", err)
	}
	logger.Logger.Debug(fmt.Sprintf("Attempting to submit RM: ReadCounter: %d, SessionRC: %d, BlobberID: %v", rm.ReadCounter, rm.SessionRC, rm.BlobberID))
	rmData, err := json.Marshal(rm)
	if err != nil {
		return fmt.Errorf("error marshaling read marker: %w", err)
	}
	httpreq, err := zboxutil.NewRedeemRequest(blobber.Baseurl, req.allocationID, req.allocationTx)
	if err != nil {
		return fmt.Errorf("error creating download request: %w", err)
	}

	header := &DownloadRequestHeader{
		PathHash:     req.remotefilepathhash,
		ReadMarker:   rmData,
		ConnectionID: req.connectionID,
	}
	header.ToHeader(httpreq)

	ctx, cancel := context.WithTimeout(req.ctx, 30*time.Second)
	defer cancel()

	err = zboxutil.HttpDo(ctx, cancel, httpreq, func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		if resp.Body != nil {
			defer resp.Body.Close()
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			logger.Logger.Info(blobber.Baseurl,
				" got too many request error. Retrying")
			var r int
			r, err = zboxutil.GetRateLimitValue(resp)
			if err != nil {
				logger.Logger.Error(err)
				return errors.New("rate_limit_error", "Error while getting rate limit value")
			}
			time.Sleep(time.Duration(r) * time.Second)
			return errors.New("rate_limit_error", "Too many requests")
		}

		if resp.StatusCode != http.StatusOK {
			return req.handleReadMarkerError(resp, blobber, rm)
		}
		incBlobberReadCtr(req.allocationID, blobber.ID, readCount)

		logger.Logger.Debug("Submit readmarker 200 OK")

		return nil
	})
	return err
}

func (req *DownloadRequest) handleReadMarkerError(resp *http.Response, blobber *blockchain.StorageNode, rm *marker.ReadMarker) error {
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	appErrorCode := resp.Header.Get("X-App-Error-Code")
	if appErrorCode != "" {
		if appErrorCode == NotEnoughTokens {
			logger.Logger.Debug(fmt.Sprintf("NotEnoughTokens - blobberID: %v", blobber.ID))
			blobber.SetSkip(true)
			return errors.New(NotEnoughTokens, string(respBody))
		}
		if appErrorCode == InvalidAuthTicket {
			logger.Logger.Debug(fmt.Sprintf("InvalidAuthTicket - blobberID: %v", blobber.ID))
			blobber.SetSkip(true)
			return errors.New(InvalidAuthTicket, string(respBody))
		}
		if appErrorCode == InvalidShare {
			logger.Logger.Debug(fmt.Sprintf("InvalidShare - blobberID: %v", blobber.ID))
			blobber.SetSkip(true)
			return errors.New(InvalidShare, string(respBody))
		}
		if appErrorCode == LockExists {
			logger.Logger.Debug(fmt.Sprintf("LockExists - blobberID: %v", blobber.ID))
			time.Sleep(time.Second * 1)
			return errors.New(LockExists, string(respBody))
		}
	}

	var rspData downloadBlock
	if err = json.Unmarshal(respBody, &rspData); err == nil && rspData.LatestRM != nil {
		if err := rm.ValidateWithOtherRM(rspData.LatestRM); err != nil {
			return err
		}

		lastBlobberReadCounter := getBlobberReadCtr(req.allocationID, blobber.ID)
		if rspData.LatestRM.ReadCounter != lastBlobberReadCounter {
			setBlobberReadCtr(req.allocationID, blobber.ID, rspData.LatestRM.ReadCounter)
			return fmt.Errorf("stale_read_marker: readmarker counter is not in sync with latest counter. Last blobber read counter: %d, but readmarker's counter was: %d", rspData.LatestRM.ReadCounter, lastBlobberReadCounter)
		}
		return fmt.Errorf("download_error: response status: %d, error: %v", resp.StatusCode, rspData.err)
	}

	return fmt.Errorf("response_error: %s", string(respBody))
}

func IsErrCode(err error, code string) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*errors.Error); ok && e.Code == code {
		return true
	}
	return strings.Contains(err.Error(), code)
}

// initEC will initialize erasure encoder/decoder
func (req *DownloadRequest) initEC() error {
	var err error
	req.ecEncoder, err = reedsolomon.New(
		req.datashards, req.parityshards,
		reedsolomon.WithAutoGoroutines(int(req.effectiveBlockSize)))

	if err != nil {
		return errors.New("init_ec",
			fmt.Sprintf("Got error %s, while initializing erasure encoder", err.Error()))
	}
	return nil
}

// initEncryption will initialize encScheme with client's keys
func (req *DownloadRequest) initEncryption() (err error) {
	req.encScheme = encryption.NewEncryptionScheme()
	mnemonic := client.GetClient().Mnemonic
	if mnemonic != "" {
		_, err = req.encScheme.Initialize(client.GetClient().Mnemonic)
		if err != nil {
			return err
		}
	} else {
		return errors.New("invalid_mnemonic", "Invalid mnemonic")
	}

	err = req.encScheme.InitForDecryption("filetype:audio", req.encryptedKey)
	if err != nil {
		return err
	}
	return nil
}

func (req *DownloadRequest) errorCB(err error, remotePathCB string) {
	var op = OpDownload
	if req.contentMode == DOWNLOAD_CONTENT_THUMB {
		op = opThumbnailDownload
	}
	if req.downloadStorer != nil && !strings.Contains(err.Error(), "context canceled") {
		req.downloadStorer.Remove() //nolint: errcheck
	}
	if req.skip {
		return
	}
	req.skip = true
	if req.localFilePath != "" {
		if info, err := req.fileHandler.Stat(); err == nil && info.Size() == 0 {
			os.Remove(req.localFilePath) //nolint: errcheck
		}
	}
	if req.fileHandler != nil {
		req.fileHandler.Close() //nolint: errcheck
	}
	if req.statusCallback != nil {
		req.statusCallback.Error(
			req.allocationID, remotePathCB, op, err)
	}
}

func (req *DownloadRequest) calculateShardsParams(
	fRef *fileref.FileRef) (chunksPerShard int64, err error) {

	size := fRef.ActualFileSize
	if req.contentMode == DOWNLOAD_CONTENT_THUMB {
		if fRef.ActualThumbnailSize == 0 {
			return 0, errors.New("invalid_request", "Thumbnail does not exist")
		}
		size = fRef.ActualThumbnailSize
	}
	req.size = size
	req.encryptedKey = fRef.EncryptedKey
	req.chunkSize = int(fRef.ChunkSize)

	effectivePerShardSize := (size + int64(req.datashards) - 1) / int64(req.datashards)
	effectiveBlockSize := fRef.ChunkSize
	if fRef.EncryptedKey != "" {
		effectiveBlockSize -= EncryptionHeaderSize + EncryptedDataPaddingSize
	}

	req.effectiveBlockSize = int(effectiveBlockSize)

	chunksPerShard = (effectivePerShardSize + effectiveBlockSize - 1) / effectiveBlockSize

	info, err := req.fileHandler.Stat()
	if err != nil {
		return 0, err
	}
	// Can be nil when using file writer in wasm
	if info != nil {
		if req.downloadStorer != nil {
			err = sys.Files.MkdirAll(filepath.Join(req.workdir, "download"), 0766)
			if err != nil {
				return 0, err
			}
			progressID := req.progressID()
			var dp *DownloadProgress
			if info.Size() > 0 {
				dp = req.downloadStorer.Load(progressID, int(req.numBlocks))
			}
			if dp != nil {
				req.startBlock = int64(dp.LastWrittenBlock)
				if req.startBlock > 0 {
					req.isResume = true
				}
			} else {
				dp = &DownloadProgress{
					ID:        progressID,
					numBlocks: int(req.numBlocks),
				}
				req.downloadStorer.Save(dp)
			}
		}
	}

	if req.endBlock == 0 || req.endBlock > chunksPerShard {
		req.endBlock = chunksPerShard
	}

	if req.startBlock >= req.endBlock {
		err = errors.New("invalid_block_num", "start block should be less than end block")
		return 0, err
	}

	return
}

type blobberFile struct {
	validationRoot []byte
	size           int64
}

func GetFileRefFromBlobber(allocationID, blobberId, remotePath string) (fRef *fileref.FileRef, err error) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	blobber, err := GetBlobber(blobberId)
	if err != nil {
		return nil, err
	}

	a, err := GetAllocation(allocationID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	listReq := &ListRequest{}

	listReq.allocationID = a.ID
	listReq.allocationTx = a.Tx
	listReq.sig = a.sig
	listReq.blobbers = []*blockchain.StorageNode{
		{ID: string(blobber.ID), Baseurl: blobber.BaseURL},
	}
	listReq.fullconsensus = 1
	listReq.consensusThresh = 1
	listReq.ctx = ctx
	listReq.remotefilepath = remotePath

	rspCh := make(chan *fileMetaResponse, 1)
	go listReq.getFileMetaInfoFromBlobber(listReq.blobbers[0], 0, rspCh)
	resp := <-rspCh
	return resp.fileref, resp.err
}

func (req *DownloadRequest) getFileRef() (fRef *fileref.FileRef, err error) {
	listReq := &ListRequest{
		remotefilepath:     req.remotefilepath,
		remotefilepathhash: req.remotefilepathhash,
		allocationID:       req.allocationID,
		allocationTx:       req.allocationTx,
		sig:                req.sig,
		blobbers:           req.blobbers,
		authToken:          req.authTicket,
		Consensus: Consensus{
			RWMutex:         &sync.RWMutex{},
			fullconsensus:   req.fullconsensus,
			consensusThresh: req.consensusThresh,
		},
		ctx: req.ctx,
	}

	fMetaResp := listReq.getFileMetaFromBlobbers()

	fRef, err = req.getFileMetaConsensus(fMetaResp)
	if err != nil {
		return
	}

	if fRef.Type == fileref.DIRECTORY {
		err = errors.New("invalid_operation", "cannot download directory")
		return nil, err
	}
	return
}

// getFileMetaConsensus will verify actual file hash signature and take consensus in it.
// Then it will use the signature to calculation validation root signature and verify signature
// of validation root send by the blobber.
func (req *DownloadRequest) getFileMetaConsensus(fMetaResp []*fileMetaResponse) (*fileref.FileRef, error) {
	var selected *fileMetaResponse
	foundMask := zboxutil.NewUint128(0)
	req.consensus = 0
	retMap := make(map[string]int)
	for _, fmr := range fMetaResp {
		if fmr.err != nil || fmr.fileref == nil {
			continue
		}
		actualHash := fmr.fileref.ActualFileHash
		actualFileHashSignature := fmr.fileref.ActualFileHashSignature

		isValid, err := sys.VerifyWith(
			req.allocOwnerPubKey,
			actualFileHashSignature,
			actualHash,
		)
		if err != nil {
			l.Logger.Error(err)
			continue
		}
		if !isValid {
			l.Logger.Error("invalid signature")
			continue
		}

		retMap[actualFileHashSignature]++
		if retMap[actualFileHashSignature] > req.consensus {
			req.consensus = retMap[actualFileHashSignature]
		}
		if req.isConsensusOk() {
			selected = fmr
			break
		}
	}

	if selected == nil {
		l.Logger.Error("File consensus not found for ", req.remotefilepath)
		return nil, errors.New("consensus_not_met", "")
	}

	blobberCount := 0
	countThreshold := req.consensusThresh + 1
	if countThreshold > req.fullconsensus {
		countThreshold = req.consensusThresh
	}
	if req.freeRead {
		countThreshold = req.fullconsensus
	}
	for i := 0; i < len(fMetaResp); i++ {
		fmr := fMetaResp[i]
		if fmr.err != nil || fmr.fileref == nil {
			continue
		}
		fRef := fmr.fileref

		if selected.fileref.ActualFileHashSignature != fRef.ActualFileHashSignature {
			continue
		}

		shift := zboxutil.NewUint128(1).Lsh(uint64(fmr.blobberIdx))
		foundMask = foundMask.Or(shift)
		req.downloadQueue[fmr.blobberIdx] = downloadPriority{
			blobberIdx: fmr.blobberIdx,
			timeTaken:  60000,
		}
		blobberCount++
		if blobberCount == countThreshold {
			break
		}
	}
	req.consensus = foundMask.CountOnes()
	if !req.isConsensusOk() {
		return nil, fmt.Errorf("consensus_not_met")
	}
	req.downloadMask = foundMask
	sort.Slice(req.downloadQueue, req.downloadQueue.Less)
	return selected.fileref, nil
}

func (req *DownloadRequest) processDownloadRequest() {
	remotePathCB := req.remotefilepath
	if remotePathCB == "" {
		remotePathCB = req.remotefilepathhash
	}
	if req.startBlock < 0 || req.endBlock < 0 {
		req.errorCB(
			fmt.Errorf("start block or end block or both cannot be negative."), remotePathCB,
		)
		return
	}
	fRef, err := req.getFileRef()
	if err != nil {
		logger.Logger.Error(err.Error())
		req.errorCB(
			fmt.Errorf("Error while getting file ref. Error: %v",
				err), remotePathCB)

		return
	}
	req.fRef = fRef
	chunksPerShard, err := req.calculateShardsParams(fRef)
	if err != nil {
		logger.Logger.Error(err.Error())
		req.errorCB(
			fmt.Errorf("Error while calculating shard params. Error: %v",
				err), remotePathCB)
		return
	}
	req.chunksPerShard = chunksPerShard
	startBlock, endBlock := req.startBlock, req.endBlock
	// remainingSize should be calculated based on startBlock number
	// otherwise end data will have null bytes.
	remainingSize := req.size - startBlock*int64(req.effectiveBlockSize)

	var wantSize int64
	if endBlock*int64(req.effectiveBlockSize) < req.size {
		wantSize = endBlock*int64(req.effectiveBlockSize) - startBlock*int64(req.effectiveBlockSize)
	} else {
		wantSize = remainingSize
	}

	if remainingSize <= 0 {
		logger.Logger.Error("Nothing to download")
		req.errorCB(
			fmt.Errorf("Size to download is %d. Nothing to download", remainingSize), remotePathCB,
		)
		return
	}

	blocksPerShard := (wantSize + int64(req.effectiveBlockSize) - 1) / int64(req.effectiveBlockSize)
	req.blocksPerShard = blocksPerShard
}

func (req *DownloadRequest) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		if offset >= req.size {
			return 0, errors.New(ExceededMaxOffsetValue, "file is already downloaded")
		}
		req.offset = offset
	case io.SeekCurrent:
		if req.offset+offset >= req.size {
			return 0, errors.New(ExceededMaxOffsetValue, "")
		}
		req.offset += offset
	case io.SeekEnd:
		newOffset := req.size - offset
		if newOffset < 0 {
			return 0, errors.New(NegativeOffsetResultantValue, "")
		}
		req.offset = offset
	default:
		return 0, errors.New(InvalidWhenceValue,
			fmt.Sprintf("expected 0, 1 or 2, provided %d", whence))
	}
	return req.offset, nil
}

func writeData(dest io.Writer, data [][][]byte, dataShards, remaining int) (int, error) {
	total := 0
	for i := 0; i < len(data); i++ {
		for j := 0; j < dataShards; j++ {
			if len(data[i][j]) <= remaining {
				n, err := dest.Write(data[i][j])
				total += n
				if err != nil {
					return total, err
				}
			} else {
				n, err := dest.Write(data[i][j][:remaining])
				total += n
				if err != nil {
					return total, err
				}
			}
			remaining -= len(data[i][j])
			if remaining <= 0 {
				return total, nil
			}
		}
	}
	return total, nil
}

func writeAtData(dest io.WriterAt, data [][][]byte, dataShards int, offset int64, lastBlock int) (int, error) {
	var total int
	for i := 0; i < len(data); i++ {
		for j := 0; j < dataShards; j++ {
			if lastBlock != -1 {
				if len(data[i][j]) <= lastBlock {
					n, err := dest.WriteAt(data[i][j], offset+int64(total))
					total += n
					if err != nil {
						logger.Logger.Error("writeAt failed: ", err, " offset: ", offset, " total: ", total, "toWriteData: ", len(data[i][j]), " lastBlock: ", lastBlock)
						return total, err
					}
				} else {
					n, err := dest.WriteAt(data[i][j][:lastBlock], offset+int64(total))
					total += n
					if err != nil {
						logger.Logger.Error("writeAt failed: ", err, " offset: ", offset, " total: ", total, "toWriteData: ", len(data[i][j]), " lastBlock: ", lastBlock)
						return total, err
					}
				}
				lastBlock -= len(data[i][j])
				if lastBlock <= 0 {
					return total, nil
				}
			} else {
				n, err := dest.WriteAt(data[i][j], offset+int64(total))
				total += n
				if err != nil {
					logger.Logger.Error("writeAt failed: ", err, " offset: ", offset, " total: ", total)
					return total, err
				}
			}
		}
	}
	return total, nil
}

func (dr *DownloadRequest) progressID() string {

	if len(dr.allocationID) > 8 {
		return filepath.Join(dr.workdir, "download", "d"+dr.allocationID[:8]+"_"+dr.fRef.MetaID())
	}

	return filepath.Join(dr.workdir, "download", dr.allocationID+"_"+dr.fRef.MetaID())
}
