package sdk

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"errors"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/common"
	coreEncryption "github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/google/uuid"
	"github.com/klauspost/reedsolomon"
	"go.uber.org/zap"
)

const (
	DefaultUploadTimeOut = 2 * time.Minute
)

var (
	// DefaultHashFunc default hash method for stream merkle tree
	DefaultHashFunc = func(left, right string) string {
		return coreEncryption.Hash(left + right)
	}

	ErrInvalidChunkSize              = errors.New("chunk: chunk size is too small. it must greater than 272 if file is uploaded with encryption")
	ErrNoEnoughSpaceLeftInAllocation = errors.New("alloc: no enough space left in allocation")
)

// DefaultChunkSize default chunk size for file and thumbnail
const DefaultChunkSize = 64 * 1024

const (
	// EncryptedDataPaddingSize additional bytes to save encrypted data
	EncryptedDataPaddingSize = 16
	// EncryptionHeaderSize encryption header size in chunk: PRE.MessageChecksum(128)+PRE.OverallChecksum(128)
	EncryptionHeaderSize = 128 + 128
	// ReEncryptionHeaderSize re-encryption header size in chunk
	ReEncryptionHeaderSize = 256
)

/*
  CreateChunkedUpload create a ChunkedUpload instance

	Caller should be careful about fileReader parameter
	io.ErrUnexpectedEOF might mean that source has completely been exhausted or there is some error
	so that source could not fill up the buffer. Due this ambiguity it is responsibility of
	developer to provide new io.Reader that sends io.EOF when source has been all read.
	For example:
		func newReader(source io.Reader) *EReader {
			return &EReader{source}
		}

		type EReader struct {
			io.Reader
		}

		func (r *EReader) Read(p []byte) (n int, err error) {
			if n, err = io.ReadAtLeast(r.Reader, p, len(p)); err != nil {
				if errors.Is(err, io.ErrUnexpectedEOF) {
					return n, io.EOF
				}
			}
			return
		}

*/

func CreateChunkedUpload(
	workdir string, allocationObj *Allocation,
	fileMeta FileMeta, fileReader io.Reader,
	isUpdate, isRepair bool,
	webStreaming bool, connectionId string,
	opts ...ChunkedUploadOption,
) (*ChunkedUpload, error) {

	if allocationObj == nil {
		return nil, thrown.Throw(constants.ErrInvalidParameter, "allocationObj")
	}

	if !isUpdate && !allocationObj.CanUpload() || isUpdate && !allocationObj.CanUpdate() {
		return nil, thrown.Throw(constants.ErrFileOptionNotPermitted, "file_option_not_permitted ")
	}

	if webStreaming {
		newFileReader, newFileMeta, err := TranscodeWebStreaming(fileReader, fileMeta)
		if err != nil {
			return nil, thrown.New("upload_failed", err.Error())
		}
		fileMeta = *newFileMeta
		fileReader = newFileReader
	}

	err := ValidateRemoteFileName(fileMeta.RemoteName)
	if err != nil {
		return nil, err
	}

	opCode := OpUpload
	spaceLeft := allocationObj.Size
	if allocationObj.Stats != nil {
		spaceLeft -= allocationObj.Stats.UsedSize
	}

	if isUpdate {
		f, err := allocationObj.GetFileMeta(fileMeta.RemotePath)
		if err != nil {
			return nil, err
		}
		spaceLeft += f.ActualFileSize
		opCode = OpUpdate
	}

	if fileMeta.ActualSize > spaceLeft {
		return nil, ErrNoEnoughSpaceLeftInAllocation
	}

	consensus := Consensus{
		RWMutex:         &sync.RWMutex{},
		consensusThresh: allocationObj.consensusThreshold,
		fullconsensus:   allocationObj.fullconsensus,
	}

	uploadMask := zboxutil.NewUint128(1).Lsh(uint64(len(allocationObj.Blobbers))).Sub64(1)
	if isRepair {
		opCode = OpUpdate
		consensus.fullconsensus = uploadMask.CountOnes()
		consensus.consensusThresh = 1
	}

	su := &ChunkedUpload{
		allocationObj: allocationObj,
		fileMeta:      fileMeta,
		fileReader:    fileReader,

		uploadMask:      uploadMask,
		chunkSize:       DefaultChunkSize,
		chunkNumber:     1,
		encryptOnUpload: false,
		webStreaming:    false,

		consensus:     consensus, //nolint
		uploadTimeOut: DefaultUploadTimeOut,
		commitTimeOut: DefaultUploadTimeOut,
		maskMu:        &sync.Mutex{},
		opCode:        opCode,
	}

	su.ctx, su.ctxCncl = context.WithCancel(allocationObj.ctx)

	if isUpdate {
		su.httpMethod = http.MethodPut
		su.buildChange = func(ref *fileref.FileRef, _ uuid.UUID, ts common.Timestamp) allocationchange.AllocationChange {
			change := &allocationchange.UpdateFileChange{}
			change.NewFile = ref
			change.NumBlocks = ref.NumBlocks
			change.Operation = constants.FileOperationUpdate
			change.Size = ref.Size
			return change
		}
	} else {
		su.httpMethod = http.MethodPost
		su.buildChange = func(ref *fileref.FileRef, uid uuid.UUID, ts common.Timestamp) allocationchange.AllocationChange {
			change := &allocationchange.NewFileChange{}
			change.File = ref
			change.NumBlocks = ref.NumBlocks
			change.Operation = constants.FileOperationInsert
			change.Size = ref.Size
			change.Uuid = uid
			return change
		}
	}

	su.workdir = filepath.Join(workdir, ".zcn")

	//create upload folder to save progress
	err = sys.Files.MkdirAll(filepath.Join(su.workdir, "upload"), 0744)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(su)
	}
	if su.progressStorer == nil {
		su.progressStorer = createFsChunkedUploadProgress(context.Background())
	}

	su.loadProgress()

	su.fileHasher = CreateHasher(getShardSize(su.fileMeta.ActualSize, su.allocationObj.DataShards, su.encryptOnUpload))

	// encrypt option has been changed. upload it from scratch
	// chunkSize has been changed. upload it from scratch

	su.createUploadProgress(connectionId)

	su.fileErasureEncoder, err = reedsolomon.New(
		su.allocationObj.DataShards,
		su.allocationObj.ParityShards,
		reedsolomon.WithAutoGoroutines(int(su.chunkSize)),
	)
	if err != nil {
		return nil, err
	}

	if su.encryptOnUpload {
		su.fileEncscheme = su.createEncscheme()

		if su.chunkSize <= EncryptionHeaderSize+EncryptedDataPaddingSize {
			return nil, ErrInvalidChunkSize
		}

	}

	su.writeMarkerMutex, err = CreateWriteMarkerMutex(client.GetClient(), su.allocationObj)
	if err != nil {
		return nil, err
	}

	blobbers := su.allocationObj.Blobbers
	if len(blobbers) == 0 {
		return nil, thrown.New("no_blobbers", "Unable to find blobbers")
	}

	su.blobbers = make([]*ChunkedUploadBlobber, len(blobbers))

	for i := 0; i < len(blobbers); i++ {

		su.blobbers[i] = &ChunkedUploadBlobber{
			writeMarkerMutex: su.writeMarkerMutex,
			progress:         su.progress.Blobbers[i],
			blobber:          su.allocationObj.Blobbers[i],
			fileRef: &fileref.FileRef{
				Ref: fileref.Ref{
					Name:         su.fileMeta.RemoteName,
					Path:         su.fileMeta.RemotePath,
					Type:         fileref.FILE,
					AllocationID: su.allocationObj.ID,
				},
			},
		}
	}
	cReader, err := createChunkReader(su.fileReader, fileMeta.ActualSize, int64(su.chunkSize), su.allocationObj.DataShards, su.encryptOnUpload, su.uploadMask, su.fileErasureEncoder, su.fileEncscheme, su.fileHasher)

	if err != nil {
		return nil, err
	}

	su.chunkReader = cReader

	su.formBuilder = CreateChunkedUploadFormBuilder()

	su.isRepair = isRepair

	return su, nil

}

// ChunkedUpload upload manager with chunked upload feature
type ChunkedUpload struct {
	consensus Consensus

	workdir string

	allocationObj  *Allocation
	progress       UploadProgress
	progressStorer ChunkedUploadProgressStorer

	uploadMask zboxutil.Uint128

	// httpMethod POST = Upload File / PUT = Update file
	httpMethod  string
	buildChange func(ref *fileref.FileRef,
		uid uuid.UUID, timestamp common.Timestamp) allocationchange.AllocationChange

	fileMeta           FileMeta
	fileReader         io.Reader
	fileErasureEncoder reedsolomon.Encoder
	fileEncscheme      encryption.EncryptionScheme
	fileHasher         Hasher

	thumbnailBytes         []byte
	thumbailErasureEncoder reedsolomon.Encoder

	chunkReader ChunkedUploadChunkReader
	formBuilder ChunkedUploadFormBuilder

	// encryptOnUpload encrypt data on upload or not.
	encryptOnUpload bool
	// webStreaming whether data has to be encoded.
	webStreaming bool
	// chunkSize how much bytes a chunk has. 64KB is default value.
	chunkSize int64
	// chunkNumber the number of chunks in a http upload request. 1 is default value
	chunkNumber int

	// shardUploadedSize how much bytes a shard has. it is original size
	shardUploadedSize int64
	// shardUploadedThumbnailSize how much thumbnail bytes a shard has. it is original size
	shardUploadedThumbnailSize int64

	// statusCallback trigger progress on StatusCallback
	statusCallback StatusCallback

	blobbers []*ChunkedUploadBlobber

	writeMarkerMutex *WriteMarkerMutex

	// isRepair identifies if upload is repair operation
	isRepair bool

	opCode        int
	uploadTimeOut time.Duration
	commitTimeOut time.Duration
	maskMu        *sync.Mutex
	ctx           context.Context
	ctxCncl       context.CancelFunc
}

// progressID build local progress id with [allocationid]_[Hash(LocalPath+"_"+RemotePath)]_[RemoteName] format
func (su *ChunkedUpload) progressID() string {

	if len(su.allocationObj.ID) > 8 {
		return filepath.Join(su.workdir, "upload", su.allocationObj.ID[:8]+"_"+su.fileMeta.FileID())
	}

	return filepath.Join(su.workdir, "upload", su.allocationObj.ID+"_"+su.fileMeta.FileID())
}

// loadProgress load progress from ~/.zcn/upload/[progressID]
func (su *ChunkedUpload) loadProgress() {
	// ChunkIndex starts with 0, so default value should be -1
	su.progress.ChunkIndex = -1

	progressID := su.progressID()

	progress := su.progressStorer.Load(progressID)

	if progress != nil {
		su.progress = *progress
		su.progress.ID = progressID
	}
}

// saveProgress save progress to ~/.zcn/upload/[progressID]
func (su *ChunkedUpload) saveProgress() {
	su.progressStorer.Save(su.progress)
}

// removeProgress remove progress info once it is done
func (su *ChunkedUpload) removeProgress() {
	su.progressStorer.Remove(su.progress.ID) //nolint
}

// createUploadProgress create a new UploadProgress
func (su *ChunkedUpload) createUploadProgress(connectionId string) {
	if su.progress.ChunkSize == 0 {
		su.progress = UploadProgress{ConnectionID: connectionId,
			ChunkIndex:   -1,
			ChunkSize:    su.chunkSize,
			UploadLength: 0,
		}
	}
	su.progress.Blobbers = make([]*UploadBlobberStatus, common.MustAddInt(su.allocationObj.DataShards, su.allocationObj.ParityShards))

	for i := 0; i < len(su.progress.Blobbers); i++ {
		su.progress.Blobbers[i] = &UploadBlobberStatus{
			Hasher: CreateHasher(getShardSize(su.fileMeta.ActualSize, su.allocationObj.DataShards, su.encryptOnUpload)),
		}
	}

	su.progress.ID = su.progressID()
}

func (su *ChunkedUpload) createEncscheme() encryption.EncryptionScheme {
	encscheme := encryption.NewEncryptionScheme()

	if len(su.progress.EncryptPrivateKey) > 0 {

		privateKey, _ := hex.DecodeString(su.progress.EncryptPrivateKey)

		err := encscheme.InitializeWithPrivateKey(privateKey)
		if err != nil {
			return nil
		}
	} else {
		privateKey, err := encscheme.Initialize(client.GetClient().Mnemonic)
		if err != nil {
			return nil
		}

		su.progress.EncryptPrivateKey = hex.EncodeToString(privateKey)
	}

	encscheme.InitForEncryption("filetype:audio")

	return encscheme
}

func (su *ChunkedUpload) process() error {
	if su.statusCallback != nil {
		su.statusCallback.Started(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, int(su.fileMeta.ActualSize)+int(su.fileMeta.ActualThumbnailSize))
	}

	for {

		startReadChunks := time.Now()
		chunks, err := su.readChunks(su.chunkNumber)
		elapsedReadChunks := time.Since(startReadChunks)

		// chunk, err := su.chunkReader.Next()
		if err != nil {
			if su.statusCallback != nil {
				su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, err)
			}
			return err
		}
		//logger.Logger.Debug("Read chunk #", chunk.Index)

		su.shardUploadedSize += chunks.totalFragmentSize
		su.progress.UploadLength += chunks.totalReadSize

		if chunks.isFinal {
			su.fileMeta.ActualHash, err = su.fileHasher.GetFileHash()
			if err != nil {
				if su.statusCallback != nil {
					su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, err)
				}
				return err
			}

			if su.fileMeta.ActualSize == 0 {
				su.fileMeta.ActualSize = su.progress.UploadLength
			} else if su.fileMeta.ActualSize != su.progress.UploadLength && su.thumbnailBytes == nil {
				if su.statusCallback != nil {
					su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, thrown.New("upload_failed", "Upload failed. Uploaded size does not match with actual size: "+fmt.Sprintf("%d != %d", su.fileMeta.ActualSize, su.progress.UploadLength)))
				}
				return thrown.New("upload_failed", "Upload failed. Uploaded size does not match with actual size: "+fmt.Sprintf("%d != %d", su.fileMeta.ActualSize, su.progress.UploadLength))
			}
		}

		//chunk has not be uploaded yet
		if chunks.chunkEndIndex > su.progress.ChunkIndex {
			startProcessUpload := time.Now()
			err = su.processUpload(
				chunks.chunkStartIndex, chunks.chunkEndIndex,
				chunks.fileShards, chunks.thumbnailShards,
				chunks.isFinal, chunks.totalReadSize,
			)
			elapsedProcessUpload := time.Since(startProcessUpload)
			if err != nil {
				if su.statusCallback != nil {
					su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, err)
				}
				return err
			}

			logger.Logger.Info(fmt.Sprintf(" >>>> [process] Timings: readChunks = %d ms, processUpload = %d ms",
				elapsedReadChunks.Milliseconds(),
				elapsedProcessUpload.Milliseconds(),
			))
		}

		// last chunk might 0 with io.EOF
		// https://stackoverflow.com/questions/41208359/how-to-test-eof-on-io-reader-in-go
		if chunks.totalReadSize > 0 && chunks.chunkEndIndex > su.progress.ChunkIndex {
			su.progress.ChunkIndex = chunks.chunkEndIndex
			su.saveProgress()

			if su.statusCallback != nil {
				su.statusCallback.InProgress(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, int(su.progress.UploadLength), nil)
			}
		}

		if chunks.isFinal {
			break
		}
	}
	return nil
}

// Start start/resume upload
func (su *ChunkedUpload) Start() error {
	now := time.Now()
	defer su.ctxCncl()

	err := su.process()
	if err != nil {
		return err
	}
	elapsedProcess := time.Since(now)
	logger.Logger.Info("Completed the upload. Submitting for commit")

	blobbers := make([]*blockchain.StorageNode, len(su.blobbers))
	for i, b := range su.blobbers {
		blobbers[i] = b.blobber
	}

	err = su.writeMarkerMutex.Lock(
		su.ctx, &su.uploadMask, su.maskMu,
		blobbers, &su.consensus, 0, su.uploadTimeOut,
		su.progress.ConnectionID)

	if err != nil {
		if su.statusCallback != nil {
			su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, err)
		}
		return err
	}
	elapsedLock := time.Since(now) - elapsedProcess

	defer su.writeMarkerMutex.Unlock(
		su.ctx, su.uploadMask, blobbers, su.uploadTimeOut, su.progress.ConnectionID) //nolint: errcheck

	//Check if the allocation is to be repaired or rolled back
	// status, err := su.allocationObj.CheckAllocStatus()
	// if err != nil {
	// 	logger.Logger.Error("Error checking allocation status", err)
	// 	if su.statusCallback != nil {
	// 		su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, err)
	// 	}
	// 	return err
	// }

	// if status == Repair {
	// 	logger.Logger.Info("Repairing allocation")
	// 	// err = su.allocationObj.RepairAlloc(su.statusCallback)
	// 	// if err != nil {
	// 	// 	return err
	// 	// }
	// }
	// if status != Commit {
	// 	return ErrRetryOperation
	// }

	defer func() {
		elapsedProcessCommit := time.Since(now) - elapsedProcess - elapsedLock
		logger.Logger.Info("[ChunkedUpload - start] Timings",
			zap.String("allocation_id", su.allocationObj.ID),
			zap.String("ChunkedUpload - process", fmt.Sprintf("%d ms", elapsedProcess.Milliseconds())),
			zap.String("ChunkedUpload - Lock", fmt.Sprintf("%d ms", elapsedLock.Milliseconds())),
			zap.String("ChunkedUpload - processCommit", fmt.Sprintf("%d ms", elapsedProcessCommit.Milliseconds())),
		)
	}()

	return su.processCommit()
}

func (su *ChunkedUpload) readChunks(num int) (*batchChunksData, error) {

	data := &batchChunksData{
		chunkStartIndex: -1,
		chunkEndIndex:   -1,
	}

	for i := 0; i < num; i++ {
		chunk, err := su.chunkReader.Next()

		if err != nil {
			return nil, err
		}
		//logger.Logger.Debug("Read chunk #", chunk.Index)
		if i == 0 {
			data.chunkStartIndex = chunk.Index
			data.chunkEndIndex = chunk.Index
		} else {
			data.chunkEndIndex = chunk.Index
		}

		data.totalFragmentSize += chunk.FragmentSize
		data.totalReadSize += chunk.ReadSize

		// upload entire thumbnail in first chunk request only
		if chunk.Index == 0 && len(su.thumbnailBytes) > 0 {
			data.totalReadSize += int64(su.fileMeta.ActualThumbnailSize)

			data.thumbnailShards, err = su.chunkReader.Read(su.thumbnailBytes)
			if err != nil {
				return nil, err
			}
		}

		if data.fileShards == nil {
			data.fileShards = make([]blobberShards, len(chunk.Fragments))
		}

		// concact blobber's fragments

		var pos uint64
		for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
			pos = uint64(i.TrailingZeros())
			data.fileShards[pos] = append(data.fileShards[pos], chunk.Fragments[pos])
			err = su.progress.Blobbers[pos].Hasher.WriteToFixedMT(chunk.Fragments[pos])
			if err != nil {
				return nil, err
			}
			err = su.progress.Blobbers[pos].Hasher.WriteToValidationMT(chunk.Fragments[pos])
			if err != nil {
				return nil, err
			}
		}

		if chunk.IsFinal {
			data.isFinal = true
			break
		}
	}

	return data, nil
}

// processUpload process upload fragment to its blobber
func (su *ChunkedUpload) processUpload(chunkStartIndex, chunkEndIndex int,
	fileShards []blobberShards, thumbnailShards blobberShards,
	isFinal bool, uploadLength int64) error {

	su.consensus.Reset()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	encryptedKey := ""
	if su.fileEncscheme != nil {
		encryptedKey = su.fileEncscheme.GetEncryptedKey()
	}

	var errCount int32

	wgErrors := make(chan error)
	wgDone := make(chan bool)
	if len(fileShards) == 0 {
		return thrown.New("upload_failed", "Upload failed. No data to upload")
	}
	wg := &sync.WaitGroup{}
	var pos uint64
	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		blobber := su.blobbers[pos]
		blobber.progress.UploadLength += uploadLength

		var thumbnailChunkData []byte

		if len(thumbnailShards) > 0 {
			thumbnailChunkData = thumbnailShards[pos]
		}

		body, formData, err := su.formBuilder.Build(
			&su.fileMeta, blobber.progress.Hasher, su.progress.ConnectionID,
			su.chunkSize, chunkStartIndex, chunkEndIndex, isFinal, encryptedKey,
			fileShards[pos], thumbnailChunkData,
		)

		if err != nil {
			return err
		}

		wg.Add(1)
		go func(b *ChunkedUploadBlobber, body *bytes.Buffer, formData ChunkedUploadFormMetadata, pos uint64) {
			defer wg.Done()

			startSendUploadRequest := time.Now()

			err = b.sendUploadRequest(ctx, su, chunkEndIndex, isFinal, encryptedKey, body, formData, pos)

			elapsedSendUploadRequest := time.Since(startSendUploadRequest)

			if err != nil {
				logger.Logger.Error("error - [sendUploadRequest] Timing", err, zap.String("elapsed time ", fmt.Sprintf("%d ms", elapsedSendUploadRequest.Milliseconds())))
				errC := atomic.AddInt32(&errCount, 1)
				if errC > int32(su.allocationObj.ParityShards-1) { // If atleast data shards + 1 number of blobbers can process the upload, it can be repaired later
					wgErrors <- err
				}

			} else {
				logger.Logger.Info("success - [sendUploadRequest] Timing", zap.String("elapsed time ", fmt.Sprintf("%d ms", elapsedSendUploadRequest.Milliseconds())))
			}

		}(blobber, body, formData, pos)
	}

	go func() {
		wg.Wait()
		close(wgDone)
	}()

	select {
	case <-wgDone:
		break
	case err := <-wgErrors:
		return thrown.New("upload_failed", fmt.Sprintf("Upload failed. %s", err))
	}

	if !su.consensus.isConsensusOk() {
		return thrown.New("consensus_not_met", fmt.Sprintf("Upload failed File not found for path %s. Required consensus atleast %d, got %d",
			su.fileMeta.RemotePath, su.consensus.consensusThresh, su.consensus.getConsensus()))
	}

	return nil
}

// processCommit commit shard upload on its blobber
func (su *ChunkedUpload) processCommit() error {
	defer su.removeProgress()

	logger.Logger.Info("Submitting for commit")
	su.consensus.Reset()

	wg := &sync.WaitGroup{}
	var pos uint64
	uid := util.GetNewUUID()
	timestamp := common.Now()
	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		blobber := su.blobbers[pos]

		//fixed numBlocks
		blobber.fileRef.ChunkSize = su.chunkSize
		blobber.fileRef.NumBlocks = int64(su.progress.ChunkIndex + 1)

		blobber.commitChanges = append(blobber.commitChanges,
			su.buildChange(blobber.fileRef, uid, timestamp))

		wg.Add(1)
		go func(b *ChunkedUploadBlobber, pos uint64) {
			defer wg.Done()
			err := b.processCommit(context.TODO(), su, pos, int64(timestamp))
			if err != nil {
				b.commitResult = ErrorCommitResult(err.Error())
			}

		}(blobber, pos)
	}

	wg.Wait()

	if !su.consensus.isConsensusOk() {
		consensus := su.consensus.getConsensus()
		err := thrown.New("consensus_not_met",
			fmt.Sprintf("Upload commit failed. Required consensus atleast %d, got %d",
				su.consensus.consensusThresh, consensus))

		if su.statusCallback != nil {
			su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, err)
		}
		return err
	}

	if su.statusCallback != nil {
		su.statusCallback.Completed(su.allocationObj.ID, su.fileMeta.RemotePath, su.fileMeta.RemoteName, su.fileMeta.MimeType, int(su.progress.UploadLength), su.opCode)
	}

	return nil
}

// getShardSize will return the size of data of a file each blobber is getting.
func getShardSize(dataSize int64, dataShards int, isEncrypted bool) int64 {
	chunkSize := int64(DefaultChunkSize)
	if isEncrypted {
		chunkSize -= (EncryptedDataPaddingSize + EncryptionHeaderSize)
	}

	totalChunkSize := chunkSize * int64(dataShards)

	n := dataSize / totalChunkSize
	r := dataSize % totalChunkSize

	var remainderShards int64
	if isEncrypted {
		remainderShards = (r+int64(dataShards)-1)/int64(dataShards) + EncryptedDataPaddingSize + EncryptionHeaderSize
	} else {
		remainderShards = (r + int64(dataShards) - 1) / int64(dataShards)
	}
	return n*DefaultChunkSize + remainderShards
}
