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

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
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
	ExceedingFailedBlobber  = "exceeding_failed_blobber"
	ReadCounterUpdate       = "rc_update"
	MarshallError           = "marshall_error"
	SigningError            = "error_while_signing"
	ReedSolomonEndocerError = "reedsolomon_endocer_error"
	ErasureReconstructError = "erasure_reconstruct_error"
	ResponseError           = "response_error"
	NoRequiredShards        = "no_required_shards"
	NotEnoughTokens         = "not_enough_tokens"
)

//errors
var (
	ErrLessThan67PercentBlobber = errors.New("less_than_67_percent", "less than 67% blobbers able to respond")
	ErrReadCounterUpdate        = errors.New(ReadCounterUpdate, "")
	ErrMarshallError            = errors.New(MarshallError, "")
	ErrSigningError             = errors.New(SigningError, "")
	ErrReedSolomonEncoderError  = errors.New(ReedSolomonEndocerError, "")
	ErrErasureReconstructError  = errors.New(ErasureReconstructError, "")
	ErrFromResponse             = errors.New(ResponseError, "")
	ErrNoRequiredShards         = errors.New(NoRequiredShards, "")
	ErrNotEnoughTokens          = errors.New(NotEnoughTokens, "")

	ErrExceedingFailedBlobber = func(failed, parity int) error {
		msg := fmt.Sprintf("number of failed %v blobbers exceeds %v parity shards", failed, parity)
		return errors.New(ExceedingFailedBlobber, msg)
	}
)

func split(buf []byte, lim int64) [][]byte {
	var chunk []byte
	chunkSize := int(math.Ceil(float64(len(buf)) / float64(lim)))
	chunks := make([][]byte, 0, chunkSize)
	for int64(len(buf)) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:])
	}
	return chunks
}

//Provide interface similar to io.Reader
//Define errors in this file temporarily

type StreamDownload struct {
	*downloadFileInfo
	blobbers []*blockchain.StorageNode
	// All error giving blobbers except for too_many_requests, context_deadline
	failedBlobbers           map[int]*blockchain.StorageNode
	fbMu                     sync.Mutex // mutex to update failedBlobbers
	dataShards, parityShards int
	// Offset Where to start to read from
	offset int64
	// File is whether opened
	opened     bool
	eofReached bool // whether end of file is reached

	downloadType    string // vertical or horizontal
	blocksPerMarker int
}

type downloadFileInfo struct {
	ownerId      string
	remotePath   string
	pathHash     string
	allocationID string
	allocationTx string
	contentMode  string
	authTicket   string
	rxPay        bool  // true--> self pays
	chunkSize    int64 // total size of a chunk used to split data to datashards numbers of blobbers
	blockSize    int64 // blockSize, chunkSize/dataShards
	totalChunks  int64 //How many chunks is file divided into
	totalBlocks  int64 //How many blocks is file divided into. Equal to chunks*dataShards
	fileSize     int64
	// encrypted Is file encrypted before uploading
	encrypted bool
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

	fileInfo *downloadFileInfo
	result   dataStatus
	//retry Number of times data will be requeste to blobber
	retry int
}

type dataStatus struct {
	err  error
	data [][]byte
	n    int //Check if we can partially get data
}

type counter struct {
	succeeded int
	failed    int
	mu        sync.Mutex
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
	// if sd.offset == 0 || sd.offset%sd.blockSize == 0 {
	// 	return int(math.Ceil(float64(wantSize) / float64(sd.blockSize)))
	// }

	offsetRemainder := sd.offset % sd.blockSize

	return int(math.Ceil(float64(offsetRemainder+int64(wantSize)) / float64(sd.blockSize)))
}

// getBlobberStartingIdx return blobber index where offset has reached
func (sd *StreamDownload) getBlobberStartingIdx() int {
	offsetBlock := sd.offset / sd.blockSize

	return int(offsetBlock) % sd.dataShards
}

// getBlobberEndIdx return blobber index that has required last block
func (sd *StreamDownload) getBlobberEndIdx(size int) int {
	endSize := sd.offset + int64(size)
	offsetBlock := int(math.Ceil(float64(endSize) / float64(sd.blockSize)))

	return offsetBlock % sd.dataShards

}

// getDataOffset return offset value to slice data from 0 to this offset value
func (sd *StreamDownload) getDataOffset(wantSize int) int {
	startingBlock := sd.offset / sd.blockSize
	blockOffset := startingBlock * sd.blockSize

	return int(sd.offset - blockOffset)
}

func (sd *StreamDownload) getBlobberStartingEndingIdx(size int) (int, int) {
	return sd.getBlobberStartingIdx(), sd.getBlobberEndIdx(size)
}

// getChunksRequired Get number m, that make m*dataShards requests
func (sd *StreamDownload) getChunksRequired(startingIdx, wantSize int) int {
	if startingIdx == 0 {
		return int(math.Ceil(float64(wantSize) / float64(sd.chunkSize)))
	}
	chunkOffset := int64(startingIdx) * sd.blockSize

	return int(math.Ceil((float64(chunkOffset) + float64(wantSize)) / float64(sd.chunkSize)))
}

func (sd *StreamDownload) getEndOffsetBlock(wantSize int) int {
	newOffset := float64(sd.offset) + float64(wantSize)
	newOffset = math.Min(float64(newOffset), float64(sd.fileSize))

	return int(newOffset) / int(sd.blockSize) / sd.dataShards
}

func (sd *StreamDownload) getDataVertical(wantSize int) (data []byte, err error) {
	startOffsetBlock := int(sd.offset / sd.blockSize / int64(sd.dataShards))
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
				fileInfo:        sd.downloadFileInfo,
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

			err = sd.reconstructVertical(results, requiredParityShards, startOffsetBlock, bpm, sd.downloadFileInfo)
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
	if newOffset >= sd.downloadFileInfo.fileSize {
		sd.eofReached = true
	} else {
		data = data[:newOffset]
	}

	sd.SetOffset(newOffset)
	return
}

func (sd *StreamDownload) getDataHorizontal(wantSize int) (data []byte, err error) {
	startingBlobberIdx := sd.getBlobberStartingIdx()
	offsetBlock := sd.offset / sd.blockSize / int64(sd.dataShards)
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
				fileInfo:        sd.downloadFileInfo,
				offsetBlock:     int(offsetBlock),
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
			err = sd.reconstructHorizontal(rawData, dataShardsCount, sd.downloadFileInfo)
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
	if newOffset >= sd.downloadFileInfo.fileSize {
		sd.eofReached = true
	} else {
		data = data[:newOffset]
	}

	sd.SetOffset(newOffset)

	return
}

func (sd *StreamDownload) reconstructVertical(results []*blobberStreamDownloadRequest, reqParity, offsetBlock, bpm int, fileInfo *downloadFileInfo) error {
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
			fileInfo:        sd.downloadFileInfo,
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

	return errors.New("could_not_get_required_shards", "")
}

func (sd *StreamDownload) reconstructHorizontal(rawData [][]byte, dataShardsCount int, fileInfo *downloadFileInfo) error {
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
			fileInfo:   fileInfo,
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

		enc, err := reedsolomon.New(sd.dataShards, sd.parityShards, reedsolomon.WithAutoGoroutines(int(sd.blockSize)))
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

func (bl *blobberStreamDownloadRequest) downloadData(errCh, downloadCh chan<- struct{}) {
	for i := 0; i < bl.fileInfo.retry; i++ {

		var latestRC int64
		rm := &marker.ReadMarker{
			ClientID:        client.GetClientID(),
			ClientPublicKey: client.GetClientPublicKey(),
			BlobberID:       bl.blobberID,
			AllocationID:    bl.fileInfo.allocationID,
			//Let's try with allocation owner id
			OwnerID:     bl.fileInfo.ownerId,
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
		formWriter.WriteField("path_hash", bl.fileInfo.pathHash)
		formWriter.WriteField("block_num", fmt.Sprint(bl.offsetBlock))
		formWriter.WriteField("num_blocks", fmt.Sprint(bl.blocksPerMarker))
		formWriter.WriteField("content", bl.fileInfo.contentMode) //TODO take from struct
		formWriter.WriteField("read_marker", string(rmData))
		formWriter.WriteField("rx_pay", fmt.Sprint(bl.fileInfo.rxPay))
		formWriter.WriteField("auth_token", bl.fileInfo.authTicket)

		formWriter.Close()

		downReq, err := zboxutil.NewDownloadRequest(bl.blobberUrl, bl.fileInfo.allocationID, body)
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
					bl.result.data = split(response, bl.fileInfo.blockSize)
					return nil
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
		case errors.Is(err, ErrReadCounterUpdate):
		case errors.Is(err, ErrNotEnoughTokens):
			bl.result.err = err
			errCh <- struct{}{}
			return
			//
		case errors.Is(err, nil):
			//
		}
	}
}

// GetDStorageFileReader Get a reader that provides io.Reader interface
func GetDStorageFileReader(allocation *Allocation, ref *fileref.FileRef, contentMode, authTicket string, rxPay bool, retry int) *StreamDownload {
	downloadRetry := Retry
	if retry > 0 {
		downloadRetry = retry
	}

	return &StreamDownload{
		dataShards:   allocation.DataShards,
		parityShards: allocation.ParityShards,
		blobbers:     allocation.Blobbers,
		downloadFileInfo: &downloadFileInfo{
			ownerId:      allocation.Owner, // TODO verify ownerId field
			contentMode:  contentMode,
			rxPay:        rxPay,
			remotePath:   ref.Path,
			pathHash:     ref.PathHash,
			allocationID: allocation.ID,
			allocationTx: allocation.Tx,
			chunkSize:    ref.ChunkSize * int64(allocation.DataShards),
			blockSize:    ref.ChunkSize,
			fileSize:     ref.ActualFileSize,
			encrypted:    ref.EncryptedKey != "", // TODO: check for encrypted_key as similar field
			retry:        downloadRetry,
		},
	}
}
