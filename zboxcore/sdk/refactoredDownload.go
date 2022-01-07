package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/klauspost/reedsolomon"
)

type FileStatus int

const (
	Closed = iota
	Open
)

const Retry = 3

const (
	// EncryptionOverHead File size increases by 16 bytes after encryption. Two checksums i.e. MessageChecksum and OverallChecksum has
	// 128 bytes size each. Checksums are separated by ","
	// So total overhead for each encrypted data is 16 + 128*2 +1 = 273
	EncryptionOverHead = 273
	ChecksumSize       = 257

	// TooManyRequestWaitTime wait for this time to re-request when too_many_requests errors ocucurs
	TooManyRequestWaitTime = time.Millisecond * 100
)

// error codes
const (
	ExceedingFailedBlobber  = "exceeding_failed_blobber"
	ReadCounterUpdate       = "rc_update"
	TooManyRequests         = "too_many_requests"
	ContextCancelled        = "context_cancelled"
	MarshallError           = "marshall_error"
	SigningError            = "error_while_signing"
	ReedSolomonEndocerError = "reedsolomon_endocer_error"
	ErasureReconstructError = "erasure_reconstruct_error"
	ResponseError           = "response_error"
	NoRequiredShards        = "no_required_shards"
	NotEnoughTokens         = "not_enough_tokens"
	Panic                   = "code_panicked"
	InvalidHeader           = "invalid_header"
	DecryptionError         = "decryption_error"
)

//errors
var (
	ErrLessThan67PercentBlobber = errors.New("less_than_67_percent", "less than 67% blobbers able to respond")
	ErrReadCounterUpdate        = errors.New(ReadCounterUpdate, "")
	ErrTooManyRequests          = errors.New(TooManyRequests, "")
	ErrContextCancelled         = errors.New(ContextCancelled, "")
	ErrMarshallError            = errors.New(MarshallError, "")
	ErrSigningError             = errors.New(SigningError, "")
	ErrReedSolomonEncoderError  = errors.New(ReedSolomonEndocerError, "")
	ErrErasureReconstructError  = errors.New(ErasureReconstructError, "")
	ErrFromResponse             = errors.New(ResponseError, "")
	ErrNoRequiredShards         = errors.New(NoRequiredShards, "")
	ErrNotEnoughTokens          = errors.New(NotEnoughTokens, "")
	ErrPanic                    = errors.New(Panic, "")
	ErrInvalidHeader            = errors.New(InvalidHeader, "")
	ErrDecryption               = errors.New(DecryptionError, "")
)

// errors func
var (
	ErrExceedingFailedBlobber = func(failed, parity int) error {
		msg := fmt.Sprintf("number of failed %v blobbers exceeds %v parity shards", failed, parity)
		return errors.New(ExceedingFailedBlobber, msg)
	}
)

//Provide interface similar to io.Reader
//Define errors in this file temporarily

type StreamDownload struct {
	allocationID string
	allocationTx string

	blobbers []*blockchain.StorageNode
	// All error giving blobbers except for too_many_requests, context_deadline
	failedBlobbers           map[int]*blockchain.StorageNode
	fbMu                     *sync.Mutex // mutex to update failedBlobbers
	dataShards, parityShards int
	// Offset Where to start to read from
	offset int64
	// File is whether opened
	opened     bool
	eofReached bool // whether end of file is reached
	//downloadType horizontal --> one block per blobber, vertical --> multiple blocks per blobber
	downloadType    string // vertical or horizontal
	blocksPerMarker int

	authTicket         string
	ownerId            string
	remotePath         string
	pathHash           string
	contentMode        string
	rxPay              bool  // true--> self pays
	chunkSize          int64 // total size of a chunk used to split data to datashards numbers of blobbers
	blockSize          int64 // blockSize, chunkSize/dataShards
	effectiveChunkSize int64 // effective chunk size is different when file is encrypted
	effectiveBlockSize int64 // effective block size is different when file is encrypted
	fileSize           int64 // Actual file size
	// encrypted Is file encrypted
	encrypted bool
	encScheme encryption.EncryptionScheme

	// retry Set this value to retry some failed requests due to too_many_requests, context_cancelled, timeout, etc. errors
	retry int
}

type blobberStreamDownloadRequest struct {
	ctx context.Context
	//blobberIdx Index of blobber in allocation's blobbers list.
	//It indicates either blobber is data or parity shard
	blobberIdx int
	blobberID  string
	blobberUrl string
	//offsetBlock Blocks will be downloaded after this block
	offsetBlock int
	//blocksPerMarker Number of blocks to download in single request/readMarker
	blocksPerMarker int

	sd     *StreamDownload
	result dataStatus
	//retry Number of times data will be requeste to blobber
	retry int
}

func (bl *blobberStreamDownloadRequest) decrypt(enData []byte) ([]byte, error) {
	header := enData[:257]
	header = bytes.Trim(header, "\x00")
	splitChar := []byte(",")
	if len(header) != ChecksumSize {
		return nil, errors.New(InvalidHeader, "incomplete header")
	}
	if !bytes.Contains(header, splitChar) {
		return nil, errors.New(InvalidHeader, "header doesn't contain \",\" character")
	}
	splittedHeader := bytes.Split(header, splitChar)
	messageChecksum, overallChecksum := splittedHeader[0], splittedHeader[1]

	encMsg := encryption.EncryptedMessage{
		EncryptedKey:    bl.sd.encScheme.GetEncryptedKey(),
		EncryptedData:   enData[257:],
		MessageChecksum: string(messageChecksum),
		OverallChecksum: string(overallChecksum),
	}

	return bl.sd.encScheme.Decrypt(&encMsg)
}

func (bl *blobberStreamDownloadRequest) split(buf []byte, lim int64) (chunks [][]byte, err error) {
	if bl.sd.encrypted {
		return bl.splitAndDecrypt(buf, lim)
	}

	var chunk []byte
	chunkSize := int(math.Ceil(float64(len(buf)) / float64(lim)))
	chunks = make([][]byte, 0, chunkSize)

	for int64(len(buf)) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}

	if len(buf) > 0 {
		chunks = append(chunks, buf[:])
	}

	return chunks, nil
}

func (bl *blobberStreamDownloadRequest) splitAndDecrypt(buf []byte, lim int64) (chunks [][]byte, err error) {
	defer func() {
		if errI := recover(); err != nil {
			err := errI.(error)
			err = errors.New(Panic, err.Error())
			chunks = nil
		}
	}()

	var chunk []byte
	chunkSize := int(math.Ceil(float64(len(buf)) / float64(lim)))
	chunks = make([][]byte, 0, chunkSize)
	for int64(len(buf)) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		decryptedChunk, err := bl.decrypt(chunk)
		if err != nil {
			return nil, errors.New(DecryptionError, fmt.Sprintf("Failed decrypting data from blobber %v. Error: %v", bl.blobberUrl, err))
		}

		chunks = append(chunks, decryptedChunk)
	}
	if len(buf) > 0 {
		chunk = buf[:]
		decryptedChunk, err := bl.decrypt(chunk)
		if err != nil {
			return nil, errors.New(DecryptionError, fmt.Sprintf("Failed decrypting data from blobber %v. Error: %v", bl.blobberUrl, err))
		}

		chunks = append(chunks, decryptedChunk)
	}

	return chunks, nil
}

type dataStatus struct {
	err  error
	data [][]byte
	n    int //Check if we can partially get data
}

func (sd *StreamDownload) SetOffset(offset int64) {
	sd.offset = offset
}

// Read io.Reader implementation
func (sd *StreamDownload) Read(p []byte) (n int, err error) {
	if !sd.opened {
		return 0, errors.New("file_closed", "s")
	}

	if sd.eofReached || sd.offset >= sd.fileSize {
		return 0, io.EOF
	}

	if len(sd.failedBlobbers) > sd.parityShards {
		return 0, ErrExceedingFailedBlobber(len(sd.failedBlobbers), sd.parityShards)
	}

	data, err := sd.getDataHorizontal(len(p))

	_ = err  //check for errors
	_ = data //check if eof is reached
	return
}

func (sd *StreamDownload) Close() {
	sd.opened = false
}

// getBlocksRequired Get number of blocks required to download want size
func (sd *StreamDownload) getBlocksRequired(wantSize int) int {
	offsetRemainder := sd.offset % sd.effectiveBlockSize

	return int(math.Ceil(float64(offsetRemainder+int64(wantSize)) / float64(sd.effectiveBlockSize)))
}

// getBlobberStartingIdx return blobber index where offset has reached
func (sd *StreamDownload) getBlobberStartingIdx() int {
	offsetBlock := sd.offset / sd.effectiveBlockSize

	return int(offsetBlock) % sd.dataShards
}

// getBlobberEndIdx return blobber index that has required last block
func (sd *StreamDownload) getBlobberEndIdx(size int) int {
	endSize := sd.offset + int64(size)
	offsetBlock := int(math.Ceil(float64(endSize) / float64(sd.effectiveBlockSize)))

	return offsetBlock % sd.dataShards

}

// getDataOffset return offset value to slice data from 0 to this offset value
func (sd *StreamDownload) getDataOffset(wantSize int) int {
	startingBlock := sd.offset / sd.effectiveBlockSize
	blockOffset := startingBlock * sd.effectiveBlockSize

	return int(sd.offset - blockOffset)
}

func (sd *StreamDownload) getBlobberStartingEndingIdx(size int) (int, int) {
	return sd.getBlobberStartingIdx(), sd.getBlobberEndIdx(size)
}

// getChunksRequired Get number m, that make m*dataShards requests
func (sd *StreamDownload) getChunksRequired(startingIdx, wantSize int) int {
	if startingIdx == 0 {
		return int(math.Ceil(float64(wantSize) / float64(sd.effectiveChunkSize)))
	}
	chunkOffset := int64(startingIdx) * sd.effectiveBlockSize

	return int(math.Ceil((float64(chunkOffset) + float64(wantSize)) / float64(sd.effectiveChunkSize)))
}

func (sd *StreamDownload) getEndOffsetBlock(wantSize int) int {
	newOffset := float64(sd.offset) + float64(wantSize)
	newOffset = math.Min(float64(newOffset), float64(sd.fileSize))

	return int(newOffset) / int(sd.effectiveBlockSize) / sd.dataShards
}

func (sd *StreamDownload) getStartOffsetBlock() int {
	return int(sd.offset / sd.effectiveBlockSize / int64(sd.dataShards))
}

func (sd *StreamDownload) getDataVertical(wantSize int) (data []byte, err error) {
	startOffsetBlock := sd.getStartOffsetBlock()
	endOffsetBlock := sd.getEndOffsetBlock(wantSize)

	totBlocksPerBlobber := endOffsetBlock - int(startOffsetBlock) + 1
	startingIdx := sd.getBlobberStartingIdx()
	chunksRequired := sd.getChunksRequired(startingIdx, wantSize)

	bpm := sd.blocksPerMarker // blocks per marker

	for totBlocksPerBlobber > 0 {
		bpm = int(math.Min(float64(totBlocksPerBlobber), float64(bpm)))
		totBlocksPerBlobber -= bpm

		results := make([]*blobberStreamDownloadRequest, sd.dataShards+sd.parityShards)
		reconstructionRequiredCh := make(chan struct{}, sd.dataShards)
		downloadCompletedCh := make(chan struct{}, sd.dataShards)

		for j := 0; j < sd.dataShards; j++ {
			blobber := sd.blobbers[j]
			if _, ok := sd.failedBlobbers[j]; ok { // For failed blobbers results[j] is nil
				continue
			}

			bsdl := &blobberStreamDownloadRequest{
				blobberID:  blobber.ID,
				blobberUrl: blobber.Baseurl,
				blobberIdx: j,

				offsetBlock:     startOffsetBlock,
				sd:              sd,
				blocksPerMarker: bpm,
			}

			results[j] = bsdl

			go bsdl.downloadData(reconstructionRequiredCh, downloadCompletedCh)
		}

		startOffsetBlock += bpm

		var requiredParityShards int
		for k := 0; k < sd.dataShards; k++ {
			select {
			case <-downloadCompletedCh:
			case <-reconstructionRequiredCh:
				requiredParityShards++
			}
		}

		if requiredParityShards > 0 {
			if requiredParityShards > sd.parityShards {
				return data, ErrExceedingFailedBlobber(requiredParityShards, sd.parityShards)
			}

			err = sd.reconstructVertical(results, requiredParityShards, startOffsetBlock, bpm)
			if err != nil {
				return
			}
		}

		for i := 0; i < chunksRequired; i++ {

			for j := 0; j < sd.dataShards; j++ {
				data = append(data, results[j].result.data[i]...)
			}
		}
	}

	dataOffset := sd.getDataOffset(wantSize)
	data = data[dataOffset:]

	newOffset := sd.offset + int64(wantSize)
	if newOffset >= sd.fileSize {
		sd.eofReached = true
	} else {
		data = data[:newOffset]
	}

	sd.SetOffset(newOffset)
	return
}

func (sd *StreamDownload) getDataHorizontal(wantSize int) (data []byte, err error) {
	startingBlobberIdx := sd.getBlobberStartingIdx()
	offsetBlock := sd.getStartOffsetBlock()
	totalBlocksRequired := sd.getBlocksRequired(wantSize)

	chunkNums := sd.getChunksRequired(startingBlobberIdx, wantSize)

	var blocksRequested int
	for i := 0; i < chunkNums; i++ {
		results := make([]*blobberStreamDownloadRequest, sd.dataShards)
		reconstructionRequiredCh := make(chan struct{}, sd.dataShards)
		downloadCompletedCh := make(chan struct{}, sd.dataShards)
		var count int
		for j := startingBlobberIdx; j < sd.dataShards; j++ {
			blobber := sd.blobbers[j]
			if _, ok := sd.failedBlobbers[j]; ok {
				//give error
				continue
			}

			bsdl := blobberStreamDownloadRequest{
				blobberID:       blobber.ID,
				blobberIdx:      j,
				blobberUrl:      blobber.Baseurl,
				sd:              sd,
				offsetBlock:     offsetBlock,
				blocksPerMarker: sd.blocksPerMarker,
			}

			go bsdl.downloadData(reconstructionRequiredCh, downloadCompletedCh)

			results[j] = &bsdl
			count++
			blocksRequested++
			if blocksRequested == totalBlocksRequired {
				break
			}
		}

		var isReconstructionRequired bool
		for k := 0; k < count; k++ { //Wait for all goroutines to complete
			select {
			case <-downloadCompletedCh:
			case <-reconstructionRequiredCh:
				isReconstructionRequired = true
			}
		}

		if len(sd.failedBlobbers) > sd.parityShards {
			return nil, ErrExceedingFailedBlobber(len(sd.failedBlobbers), sd.parityShards)
		}

		rawData := make([][]byte, sd.dataShards+sd.parityShards)
		var dataShardsCount int
		for k := startingBlobberIdx; k < sd.dataShards; k++ {
			res := results[k]
			if res == nil || res.result.err != nil {
				continue
			}

			rawData[k] = res.result.data[0]
			dataShardsCount++
		}

		if isReconstructionRequired {
			err = sd.reconstructHorizontal(rawData, dataShardsCount)
			if err != nil {
				return
			}
		}

		for i := startingBlobberIdx; i < sd.dataShards; i++ {
			data = append(data, rawData[i]...)
		}

		startingBlobberIdx = 0 //Only first chunk requires initial value other than 0

	}

	// Put block below in Read method; Calculate dataOffset, new offset, etc.
	dataOffset := sd.getDataOffset(wantSize)
	newOffset := sd.offset + int64(wantSize)

	data = data[dataOffset:]
	if newOffset >= sd.fileSize {
		sd.eofReached = true
	} else {
		data = data[:newOffset]
	}

	sd.SetOffset(newOffset)

	return
}

func (sd *StreamDownload) reconstructVertical(results []*blobberStreamDownloadRequest, reqParity, offsetBlock, bpm int) error {
	ctx, ctxCncl := context.WithCancel(context.Background())
	defer ctxCncl()

	requireNextBlobberCh := make(chan struct{}, reqParity)
	downloadCompletedCh := make(chan struct{}, reqParity)
	breakLoopCh := make(chan struct{}, 1)
	var gotRequiredShards bool

	go func() {
		var count int
	outerloop:
		for {
			select {
			case <-downloadCompletedCh:
				count++
				if count == reqParity {
					breakLoopCh <- struct{}{}
					gotRequiredShards = true
					break
				}
			case <-ctx.Done():
				breakLoopCh <- struct{}{}
				break outerloop
			}
		}
	}()

	var requestedShards int
outerloop:
	for i := 0; i < sd.parityShards; i++ {
		idx := i + sd.dataShards
		blobber := sd.blobbers[idx]
		if _, ok := sd.failedBlobbers[idx]; ok {
			continue
		}

		bsdl := &blobberStreamDownloadRequest{
			blobberID:  blobber.ID,
			blobberUrl: blobber.Baseurl,
			blobberIdx: idx,

			offsetBlock:     offsetBlock,
			blocksPerMarker: bpm,
			sd:              sd,
		}

		go bsdl.downloadData(requireNextBlobberCh, downloadCompletedCh)

		results[idx] = bsdl

		requestedShards++
		if requestedShards == reqParity {
			select {
			case <-downloadCompletedCh:
				break outerloop
			case <-requireNextBlobberCh:
				requestedShards--
			}
		}
	}

	if gotRequiredShards {
		enc, err := reedsolomon.New(sd.dataShards, sd.parityShards)
		if err != nil {
			return errors.New("reedsolomon_encoder_error", err.Error())
		}

		for i := 0; i < bpm; i++ {
			rawData := make([][]byte, sd.dataShards+sd.parityShards)
			for j := 0; j < sd.dataShards+sd.parityShards; j++ {
				if results[j] != nil {
					rawData[j] = results[j].result.data[i]
				}
			}

			if err := enc.ReconstructData(rawData); err != nil {
				return errors.New("erasure_reconstruct_error", err.Error())
			}

			for k := 0; k < sd.dataShards; k++ {
				if results[k] == nil {
					results[k] = new(blobberStreamDownloadRequest)
					results[k].result.data[i] = rawData[k]
				}
			}
		}
	}

	return ErrNoRequiredShards
}

func (sd *StreamDownload) reconstructHorizontal(rawData [][]byte, dataShardsCount int) error {
	ctx, ctxCncl := context.WithCancel(context.Background())
	defer ctxCncl()

	var requestedShards int
	requiredShards := sd.dataShards - dataShardsCount
	nextBlobberRequiredChan := make(chan struct{}, requiredShards)
	downloadCompletedChan := make(chan struct{}, requiredShards)
	breakLoopCh := make(chan struct{})
	results := make([]*blobberStreamDownloadRequest, sd.dataShards+sd.parityShards)

	var gotRequiredShards bool

	go func() {
		i := 0
	outerloop:
		for {
			select {
			case <-downloadCompletedChan:
				i++
				if i == requiredShards {
					gotRequiredShards = true
					breakLoopCh <- struct{}{}
					break outerloop
				}
			case <-ctx.Done():
				break outerloop
			}
		}
	}()

outerloop:
	for i := sd.dataShards + sd.parityShards - 1; i >= 0; i-- { //Give priority to parity blobbers for reconstruction
		blobber := sd.blobbers[i]

		if _, ok := sd.failedBlobbers[i]; ok || rawData[i] != nil {
			continue
		}
		bsdl := &blobberStreamDownloadRequest{
			blobberID:  blobber.ID,
			blobberIdx: i,
			blobberUrl: blobber.Baseurl,
			sd:         sd,
		}

		go bsdl.downloadData(nextBlobberRequiredChan, downloadCompletedChan)

		results[i] = bsdl

		requestedShards++
		if requestedShards == requiredShards {
			select {
			case <-breakLoopCh:
				break outerloop
			case <-nextBlobberRequiredChan:
				requestedShards--
				//blocking case
			}
		}

	}

	if gotRequiredShards {
		for _, res := range results {
			rawData[res.blobberIdx] = res.result.data[0]
		}

		enc, err := reedsolomon.New(sd.dataShards, sd.parityShards, reedsolomon.WithAutoGoroutines(int(sd.effectiveBlockSize)))
		if err != nil {
			return errors.New(ReedSolomonEndocerError, err.Error())
		}

		err = enc.ReconstructData(rawData)
		if err != nil {
			return errors.New(ErasureReconstructError, err.Error())
		}
	}

	return ErrNoRequiredShards
}

func (bl *blobberStreamDownloadRequest) downloadData(errCh, successCh chan<- struct{}) {
	for retry := 0; retry < bl.sd.retry; retry++ {

		var latestRC int64
		rm := &marker.ReadMarker{
			ClientID:        client.GetClientID(),
			ClientPublicKey: client.GetClientPublicKey(),
			BlobberID:       bl.blobberID,
			AllocationID:    bl.sd.allocationID,
			//Let's try with allocation owner id
			OwnerID:     bl.sd.ownerId,
			Timestamp:   common.Now(),
			ReadCounter: latestRC + int64(bl.blocksPerMarker),
		}

		if err := rm.Sign(); err != nil {
			bl.result.err = errors.New(SigningError, err.Error())
			return
		}

		rmData, err := json.Marshal(rm)
		if err != nil {
			bl.result.err = errors.New(MarshallError, err.Error())
			return
		}

		body := new(bytes.Buffer)
		formWriter := multipart.NewWriter(body)
		formWriter.WriteField("path_hash", bl.sd.pathHash)
		formWriter.WriteField("block_num", fmt.Sprint(bl.offsetBlock))
		formWriter.WriteField("num_blocks", fmt.Sprint(bl.blocksPerMarker))
		formWriter.WriteField("content", bl.sd.contentMode) //TODO take from struct
		formWriter.WriteField("read_marker", string(rmData))
		formWriter.WriteField("rx_pay", fmt.Sprint(bl.sd.rxPay))
		formWriter.WriteField("auth_token", bl.sd.authTicket)

		formWriter.Close()

		downReq, err := zboxutil.NewDownloadRequest(bl.blobberUrl, bl.sd.allocationID, body)
		if err != nil {
			bl.result.err = err
			return
		}
		downReq.Header.Add("Content-Type", formWriter.FormDataContentType())
		//Update sd.failedBlobbers; update with mutex if failed to get data. Wait for some milliseconds
		//Handle too_many_requests, context_cancelled, timeout, etc errors in this function

		ctx, cncl := context.WithCancel(context.Background())
		err = zboxutil.HttpDo(ctx, cncl, downReq, func(resp *http.Response, err error) error {
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusOK:
				response, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				downloadedBlock := new(downloadBlock)
				downloadedBlock.idx = bl.blobberIdx
				err = json.Unmarshal(response, downloadedBlock)
				if err != nil { //It means response is file data
					downloadedBlock.Success = true
					bl.result.data, err = bl.split(response, bl.sd.blockSize)
					return err
				}

				if !downloadedBlock.Success && downloadedBlock.LatestRM != nil {
					latestRC = downloadedBlock.LatestRM.ReadCounter
					return ErrReadCounterUpdate
				}

			default:
				//
				response, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				if bytes.Contains(response, []byte(NotEnoughTokens)) {
					return errors.New(NotEnoughTokens, string(response))
				}

				return errors.New(ResponseError, string(response))

			}
			return nil
		})

		switch {
		case errors.Is(err, nil):
			successCh <- struct{}{}
			return
		case errors.Is(err, ErrReadCounterUpdate):
			retry = 0 // Retry indefinitely
			//
		case errors.Is(err, ErrTooManyRequests):
			if retry == bl.sd.retry-1 {
				bl.result.err = err
				errCh <- struct{}{}
			}
			time.Sleep(TooManyRequestWaitTime)
			//
		case errors.Is(err, ErrContextCancelled):
			if retry == bl.sd.retry-1 {
				bl.result.err = err
				errCh <- struct{}{}
			}
		default:
			bl.result.err = err
			bl.sd.fbMu.Lock()
			bl.sd.failedBlobbers[bl.blobberIdx] = bl.sd.blobbers[bl.blobberIdx]
			bl.sd.fbMu.Unlock()

			errCh <- struct{}{}
			return
		}
	}
}

// GetDStorageFileReader Get a reader that provides io.Reader interface
func GetDStorageFileReader(allocation *Allocation, ref *fileref.FileRef, contentMode, authTicket string, rxPay bool, retry int) (*StreamDownload, error) {
	downloadRetry := Retry
	if retry > 0 {
		downloadRetry = retry
	}

	var isEncrypted bool
	var effectiveBlockSize, effectiveChunkSize int64
	var encScheme encryption.EncryptionScheme
	if ref.EncryptedKey != "" { // TODO: check for encrypted_key as similar field
		isEncrypted = true
		effectiveBlockSize = ref.ChunkSize - EncryptionOverHead
		effectiveChunkSize = effectiveBlockSize * int64(allocation.DataShards)
		encScheme = encryption.NewEncryptionScheme()
		if _, err := encScheme.Initialize(client.GetClient().Mnemonic); err != nil {
			return nil, err
		}

		if err := encScheme.InitForDecryption("filetype:audio", ref.EncryptedKey); err != nil {
			return nil, err
		}

	}

	return &StreamDownload{
		allocationID:       allocation.ID,
		allocationTx:       allocation.Tx,
		ownerId:            allocation.Owner, // TODO verify ownerId field
		dataShards:         allocation.DataShards,
		parityShards:       allocation.ParityShards,
		blobbers:           allocation.Blobbers,
		encScheme:          encScheme,
		chunkSize:          ref.ChunkSize * int64(allocation.DataShards),
		blockSize:          ref.ChunkSize,
		effectiveChunkSize: effectiveChunkSize,
		effectiveBlockSize: effectiveBlockSize,
		encrypted:          isEncrypted,
		contentMode:        contentMode,
		rxPay:              rxPay,
		remotePath:         ref.Path,
		pathHash:           ref.PathHash,
		fileSize:           ref.ActualFileSize,
		retry:              downloadRetry,
	}, nil
}
