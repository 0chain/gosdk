package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/klauspost/reedsolomon"
	"go.dedis.ch/kyber/v3/group/edwards25519"
)

type FileStatus int

const (
	Closed = iota
	Open
)

const Retry = 3

const (
	// EncryptionOverHead File size increases by 16 bytes after encryption. Two checksums i.e. MessageChecksum and OverallChecksum has
	// 128 bytes size each.
	// So total overhead for each encrypted data is 16 + 128*2 = 272
	EncryptionOverHead = 272
	ChecksumSize       = 256
	HeaderSize         = 128

	// TooManyRequestWaitTime wait for this time to re-request when too_many_requests errors ocucurs
	TooManyRequestWaitTime = time.Millisecond * 100

	Vertical   = "vertical"
	Horizontal = "horizontal"
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
	UnknownDownloadType     = "unknown_download_type"
	InvalidBlocksPerMarker  = "invalid_blocks_per_marker"
	ReDecryptUnmarshallFail = "redecrypt_unmarshall_fail"
	ReDecryptionFail        = "redecryption_fail"
	InvalidRead             = "invalid_read"
	InvalidDownloadType     = "invalid_download_type"
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
	ErrUnknownDownloadType      = errors.New(UnknownDownloadType, "")
	ErrInvalidBlocksPerMarker   = errors.New(InvalidBlocksPerMarker, "")
	ErrReDecryptUnmarshallFail  = errors.New(ReDecryptUnmarshallFail, "")
	ErrReDecryptionFail         = errors.New(ReDecryptionFail, "")
	ErrInvalidRead              = errors.New(InvalidRead, "want_size is <= 0")
)

// errors func
var (
	ErrExceedingFailedBlobber = func(failed, parity int) error {
		msg := fmt.Sprintf("number of failed %v blobbers exceeds %v parity shards", failed, parity)
		return errors.New(ExceedingFailedBlobber, msg)
	}

	ErrInvalidDownloadType = func(downloadType string) error {
		msg := fmt.Sprintf("%v download type is not supported", downloadType)
		return errors.New(InvalidDownloadType, msg)
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
	//retry Number of times data will be requested to blobber
	retry int
}

func (bl *blobberStreamDownloadRequest) proxyDecrypt(enData []byte) ([]byte, error) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	reEncMessage := &encryption.ReEncryptedMessage{
		D1: suite.Point(),
		D4: suite.Point(),
		D5: suite.Point(),
	}
	if err := reEncMessage.Unmarshal(enData); err != nil {
		return nil, errors.New(ReDecryptUnmarshallFail, err.Error())
	}

	decryptedData, err := bl.sd.encScheme.ReDecrypt(reEncMessage)
	if err != nil {
		return nil, errors.New(ReDecryptionFail, err.Error())
	}

	return decryptedData, nil
}

func (bl *blobberStreamDownloadRequest) decrypt(enData []byte) ([]byte, error) {
	if bl.sd.authTicket != "" {
		return bl.proxyDecrypt(enData)
	}

	header := enData[:ChecksumSize]
	header = bytes.Trim(header, "\x00")
	// splitChar := []byte(",")
	if len(header) != ChecksumSize {
		return nil, errors.New(InvalidHeader, "incomplete header")
	}
	// if !bytes.Contains(header, splitChar) {
	// 	return nil, errors.New(InvalidHeader, "header doesn't contain \",\" character")
	// }
	// splittedHeader := bytes.Split(header, splitChar)

	encMsg := encryption.EncryptedMessage{
		EncryptedKey:    bl.sd.encScheme.GetEncryptedKey(),
		EncryptedData:   enData[ChecksumSize:],
		MessageChecksum: string(header[:HeaderSize]),
		OverallChecksum: string(header[HeaderSize:]),
	}

	data, err := bl.sd.encScheme.Decrypt(&encMsg)
	if err != nil {
		return nil, errors.New(DecryptionError, err.Error())
	}
	return data, nil
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

type dataStatus struct {
	err  error
	data [][]byte
	n    int //Check if we can partially get data
}

func (sd *StreamDownload) SetOffset(offset int64) {
	sd.offset = offset
}

// getDownloadType get download type and blocks per marker that fits the download requirement in optimal way
// It tries to get at max 100MB data per request
func (sd *StreamDownload) getDownloadType(wantSize int) (downloadType string, bpm int) {
	if sd.blockSize >= 100*MB {
		return Horizontal, 1
	}

	startingIdx := sd.getBlobberStartingIdx()
	chunksRequired := sd.getChunksRequired(startingIdx, wantSize)
	if chunksRequired == 1 {
		return Horizontal, 1
	}

	return Vertical, int(100 * MB / (sd.blockSize))
}

// Read io.Reader implementation
func (sd *StreamDownload) Read(p []byte) (n int, err error) {
	if !sd.opened {
		return 0, errors.New("file_closed", "")
	}

	if sd.eofReached || sd.offset >= sd.fileSize {
		return 0, io.EOF
	}

	if len(sd.failedBlobbers) > sd.parityShards {
		return 0, ErrExceedingFailedBlobber(len(sd.failedBlobbers), sd.parityShards)
	}

	wantSize := int(math.Min(float64(len(p)), float64(sd.fileSize-sd.offset)))
	if wantSize <= 0 {
		return 0, ErrInvalidRead
	}

	downloadType := sd.downloadType
	if downloadType == "" {
		downloadType, sd.blocksPerMarker = sd.getDownloadType(wantSize)
	}

	var data []byte
	switch downloadType {
	case Vertical:
		data, err = sd.getDataVertical(wantSize)
	case Horizontal:
		data, err = sd.getDataHorizontal(wantSize)
	default:
		return 0, ErrInvalidDownloadType(downloadType)
	}

	n = len(data)
	copy(p[:n], data)

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
	return int(sd.offset/sd.effectiveBlockSize/int64(sd.dataShards)) + 1
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
	newOffset := sd.offset + int64(wantSize)

	data = data[dataOffset:]
	if newOffset >= sd.fileSize {
		sd.eofReached = true
	}
	lastIdx := int(math.Min(float64(len(data)), float64(wantSize)))
	data = data[:lastIdx]

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
				reconstructionRequiredCh <- struct{}{}
				continue
			}

			bsdl := blobberStreamDownloadRequest{
				blobberID:       blobber.ID,
				blobberIdx:      j,
				blobberUrl:      blobber.Baseurl,
				sd:              sd,
				offsetBlock:     offsetBlock,
				blocksPerMarker: 1,
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
	}
	lastIdx := int(math.Min(float64(len(data)), float64(wantSize)))
	data = data[:lastIdx]

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
					break outerloop
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
			blobberID:       blobber.ID,
			blobberIdx:      i,
			blobberUrl:      blobber.Baseurl,
			sd:              sd,
			blocksPerMarker: 1,
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

		return nil
	}

	return ErrNoRequiredShards
}

func (bl *blobberStreamDownloadRequest) downloadData(errCh, successCh chan<- struct{}) {
	for retry := 0; retry < bl.sd.retry; retry++ {
		rm := &marker.ReadMarker{
			ClientID:        client.GetClientID(),
			ClientPublicKey: client.GetClientPublicKey(),
			BlobberID:       bl.blobberID,
			AllocationID:    bl.sd.allocationID,
			//Let's try with allocation owner id
			OwnerID:   bl.sd.ownerId,
			Timestamp: common.Now(),
			ReadSize:  int64(bl.sd.blocksPerMarker) * (bl.sd.chunkSize),
		}

		if err := rm.Sign(); err != nil {
			bl.result.err = errors.New(SigningError, err.Error())
			errCh <- struct{}{}
			return
		}

		rmData, err := json.Marshal(rm)
		if err != nil {
			bl.result.err = errors.New(MarshallError, err.Error())
			errCh <- struct{}{}
			return
		}

		downReq, err := zboxutil.NewDownloadRequest(bl.blobberUrl, bl.sd.allocationTx)
		if err != nil {
			bl.result.err = err
			errCh <- struct{}{}
			return
		}

		header := DownloadRequestHeader{
			PathHash:     bl.sd.pathHash,
			RxPay:        bl.sd.rxPay,
			BlockNum:     int64(bl.offsetBlock),
			NumBlocks:    int64(bl.blocksPerMarker),
			ReadMarker:   rmData,
			AuthToken:    []byte(bl.sd.authTicket),
			DownloadMode: bl.sd.contentMode,
		}

		header.ToHeader(downReq)

		// downReq.Header.Add("Content-Type", formWriter.FormDataContentType())
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

				if !downloadedBlock.Success {
					if downloadedBlock.err != nil {
						return errors.New("download_error", downloadedBlock.err.Error())
					}
					return errors.New("download_error", "unknown error which downloading data")
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
func GetDStorageFileReader(allocation *Allocation, ref *ORef, sdo *StreamDownloadOption) (*StreamDownload, error) {
	switch sdo.DownloadType {
	case Horizontal, "":
	case Vertical:
		if sdo.BlocksPerMarker == 0 {
			return nil, errors.New(InvalidBlocksPerMarker, "blocks per marker value should be greater than 0")
		}
	default:
		return nil, ErrInvalidDownloadType(sdo.DownloadType)
	}

	downloadRetry := Retry
	if sdo.Retry > 0 {
		downloadRetry = sdo.Retry
	}

	var isEncrypted bool
	effectiveBlockSize := ref.ChunkSize
	effectiveChunkSize := int64(allocation.DataShards) * ref.ChunkSize

	var encScheme encryption.EncryptionScheme
	if ref.EncryptedKey != "" {
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
		opened:             true,
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
		remotePath:         ref.Path,
		pathHash:           ref.PathHash,
		fileSize:           ref.ActualFileSize,
		authTicket:         sdo.AuthTicket,
		contentMode:        sdo.ContentMode,
		rxPay:              sdo.RxPay,
		downloadType:       sdo.DownloadType,
		blocksPerMarker:    int(sdo.BlocksPerMarker),
		retry:              downloadRetry,
		fbMu:               &sync.Mutex{},
		failedBlobbers:     make(map[int]*blockchain.StorageNode),
	}, nil
}

// StreamDownloadOption options that manipulate stream download
type StreamDownloadOption struct {
	ContentMode     string
	AuthTicket      string
	DownloadType    string // vertical, horizontail or ""
	RxPay           bool
	Retry           int
	BlocksPerMarker uint // Number of blocks to download per request
}
