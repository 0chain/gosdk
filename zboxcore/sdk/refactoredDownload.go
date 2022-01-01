package sdk

import (
	"context"
	"fmt"
	"io"
	"math"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/klauspost/reedsolomon"
)

type FileStatus int

const (
	Closed = iota
	Open
)

const Retry = 3

const (
	ExceedingFailedBlobber = "exceeding_failed_blobber"
)

//errors
var (
	errLessThan67PercentBlobber = errors.New("less_than_67_percent", "less than 67% blobbers able to respond")
	ErrExceedingFailedBlobber   = func(failed, parity int) error {
		msg := fmt.Sprintf("number of failed %v blobbers exceeds %v parity shards", failed, parity)
		return errors.New(ExceedingFailedBlobber, msg)
	}
)

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
	Retry  int
	// File is whether opened
	opened     bool
	eofReached bool // whether end of file is reached
}

type downloadFileInfo struct {
	allocationID string
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
	blobberIdx int
	blobberID  string
	blobberUrl string

	fileInfo *downloadFileInfo
	result   dataStatus
}

type dataStatus struct {
	err  error
	data []byte
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

	data, err := sd.getData(len(p))

	_ = err  //check for errors
	_ = data //check if eof is reached
	return
}

func (sd *StreamDownload) Close() {
	sd.opened = false
}

//TODO optimize this function
func (sd *StreamDownload) getStartingBlobberIdx() int {
	offset := sd.offset

	if offset < sd.downloadFileInfo.blockSize {
		return 0
	} else {
		blobNum := 0
		for startSize := int64(0); startSize < offset; startSize += sd.downloadFileInfo.blockSize {
			if blobNum == sd.dataShards {
				blobNum = 0
			} else {
				blobNum++
			}
		}

		return blobNum
	}
}

func (sd *StreamDownload) getData(wantSize int) (data []byte, err error) {
	nextBlobberIdx := sd.getStartingBlobberIdx()

	totalBlocksRequired := int(math.Ceil(float64(wantSize) / float64(sd.blockSize)))
	totalSizeToDownload := int64(totalBlocksRequired) * sd.blockSize

	chunkNums := 1
	if totalSizeToDownload > sd.chunkSize {
		chunkNums = int(math.Ceil(float64((int64(nextBlobberIdx+1)*sd.blockSize)+totalSizeToDownload) / float64(sd.chunkSize)))
	}

	var blocksRequested int
	for i := 0; i < chunkNums; i++ {
		results := make([]*blobberStreamDownloadRequest, sd.dataShards)
		reconstructionRequiredCh := make(chan struct{}, sd.dataShards)
		downloadCompletedCh := make(chan struct{}, sd.dataShards)
		var count int
		for j := nextBlobberIdx; j < sd.dataShards; j++ {
			blobber := sd.blobbers[j]
			if _, ok := sd.failedBlobbers[j]; ok {
				//give error
				continue
			}

			bsdl := blobberStreamDownloadRequest{
				blobberID:  blobber.ID,
				blobberUrl: blobber.Baseurl,
				fileInfo:   sd.downloadFileInfo,
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
		for k := nextBlobberIdx; k < sd.dataShards; k++ {
			res := results[k]
			if res == nil || res.result.err != nil {
				continue
			}

			rawData[k] = res.result.data
			dataShardsCount++
		}

		if isReconstructionRequired {
			err = sd.reconstruct(rawData, dataShardsCount, sd.downloadFileInfo)
			if err != nil {
				return
			}
		}

		for i := nextBlobberIdx; i < sd.dataShards; i++ {
			data = append(data, rawData[i]...)
		}

		nextBlobberIdx = 0 //Only first chunk requires initial value other than 0
		//Reconstruct data; consider offset as well

	}

	newOffset := sd.offset + int64(wantSize)
	data = data[sd.offset:newOffset]

	sd.SetOffset(newOffset)

	return
}

func (sd *StreamDownload) reconstruct(rawData [][]byte, dataShardsCount int, fileInfo *downloadFileInfo) error {
	ctx, ctxCncl := context.WithCancel(context.Background())
	defer ctxCncl()

	var requestedShards int
	requiredShards := sd.dataShards - dataShardsCount
	nextBlobberRequiredChan := make(chan struct{}, requiredShards)
	downloadCompletedChan := make(chan struct{}, requiredShards)
	breakLoopCh := make(chan struct{})
	results := make([]*blobberStreamDownloadRequest, requiredShards)

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
			rawData[res.blobberIdx] = res.result.data
		}

		enc, err := reedsolomon.New(sd.dataShards, sd.parityShards, reedsolomon.WithAutoGoroutines(int(sd.blockSize)))
		if err != nil {
			return errors.New("reedsolomon_encoder_error", err.Error())
		}

		err = enc.ReconstructData(rawData)
		if err != nil {
			return errors.New("erasure_reconstruct_error", err.Error())
		}
	}

	return errors.New("could_not_get_required_shards", "")
}

func (bl *blobberStreamDownloadRequest) downloadData(errCh, downloadCh <-chan struct{}) {

	//Update sd.failed blobbers if failed to get data
	//Handle too_many_requests, context_cancelled, timeout, etc errors in this function

	return
}

// GetDStorageFileReader Get a reader that provides io.Reader interface
func GetDStorageFileReader(allocation *Allocation, ref *fileref.FileRef, authTicket string) *StreamDownload {
	return &StreamDownload{
		dataShards:   allocation.DataShards,
		parityShards: allocation.ParityShards,
		blobbers:     allocation.Blobbers,
		Retry:        Retry,
		downloadFileInfo: &downloadFileInfo{
			allocationID: allocation.ID,
			chunkSize:    ref.ChunkSize * int64(allocation.DataShards),
			blockSize:    ref.ChunkSize,
			fileSize:     ref.ActualFileSize,
			encrypted:    ref.EncryptedKey != "", // TODO: check for encrypted_key as similar field
		},
	}
}

//Why show numblocks double
