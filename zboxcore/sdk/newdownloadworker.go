package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/klauspost/reedsolomon"
	"go.dedis.ch/kyber/v3/group/edwards25519"
)

type NewDownloadRequest struct {
	DownloadRequest
	effectiveChunkSize int
	ecEncoder          reedsolomon.Encoder
	maskMu             *sync.Mutex
	encScheme          encryption.EncryptionScheme
}

func (req *NewDownloadRequest) removeFromMask(pos uint64) {
	req.maskMu.Lock()
	req.downloadMask = req.downloadMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
	req.maskMu.Unlock()
}

// comment.
// use context everywhere
func (req *NewDownloadRequest) getBlocksData(
	startBlock, totalBlock int64,
	mask zboxutil.Uint128, requiredDownloads int) ([]byte, error) {

	shards := make([][][]byte, totalBlock)
	for i := range shards {
		shards[i] = make([][]byte, len(req.blobbers))
	}

	var (
		remainingMask zboxutil.Uint128
		failed        int
		err           error
	)

	curReqDownloads := requiredDownloads
	for {
		remainingMask, failed, err = req.downloadBlock(startBlock, totalBlock, mask, curReqDownloads, shards)
		if err != nil {
			return nil, err
		}
		if failed == 0 {
			break
		}

		if failed > remainingMask.CountOnes() {
			return nil, errors.New("download_failed", "")
		}

		curReqDownloads = failed
		mask = remainingMask
	}

	// erasure decoding
	// Can we benefit from goroutine for erasure decoding??
	c := req.datashards * req.effectiveChunkSize
	data := make([]byte, req.datashards*req.effectiveChunkSize*int(totalBlock))
	var isValid bool
	for i := range shards {
		var d []byte
		var err error
		d, isValid, err = req.decodeEC(shards[i])
		if err != nil {
			return nil, err
		}

		if !isValid {
			return nil, errors.New("invalid_data", "some blobber responded with wrong data")
		}
		index := i * c
		copy(data[index:index+c], d)

	}
	return data, nil

	// if isValid {
	// }

	// data = nil
	// requiredDownloads := remainingMask.CountOnes()
	// _, failed, err = req.downloadBlock(startBlock, totalBlock, remainingMask, requiredDownloads, shards)
	// if err != nil {
	// 	return nil, err
	// }

	// if failed == requiredDownloads {
	// 	return nil, errors.New("download_failed",
	// 		"Downloading blocks from remaining blobbers all failed")
	// }

	// data = make([]byte, req.datashards*req.effectiveChunkSize, totalBlock)
	// for i := range shards {
	// 	d, isValid, err := req.shuffleAndReconstruct(shards[i])
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	if !isValid {
	// 		return nil, errors.New("too_many_invalid_data",
	// 			"Too many blobbers returned wrong data than can be handled")
	// 	}

	// 	index := i * c
	// 	copy(data[index:index+c], d)
	// }

	// return data, nil

}

// comment.
func (req *NewDownloadRequest) downloadBlock(
	startBlock, totalBlock int64,
	mask zboxutil.Uint128, requiredDownloads int,
	shards [][][]byte) (zboxutil.Uint128, int, error) {

	var remainingMask zboxutil.Uint128
	activeBlobbers := mask.CountOnes()
	if activeBlobbers < requiredDownloads {
		return zboxutil.NewUint128(0), 0, errors.New("insufficient_blobbers",
			fmt.Sprintf("Required downloads %d, remaining active blobber %d",
				req.consensusThresh, activeBlobbers))
	}
	rspCh := make(chan *downloadBlock, req.consensusThresh)

	var pos uint64
	var c int

	for i := req.downloadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		blockDownloadReq := &BlockDownloadRequest{
			allocationID:       req.allocationID,
			allocationTx:       req.allocationTx,
			allocOwnerID:       req.allocOwnerID,
			authTicket:         req.authTicket,
			blobber:            req.blobbers[pos],
			blobberIdx:         int(pos),
			chunkSize:          req.chunkSize,
			blockNum:           startBlock,
			contentMode:        req.contentMode,
			result:             rspCh,
			ctx:                req.ctx,
			remotefilepath:     req.remotefilepath,
			remotefilepathhash: req.remotefilepathhash,
			numBlocks:          totalBlock,
			encryptedKey:       req.encryptedKey,
		}

		go AddBlockDownloadReq(blockDownloadReq)
		c++
		if c == requiredDownloads {
			remainingMask = i
			break
		}

	}

	var failed int
	for i := 0; i < requiredDownloads; i++ {
		result := <-rspCh
		err := req.fillShards(shards, result)
		if err != nil {
			logger.Logger.Error(err)
			failed++
		}
	}

	return remainingMask, failed, nil
}

func (req *NewDownloadRequest) decodeEC(shards [][]byte) (data []byte, isValid bool, err error) {
	err = req.ecEncoder.Reconstruct(shards)
	if err != nil {
		return
	}

	isValid, err = req.ecEncoder.Verify(shards)
	if err != nil || !isValid {
		return
	}

	c := len(shards[0])
	data = make([]byte, req.datashards*c)
	for i := 0; i < req.datashards; i++ {
		index := i * c
		copy(data[index:index+c], shards[i])
	}
	return data, true, nil
}

func (req *NewDownloadRequest) fillShards(shards [][][]byte, result *downloadBlock) (err error) {
	defer func() {
		if err != nil {
			req.removeFromMask(uint64(result.idx))
		}
	}()

	if !result.Success {
		return result.err
	}

	for i := 0; i < len(result.BlockChunks); i++ {
		var data []byte
		if req.encryptedKey != "" {
			data, err = req.getDecryptedData(result, i)
			if err != nil {
				shards[i] = nil
				return err
			}
		} else {
			data = result.BlockChunks[i]
		}
		shards[i][result.idx] = data
	}
	return
}

func (req *NewDownloadRequest) getDecryptedData(result *downloadBlock, blockNum int) (data []byte, err error) {
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

func (req *NewDownloadRequest) getDecryptedDataForAuthTicket(result *downloadBlock, blockNum int) (data []byte, err error) {
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

func (req *NewDownloadRequest) processDownload(ctx context.Context) {
	if req.completedCallback != nil {
		defer req.completedCallback(req.remotefilepath, req.remotefilepathhash)
	}

	remotePathCB := req.remotefilepath
	if remotePathCB == "" {
		remotePathCB = req.remotefilepathhash
	}
	fRef, err := req.getFileRef(remotePathCB)
	if err != nil {
		logger.Logger.Error(err.Error())
		return
	}

	size, chunksPerShard, actualPerShard, err := req.calculateShardsParams(fRef, remotePathCB)
	if err != nil {
		logger.Logger.Error(err.Error())
		return
	}

	logger.Logger.Info(
		fmt.Sprintf("Downloading file with size: %d from start block: %d and end block: %d. "+
			"Actual size per blobber: %d", size, req.startBlock, req.endBlock, actualPerShard),
	)

	f, err := req.getFileHandler(remotePathCB)
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	defer f.Close()

	var isFullDownload bool
	fileHasher := createDownloadHasher(req.chunkSize, req.datashards, fRef.EncryptedKey != "")
	var mW io.Writer
	if req.startBlock == 0 && req.endBlock == chunksPerShard {
		isFullDownload = true
		mW = io.MultiWriter(fileHasher, f)
	} else {
		mW = io.MultiWriter(f)
	}

	err = req.initEC()
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	if req.encryptedKey != "" {
		req.initEncryption()
	}

	var downloaded int
	startBlock, endBlock, numBlocks := req.startBlock, req.endBlock, req.numBlocks
	// remainingSize should be calculated based on startBlock number
	// otherwise end data will have null bytes.
	remainingSize := size - startBlock*int64(req.effectiveChunkSize)

	if req.statusCallback != nil {
		req.statusCallback.Started(req.allocationID, remotePathCB, OpDownload, int(size))
	}

	for startBlock < endBlock {
		if startBlock+numBlocks > endBlock {
			numBlocks = endBlock - startBlock
		}
		logger.Logger.Info("Downloading block ", startBlock+1, " - ", startBlock+numBlocks)

		data, err := req.getBlocksData(startBlock+1, numBlocks, req.downloadMask, req.consensusThresh)
		if err != nil {
			req.errorCB(errors.Wrap(err, fmt.Sprintf("Download failed for block %d. ", startBlock+1)), remotePathCB)
			return
		}
		if req.isDownloadCanceled {
			req.errorCB(errors.New("download_abort", "Download aborted by user"), remotePathCB)
			return
		}

		n := int64(math.Min(float64(remainingSize), float64(len(data))))
		_, err = mW.Write(data[:n])

		if err != nil {
			req.errorCB(errors.Wrap(err, "Write file failed"), remotePathCB)
			return
		}
		downloaded = downloaded + int(n)
		remainingSize -= n

		if req.statusCallback != nil {
			req.statusCallback.InProgress(req.allocationID, remotePathCB, OpDownload, downloaded, data)
		}
		if (startBlock + numBlocks) > endBlock {
			startBlock += endBlock - startBlock
		} else {
			startBlock += numBlocks
		}
	}

	if isFullDownload {
		err := req.checkContentHash(fRef, fileHasher, remotePathCB)
		if err != nil {
			logger.Logger.Error(err)
			return
		}
	}

	f.Sync()
	if req.statusCallback != nil {
		req.statusCallback.Completed(
			req.allocationID, remotePathCB, fRef.Name, "", int(fRef.ActualFileSize), OpDownload)
	}
}

func (req *NewDownloadRequest) initEC() error {
	var err error
	req.ecEncoder, err = reedsolomon.New(
		req.datashards, req.parityshards,
		reedsolomon.WithAutoGoroutines(int(req.effectiveChunkSize)))

	if err != nil {
		return errors.New("init_ec",
			fmt.Sprintf("Got error %s, while initializing erasure encoder", err.Error()))
	}
	return nil
}

func (req *NewDownloadRequest) initEncryption() {
	req.encScheme = encryption.NewEncryptionScheme()
	req.encScheme.Initialize(client.GetClient().Mnemonic)
	req.encScheme.InitForDecryption("filetype:audio", req.encryptedKey)
}

func (req *NewDownloadRequest) errorCB(err error, remotePathCB string) {
	sys.Files.Remove(req.localpath) //nolint: errcheck
	if req.statusCallback != nil {
		req.statusCallback.Error(
			req.allocationID, remotePathCB, OpDownload, err)
	}
	return
}

func (req *NewDownloadRequest) checkContentHash(
	fRef *fileref.FileRef, fileHasher *downloadHasher, remotepathCB string) (err error) {

	hash := fileHasher.GetHash()
	expectedHash := fRef.ActualFileHash
	if req.contentMode == DOWNLOAD_CONTENT_THUMB {
		expectedHash = fRef.ActualThumbnailHash
	}

	if expectedHash != hash {
		err = errors.New("merkle_root_mismatch", "File content didn't match with uploaded file")
		req.errorCB(err, remotepathCB)
		return
	}
	return nil
}

func (req *NewDownloadRequest) getFileHandler(remotePathCB string) (sys.File, error) {
	f, err := sys.Files.OpenFile(req.localpath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		req.errorCB(errors.Wrap(err, "Can't create local file"), remotePathCB)
		return nil, err
	}
	return f, nil
}

func (req *NewDownloadRequest) calculateShardsParams(
	fRef *fileref.FileRef, remotePathCB string) (
	size, chunksPerShard, actualPerShard int64, err error) {

	size = fRef.ActualFileSize
	if req.contentMode == DOWNLOAD_CONTENT_THUMB {
		size = fRef.ActualThumbnailSize
	}
	req.encryptedKey = fRef.EncryptedKey
	req.chunkSize = int(fRef.ChunkSize)

	// fRef.ActualFileSize is size of file that does not include encryption bytes.
	// that is why, actualPerShard will have different value for encrypted file.
	effectivePerShardSize := (size + int64(req.datashards) - 1) / int64(req.datashards)
	effectiveChunkSize := fRef.ChunkSize
	if fRef.EncryptedKey != "" {
		effectiveChunkSize -= EncryptionHeaderSize + EncryptedDataPaddingSize
	}

	// TODO re-check out this assignment
	req.effectiveChunkSize = int(effectiveChunkSize)

	chunksPerShard = (effectivePerShardSize + effectiveChunkSize - 1) / effectiveChunkSize
	actualPerShard = chunksPerShard * fRef.ChunkSize
	if req.endBlock == 0 {
		req.endBlock = chunksPerShard
	}

	if req.startBlock >= req.endBlock {
		err = errors.New("invalid_block_num", "start block should be less than end block")
		req.errorCB(err, remotePathCB)
		return 0, 0, 0, err
	}

	return
}

func (req *NewDownloadRequest) getFileRef(remotePathCB string) (fRef *fileref.FileRef, err error) {
	listReq := &ListRequest{
		remotefilepath:     req.remotefilepath,
		remotefilepathhash: req.remotefilepathhash,
		allocationID:       req.allocationID,
		allocationTx:       req.allocationTx,
		blobbers:           req.blobbers,
		authToken:          req.authTicket,
		Consensus: Consensus{
			mu:              &sync.RWMutex{},
			fullconsensus:   req.fullconsensus,
			consensusThresh: req.consensusThresh,
		},
		ctx: req.ctx,
	}

	req.downloadMask, fRef, _ = listReq.getFileConsensusFromBlobbers()
	if req.downloadMask.Equals64(0) || fRef == nil {
		err = errors.New("consensus_not_met", "No minimum consensus for file meta data of file")
		req.errorCB(err, remotePathCB)
		return
	}

	if fRef.Type == fileref.DIRECTORY {
		err = errors.New("invalid_operation", "cannot downoad directory")
		req.errorCB(err, remotePathCB)
		return
	}

	// the ChunkSize value can't be less than 0kb
	if fRef.ChunkSize <= 0 {
		err = errors.New("invalid_chunk_size", "File ChunkSize value is not permitted")
		req.errorCB(err, remotePathCB)
		return
	}

	return
}
