package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/0chain/errors"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encoder"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"go.dedis.ch/kyber/v3/group/edwards25519"
)

const (
	DOWNLOAD_CONTENT_FULL  = "full"
	DOWNLOAD_CONTENT_THUMB = "thumbnail"
)

var (
	FS common.FS = common.NewDiskFS()
)

type DownloadRequest struct {
	allocationID       string
	allocationTx       string
	blobbers           []*blockchain.StorageNode
	datashards         int
	parityshards       int
	remotefilepath     string
	remotefilepathhash string
	localpath          string
	startBlock         int64
	endBlock           int64
	chunkSize          int
	numBlocks          int64
	rxPay              bool
	statusCallback     StatusCallback
	ctx                context.Context
	ctxCncl            context.CancelFunc
	authTicket         *marker.AuthTicket
	wg                 *sync.WaitGroup
	downloadMask       zboxutil.Uint128
	encryptedKey       string
	isDownloadCanceled bool
	completedCallback  func(remotepath string, remotepathhash string)
	contentMode        string
	Consensus
}

func (req *DownloadRequest) downloadBlock(blockNum int64, blockChunksMax int) ([]byte, error) {
	req.consensus = 0
	numDownloads := req.downloadMask.CountOnes()
	req.wg = &sync.WaitGroup{}
	req.wg.Add(numDownloads)
	rspCh := make(chan *downloadBlock, numDownloads)
	// Download from only specific blobbers
	var c, pos int
	for i := req.downloadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(uint64(pos)).Not()) {
		pos = i.TrailingZeros()
		blockDownloadReq := &BlockDownloadRequest{}
		blockDownloadReq.allocationID = req.allocationID
		blockDownloadReq.allocationTx = req.allocationTx
		blockDownloadReq.authTicket = req.authTicket
		blockDownloadReq.blobber = req.blobbers[pos]
		blockDownloadReq.blobberIdx = pos
		blockDownloadReq.chunkSize = req.chunkSize
		blockDownloadReq.blockNum = blockNum
		blockDownloadReq.contentMode = req.contentMode
		blockDownloadReq.result = rspCh
		blockDownloadReq.wg = req.wg
		blockDownloadReq.ctx = req.ctx
		blockDownloadReq.remotefilepath = req.remotefilepath
		blockDownloadReq.remotefilepathhash = req.remotefilepathhash
		blockDownloadReq.numBlocks = req.numBlocks
		blockDownloadReq.rxPay = req.rxPay
		blockDownloadReq.encryptedKey = req.encryptedKey
		go AddBlockDownloadReq(blockDownloadReq)
		//go obj.downloadBlobberBlock(&obj.blobbers[pos], pos, path, blockNum, rspCh, isPathHash, authTicket)
		c++
	}
	//req.wg.Wait()
	shards := make([][][]byte, req.numBlocks)
	for i := int64(0); i < req.numBlocks; i++ {
		shards[i] = make([][]byte, len(req.blobbers))
	}
	//shards := make([][]byte, len(req.blobbers))
	decodeLen := make([]int, req.numBlocks)
	var decodeNumBlocks int

	retData := make([]byte, 0)
	success := 0
	logger.Logger.Info("downloadBlock ", blockNum, " numDownloads ", numDownloads)

	var encscheme encryption.EncryptionScheme
	if len(req.encryptedKey) > 0 {
		encscheme = encryption.NewEncryptionScheme()
		encscheme.Initialize(client.GetClient().Mnemonic)
		encscheme.InitForDecryption("filetype:audio", req.encryptedKey)
	}

	for i := 0; i < numDownloads; i++ {
		result := <-rspCh

		downloadChunks := len(result.BlockChunks)
		if !result.Success {
			logger.Logger.Error("Download block : ", req.blobbers[result.idx].Baseurl, " ", result.err)
		} else {
			blockSuccess := false
			if blockChunksMax < len(result.BlockChunks) {
				downloadChunks = blockChunksMax
			}

			for blockNum := 0; blockNum < downloadChunks; blockNum++ {
				if len(req.encryptedKey) > 0 {

					// dirty, but can't see other way right now
					if req.authTicket == nil {
						headerBytes := result.BlockChunks[blockNum][:(2 * 1024)]
						headerBytes = bytes.Trim(headerBytes, "\x00")
						headerString := string(headerBytes)

						encMsg := &encryption.EncryptedMessage{}
						encMsg.EncryptedData = result.BlockChunks[blockNum][(2 * 1024):]

						headerChecksums := strings.Split(headerString, ",")
						if len(headerChecksums) != 2 {
							logger.Logger.Error("Block has invalid header", req.blobbers[result.idx].Baseurl)
							continue
						}
						encMsg.MessageChecksum, encMsg.OverallChecksum = headerChecksums[0], headerChecksums[1]
						encMsg.EncryptedKey = encscheme.GetEncryptedKey()
						decryptedBytes, err := encscheme.Decrypt(encMsg)
						if err != nil {
							logger.Logger.Error("Block decryption failed", req.blobbers[result.idx].Baseurl, err)
							continue
						}
						shards[blockNum][result.idx] = decryptedBytes
					} else {
						suite := edwards25519.NewBlakeSHA256Ed25519()
						reEncMessage := &encryption.ReEncryptedMessage{
							D1: suite.Point(),
							D4: suite.Point(),
							D5: suite.Point(),
						}
						err := reEncMessage.Unmarshal(result.BlockChunks[blockNum])
						if err != nil {
							logger.Logger.Error("ReEncrypted Block unmarshall failed", req.blobbers[result.idx].Baseurl, err)
							break
						}
						decrypted, err := encscheme.ReDecrypt(reEncMessage)
						if err != nil {
							logger.Logger.Error("Block redecryption failed", req.blobbers[result.idx].Baseurl, err)
							break
						}
						shards[blockNum][result.idx] = decrypted
					}
				} else {
					shards[blockNum][result.idx] = result.BlockChunks[blockNum]
				}

				// All share should have equal length
				decodeLen[blockNum] = len(shards[blockNum][result.idx])
				blockSuccess = true
			}

			if !blockSuccess {
				continue
			}

			success++
			if success >= req.datashards {
				decodeNumBlocks = downloadChunks
				break
			}
		}
	}
	erasureencoder, err := encoder.NewEncoder(req.datashards, req.parityshards)
	if err != nil {
		return []byte{}, errors.Wrap(err, "encoder init error")
	}
	for blockNum := 0; blockNum < decodeNumBlocks; blockNum++ {
		data, err := erasureencoder.Decode(shards[blockNum], decodeLen[blockNum])
		if err != nil {
			return []byte{}, errors.Wrap(err, "Block decode error")
		}
		retData = append(retData, data...)
	}
	return retData, nil
}

func (req *DownloadRequest) processDownload(ctx context.Context) {
	defer req.ctxCncl()
	remotePathCallback := req.remotefilepath
	if len(req.remotefilepath) == 0 {
		remotePathCallback = req.remotefilepathhash
	}
	if req.completedCallback != nil {
		defer req.completedCallback(req.remotefilepath, req.remotefilepathhash)
	}

	// Only download from the Blobbers passes the consensus
	var fileRef *fileref.FileRef
	listReq := &ListRequest{
		remotefilepath:     req.remotefilepath,
		remotefilepathhash: req.remotefilepathhash,
		allocationID:       req.allocationID,
		allocationTx:       req.allocationTx,
		blobbers:           req.blobbers,
		ctx:                req.ctx,
	}
	listReq.authToken = req.authTicket
	listReq.fullconsensus = req.fullconsensus
	listReq.consensusThresh = req.consensusThresh
	req.downloadMask, fileRef, _ = listReq.getFileConsensusFromBlobbers()
	if req.downloadMask.Equals64(0) || fileRef == nil {
		if req.statusCallback != nil {
			req.statusCallback.Error(req.allocationID, remotePathCallback, OpDownload, errors.New("", "No minimum consensus for file meta data of file"))
		}
		return
	}

	// the ChunkSize value can't be less than 0kb
	if fileRef.Type == fileref.FILE && fileRef.ChunkSize <= 0 {
		if req.statusCallback != nil {
			req.statusCallback.Error(req.allocationID, remotePathCallback, OpDownload, errors.New("", "File ChunkSize value is not permitted"))
		}
		return
	}

	if fileRef.Type == fileref.DIRECTORY {
		if req.statusCallback != nil {
			req.statusCallback.Error(req.allocationID, remotePathCallback, OpDownload, errors.New("", "please get files from folder, and download them one by one"))
		}
		return
	}

	size := fileRef.ActualFileSize
	if req.contentMode == DOWNLOAD_CONTENT_THUMB {
		size = fileRef.ActualThumbnailSize
	}
	req.encryptedKey = fileRef.EncryptedKey
	req.chunkSize = int(fileRef.ChunkSize)
	logger.Logger.Info("Encrypted key from fileref", req.encryptedKey)
	// Calculate number of bytes per shard.
	perShard := (size + int64(req.datashards) - 1) / int64(req.datashards)
	chunkSizeWithHeader := int64(fileRef.ChunkSize)
	if len(fileRef.EncryptedKey) > 0 {
		chunkSizeWithHeader -= 16
		chunkSizeWithHeader -= 2 * 1024
	}
	chunksPerShard := (perShard + chunkSizeWithHeader - 1) / chunkSizeWithHeader
	if len(fileRef.EncryptedKey) > 0 {
		perShard += chunksPerShard * (16 + (2 * 1024))
	}

	wrFile, err := FS.OpenFile(req.localpath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		if req.statusCallback != nil {
			logger.Logger.Error(err.Error())
			req.statusCallback.Error(req.allocationID, remotePathCallback, OpDownload, errors.Wrap(err, "Can't create local file"))
		}
		return
	}
	defer wrFile.Close()
	req.isDownloadCanceled = false
	if req.statusCallback != nil {
		req.statusCallback.Started(req.allocationID, remotePathCallback, OpDownload, int(size))
	}

	if req.endBlock == 0 {
		req.endBlock = chunksPerShard
	}

	logger.Logger.Info("Download Size:", size, " Shard:", perShard, " chunks/shard:", chunksPerShard)
	logger.Logger.Info("Start block: ", req.startBlock+1, " End block: ", req.endBlock, " Num blocks: ", req.numBlocks)

	downloaded := int(0)
	fileHasher := createDownloadHasher(req.chunkSize, req.datashards, len(fileRef.EncryptedKey) > 0)
	mW := io.MultiWriter(fileHasher, wrFile)

	startBlock := req.startBlock
	endBlock := req.endBlock
	numBlocks := req.numBlocks

	for startBlock < endBlock {
		cnt := startBlock
		logger.Logger.Info("Downloading block ", cnt+1)
		if (startBlock + numBlocks) > endBlock {
			numBlocks = endBlock - startBlock
		}

		data, err := req.downloadBlock(cnt+1, int(numBlocks))
		if err != nil {
			FS.Remove(req.localpath)
			if req.statusCallback != nil {
				req.statusCallback.Error(req.allocationID, remotePathCallback, OpDownload, errors.Wrap(err, fmt.Sprintf("Download failed for block %d. ", cnt+1)))
			}
			return
		}
		if req.isDownloadCanceled {
			req.isDownloadCanceled = false
			FS.Remove(req.localpath)
			if req.statusCallback != nil {
				req.statusCallback.Error(req.allocationID, remotePathCallback, OpDownload, errors.New("", "Download aborted by user"))
			}
			return
		}

		n := int64(math.Min(float64(size), float64(len(data))))
		_, err = mW.Write(data[:n])

		if err != nil {
			FS.Remove(req.localpath)
			if req.statusCallback != nil {
				req.statusCallback.Error(req.allocationID, remotePathCallback, OpDownload, errors.Wrap(err, "Write file failed"))
			}
			return
		}
		downloaded = downloaded + int(n)
		size = size - n

		if req.statusCallback != nil {
			req.statusCallback.InProgress(req.allocationID, remotePathCallback, OpDownload, downloaded, data)
		}

		if (startBlock + numBlocks) > endBlock {
			startBlock += endBlock - startBlock
		} else {
			startBlock += numBlocks
		}
	}

	// Only check hash when the download request is not by block/partial.
	if req.endBlock == chunksPerShard && req.startBlock == 0 {
		//calcHash := fileHasher.GetHash()
		merkleRoot := fileHasher.GetMerkleRoot()

		expectedHash := fileRef.ActualFileHash
		if req.contentMode == DOWNLOAD_CONTENT_THUMB {
			expectedHash = fileRef.ActualThumbnailHash
		}

		//if calcHash != expectedHash && expectedHash != merkleRoot {
		if expectedHash != merkleRoot {
			FS.Remove(req.localpath)
			if req.statusCallback != nil {
				req.statusCallback.Error(req.allocationID, remotePathCallback, OpDownload, errors.New("", "File content didn't match with uploaded file"))
			}
			return
		}
	}

	wrFile.Sync()
	wrFile.Close()
	wrFile, _ = FS.Open(req.localpath)
	defer wrFile.Close()
	wrFile.Seek(0, 0)
	mimetype, _ := zboxutil.GetFileContentType(wrFile)
	if req.statusCallback != nil {
		req.statusCallback.Completed(req.allocationID, remotePathCallback, fileRef.Name, mimetype, int(fileRef.ActualFileSize), OpDownload)
	}
}
