package sdk

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"hash"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/0chain/errors"
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
	"golang.org/x/crypto/sha3"
	"golang.org/x/sync/errgroup"
)

const (
	DOWNLOAD_CONTENT_FULL  = "full"
	DOWNLOAD_CONTENT_THUMB = "thumbnail"
)

type DownloadRequest struct {
	allocationID       string
	allocationTx       string
	allocOwnerID       string
	allocOwnerPubKey   string
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
	validationRootMap  map[string]*blobberFile
	statusCallback     StatusCallback
	ctx                context.Context
	ctxCncl            context.CancelFunc
	authTicket         *marker.AuthTicket
	downloadMask       zboxutil.Uint128
	encryptedKey       string
	isDownloadCanceled bool
	completedCallback  func(remotepath string, remotepathhash string)
	contentMode        string
	Consensus
	effectiveChunkSize int
	ecEncoder          reedsolomon.Encoder
	maskMu             *sync.Mutex
	encScheme          encryption.EncryptionScheme
	shouldVerify       bool
}

func (req *DownloadRequest) removeFromMask(pos uint64) {
	req.maskMu.Lock()
	req.downloadMask = req.downloadMask.And(zboxutil.NewUint128(1).Lsh(pos).Not())
	req.maskMu.Unlock()
}

// getBlocksData will get data blocks for some interval from minimal blobers and aggregate them and
// return to the caller
func (req *DownloadRequest) getBlocksData(startBlock, totalBlock int64) ([]byte, error) {

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
			startBlock, totalBlock, mask, curReqDownloads, shards)
		if err != nil {
			return nil, err
		}
		if failed == 0 {
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
}

// downloadBlock This function will add download requests to the download channel which picks up
// download requests and processes it.
// This function will fill up `shards` in respective position and also return failed number of
// blobbers along with remainingMask that are the blobbers that are not yet requested.
func (req *DownloadRequest) downloadBlock(
	startBlock, totalBlock int64,
	mask zboxutil.Uint128, requiredDownloads int,
	shards [][][]byte) (zboxutil.Uint128, int, []string, error) {

	var remainingMask zboxutil.Uint128
	activeBlobbers := mask.CountOnes()
	if activeBlobbers < requiredDownloads {
		return zboxutil.NewUint128(0), 0, nil, errors.New("insufficient_blobbers",
			fmt.Sprintf("Required downloads %d, remaining active blobber %d",
				req.consensusThresh, activeBlobbers))
	}
	rspCh := make(chan *downloadBlock, requiredDownloads)

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
			shouldVerify:       req.shouldVerify,
		}

		bf := req.validationRootMap[blockDownloadReq.blobber.ID]
		blockDownloadReq.blobberFile = bf

		go AddBlockDownloadReq(blockDownloadReq)
		c++
		if c == requiredDownloads {
			remainingMask = i
			break
		}

	}

	var failed int32
	downloadErrors := make([]string, requiredDownloads)
	wg := &sync.WaitGroup{}
	for i := 0; i < requiredDownloads; i++ {
		result := <-rspCh
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var err error
			defer func() {
				if err != nil {
					atomic.AddInt32(&failed, 1)
					req.removeFromMask(uint64(result.idx))
					downloadErrors[i] = fmt.Sprintf("Error %s from %s",
						err.Error(), req.blobbers[result.idx].Baseurl)
					logger.Logger.Error(err)
				}
			}()
			if !result.Success {
				err = fmt.Errorf("Unsuccessful download. Error: %v", result.err)
				return
			}
			err = req.fillShards(shards, result)
			return
		}(i)
	}

	wg.Wait()
	return remainingMask, int(failed), downloadErrors, nil
}

// decodeEC will reconstruct shards and verify it
func (req *DownloadRequest) decodeEC(shards [][]byte) (data []byte, isValid bool, err error) {
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

// fillShards will fill `shards` with data from blobbers that belongs to specific
// blockNumber and blobber's position index in an allocation
func (req *DownloadRequest) fillShards(shards [][][]byte, result *downloadBlock) (err error) {
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
// This will also write data to the file and will verify content by calculating content hash.
func (req *DownloadRequest) processDownload(ctx context.Context) {
	if req.completedCallback != nil {
		defer req.completedCallback(req.remotefilepath, req.remotefilepathhash)
	}

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

	var op = OpDownload
	if req.contentMode == DOWNLOAD_CONTENT_THUMB {
		op = opThumbnailDownload
	}
	fRef, err := req.getFileRef(remotePathCB)
	if err != nil {
		logger.Logger.Error(err.Error())
		req.errorCB(
			fmt.Errorf("Error while getting file ref. Error: %v",
				err), remotePathCB)

		return
	}

	size, chunkPerShard, actualPerShard, err := req.calculateShardsParams(fRef, remotePathCB)
	if err != nil {
		logger.Logger.Error(err.Error())
		req.errorCB(
			fmt.Errorf("Error while calculating shard params. Error: %v",
				err), remotePathCB)
		return
	}

	logger.Logger.Info(
		fmt.Sprintf("Downloading file with size: %d from start block: %d and end block: %d. "+
			"Actual size per blobber: %d", size, req.startBlock, req.endBlock, actualPerShard),
	)

	f, err := req.openFile()
	if err != nil {
		logger.Logger.Error(err)
		req.errorCB(
			fmt.Errorf("Error while getting file handler. Error: %v",
				err), remotePathCB)
		return
	}
	defer f.Close()

	err = req.initEC()
	if err != nil {
		logger.Logger.Error(err)
		req.errorCB(
			fmt.Errorf("Error while initializing file ref. Error: %v",
				err), remotePathCB)
		return
	}
	if req.encryptedKey != "" {
		err = req.initEncryption()
		if err != nil {
			req.errorCB(
				fmt.Errorf("Error while initializing encryption"), remotePathCB,
			)
			return
		}
	}

	var downloaded int
	startBlock, endBlock, numBlocks := req.startBlock, req.endBlock, req.numBlocks
	// remainingSize should be calculated based on startBlock number
	// otherwise end data will have null bytes.
	remainingSize := size - startBlock*int64(req.effectiveChunkSize)
	if remainingSize <= 0 {
		logger.Logger.Error("Nothing to download")
		req.errorCB(
			fmt.Errorf("Size to download is %d. Nothing to download", remainingSize), remotePathCB,
		)
		return
	}

	if req.statusCallback != nil {
		// Started will also initialize progress bar. So without calling this function
		// other callback's call will panic

		req.statusCallback.Started(req.allocationID, remotePathCB, op, int(size))
	}

	if req.shouldVerify {
		if req.authTicket != nil && req.encryptedKey != "" {
			req.shouldVerify = false
		}
	}
	var actualFileHasher hash.Hash
	var isPREAndWholeFile bool
	if !req.shouldVerify && (startBlock == 0 && endBlock == chunkPerShard) {
		actualFileHasher = sha3.New256()
		isPREAndWholeFile = true
	}
	n := int((endBlock - startBlock + numBlocks - 1) / numBlocks)
	res := make([][]byte, n)

	eg, _ := errgroup.WithContext(ctx)

	for i := 0; i < n; i++ {
		j := i
		eg.Go(func() error {
			blocksToDownload := numBlocks
			if startBlock+int64(j)*numBlocks+numBlocks > endBlock {
				blocksToDownload = endBlock - (startBlock + int64(j)*numBlocks)
			}
			res[j], err = req.getBlocksData(startBlock+int64(j)*numBlocks, blocksToDownload)
			if req.isDownloadCanceled {
				return errors.New("download_abort", "Download aborted by user")
			}
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Download failed for block %d. ", startBlock+1))
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		req.errorCB(err, remotePathCB)
		return
	}

	for _, data := range res {
		n := int64(math.Min(float64(remainingSize), float64(len(data))))
		_, err = f.Write(data[:n])

		if err != nil {
			req.errorCB(errors.Wrap(err, "Write file failed"), remotePathCB)
			return
		}
		if isPREAndWholeFile {
			actualFileHasher.Write(data[:n])
		}

		downloaded = downloaded + int(n)
		remainingSize -= n

		if req.statusCallback != nil {
			req.statusCallback.InProgress(req.allocationID, remotePathCB, op, downloaded, data)
		}
	}

	f.Sync()

	if isPREAndWholeFile {
		calculatedFileHash := hex.EncodeToString(actualFileHasher.Sum(nil))
		var actualHash string
		if req.contentMode == DOWNLOAD_CONTENT_THUMB {
			actualHash = fRef.ActualThumbnailHash
		} else {
			actualHash = fRef.ActualFileHash
		}

		if calculatedFileHash != actualHash {
			req.errorCB(fmt.Errorf("Expected actual file hash %s, calculated file hash %s",
				fRef.ActualFileHash, calculatedFileHash), remotePathCB)
			return
		}
	}

	if req.statusCallback != nil {
		req.statusCallback.Completed(
			req.allocationID, remotePathCB, fRef.Name, "", int(fRef.ActualFileSize), op)
	}
}

func (req *DownloadRequest) initEC() error {
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

func (req *DownloadRequest) initEncryption() error {
	req.encScheme = encryption.NewEncryptionScheme()
	mnemonic := client.GetClient().Mnemonic
	if mnemonic != "" {
		_, err := req.encScheme.Initialize(client.GetClient().Mnemonic)
		if err != nil {
			return err
		}
	} else {
		key, err := hex.DecodeString(client.GetClientPrivateKey())
		if err != nil {
			return err
		}
		req.encScheme.InitializeWithPrivateKey(key)
	}

	req.encScheme.InitForDecryption("filetype:audio", req.encryptedKey)
	return nil
}

func (req *DownloadRequest) errorCB(err error, remotePathCB string) {
	var op = OpDownload
	if req.contentMode == DOWNLOAD_CONTENT_THUMB {
		op = opThumbnailDownload
	}
	sys.Files.Remove(req.localpath) //nolint: errcheck
	if req.statusCallback != nil {
		req.statusCallback.Error(
			req.allocationID, remotePathCB, op, err)
	}
	return
}

func (req *DownloadRequest) openFile() (sys.File, error) {
	f, err := sys.Files.OpenFile(req.localpath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "Can't create local file")
	}
	return f, nil
}

func (req *DownloadRequest) calculateShardsParams(
	fRef *fileref.FileRef, remotePathCB string) (
	size, chunksPerShard, actualPerShard int64, err error) {

	size = fRef.ActualFileSize
	if req.contentMode == DOWNLOAD_CONTENT_THUMB {
		if fRef.ActualThumbnailSize == 0 {
			return 0, 0, 0, errors.New("invalid_request", "Thumbnail does not exist")
		}
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

	req.effectiveChunkSize = int(effectiveChunkSize)

	chunksPerShard = (effectivePerShardSize + effectiveChunkSize - 1) / effectiveChunkSize
	actualPerShard = chunksPerShard * fRef.ChunkSize
	if req.endBlock == 0 || req.endBlock > chunksPerShard {
		req.endBlock = chunksPerShard
	}

	if req.startBlock >= req.endBlock {
		err = errors.New("invalid_block_num", "start block should be less than end block")
		return 0, 0, 0, err
	}

	return
}

type blobberFile struct {
	validationRoot []byte
	size           int64
}

func (req *DownloadRequest) getFileRef(remotePathCB string) (fRef *fileref.FileRef, err error) {
	listReq := &ListRequest{
		remotefilepath:     req.remotefilepath,
		remotefilepathhash: req.remotefilepathhash,
		allocationID:       req.allocationID,
		allocationTx:       req.allocationTx,
		blobbers:           req.blobbers,
		authToken:          req.authTicket,
		Consensus: Consensus{
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

	req.validationRootMap = make(map[string]*blobberFile)
	for i := 0; i < len(fMetaResp); i++ {
		fmr := fMetaResp[i]
		if fmr.err != nil || fmr.fileref == nil {
			continue
		}
		fRef := fmr.fileref

		if selected.fileref.ActualFileHashSignature != fRef.ActualFileHashSignature {
			continue
		}

		isValid, err := sys.VerifyWith(
			req.allocOwnerPubKey,
			fRef.ValidationRootSignature,
			fRef.ActualFileHashSignature+fRef.ValidationRoot,
		)
		if err != nil {
			l.Logger.Error(err)
			continue
		}
		if !isValid {
			l.Logger.Error("invalid validation root signature")
			continue
		}

		blobber := req.blobbers[fmr.blobberIdx]
		vr, _ := hex.DecodeString(fmr.fileref.ValidationRoot)
		req.validationRootMap[blobber.ID] = &blobberFile{
			size:           fmr.fileref.Size,
			validationRoot: vr,
		}
		shift := zboxutil.NewUint128(1).Lsh(uint64(fmr.blobberIdx))
		foundMask = foundMask.Or(shift)
	}
	req.consensus = foundMask.CountOnes()
	if !req.isConsensusOk() {
		return nil, fmt.Errorf("consensus_not_met")
	}
	req.downloadMask = foundMask
	return selected.fileref, nil
}
