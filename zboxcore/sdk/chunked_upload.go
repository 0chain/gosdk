package sdk

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
)

const (
	DefaultUploadTimeOut = 180 * time.Second
)

var (
	CmdFFmpeg = "ffmpeg"
	// DefaultHashFunc default hash method for stream merkle tree
	DefaultHashFunc = func(left, right string) string {
		return coreEncryption.Hash(left + right)
	}

	ErrInvalidChunkSize              = errors.New("chunk: chunk size is too small. it must greater than 272 if file is uploaded with encryption")
	ErrNoEnoughSpaceLeftInAllocation = errors.New("alloc: no enough space left in allocation")
	CancelOpCtx                      = make(map[string]context.CancelCauseFunc)
	cancelLock                       sync.Mutex
	CurrentMode                      = UploadModeMedium
	shouldSaveProgress               = true
	HighModeWorkers                  = 4
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

type UploadMode byte

const (
	UploadModeLow UploadMode = iota
	UploadModeMedium
	UploadModeHigh
)

func SetUploadMode(mode UploadMode) {
	CurrentMode = mode
}

func SetHighModeWorkers(workers int) {
	HighModeWorkers = workers
}

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
	ctx context.Context,
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
		newFileReader, newFileMeta, f, err := TranscodeWebStreaming(workdir, fileReader, fileMeta)
		defer os.Remove(f)

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

	if isUpdate {
		opCode = OpUpdate
	}

	consensus := Consensus{
		RWMutex:         &sync.RWMutex{},
		consensusThresh: allocationObj.consensusThreshold,
		fullconsensus:   allocationObj.fullconsensus,
	}

	uploadMask := zboxutil.NewUint128(1).Lsh(uint64(len(allocationObj.Blobbers))).Sub64(1)

	su := &ChunkedUpload{
		allocationObj: allocationObj,
		client:        zboxutil.Client,
		fileMeta:      fileMeta,
		fileReader:    fileReader,

		uploadMask:      uploadMask,
		chunkSize:       DefaultChunkSize,
		chunkNumber:     100,
		encryptOnUpload: false,
		webStreaming:    false,

		consensus:     consensus, //nolint
		uploadTimeOut: DefaultUploadTimeOut,
		commitTimeOut: DefaultUploadTimeOut,
		maskMu:        &sync.Mutex{},
		opCode:        opCode,
	}

	// su.ctx, su.ctxCncl = context.WithCancel(allocationObj.ctx)
	su.ctx, su.ctxCncl = context.WithCancelCause(ctx)
	su.httpMethod = http.MethodPost

	if isUpdate {
		su.buildChange = func(ref *fileref.FileRef, _ uuid.UUID, ts common.Timestamp) allocationchange.AllocationChange {
			change := &allocationchange.UpdateFileChange{}
			change.NewFile = ref
			change.NumBlocks = ref.NumBlocks
			change.Operation = constants.FileOperationUpdate
			change.Size = ref.Size
			return change
		}
	} else {
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
	err = sys.Files.MkdirAll(filepath.Join(su.workdir, "upload"), 0766)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(su)
	}

	if isRepair {
		opCode = OpUpdate
		su.consensus.fullconsensus = su.uploadMask.CountOnes()
		su.consensus.consensusThresh = su.uploadMask.CountOnes()
	}

	if su.progressStorer == nil && shouldSaveProgress {
		su.progressStorer = createFsChunkedUploadProgress(context.Background())
	}

	su.loadProgress()
	su.shardSize = getShardSize(su.fileMeta.ActualSize, su.allocationObj.DataShards, su.encryptOnUpload)
	if su.fileHasher == nil {
		su.fileHasher = CreateFileHasher()
	}

	// encrypt option has been changed. upload it from scratch
	// chunkSize has been changed. upload it from scratch
	// actual size has been changed. upload it from scratch
	if su.progress.ChunkSize != su.chunkSize || su.progress.EncryptOnUpload != su.encryptOnUpload || su.progress.ActualSize != su.fileMeta.ActualSize || su.progress.ChunkNumber != su.chunkNumber || su.progress.ConnectionID == "" {
		su.progress.ChunkSize = 0 // reset chunk size
	}

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
		if su.fileEncscheme == nil {
			return nil, thrown.New("upload_failed", "Failed to create encryption scheme")
		}
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
	cReader, err := createChunkReader(su.fileReader, fileMeta.ActualSize, int64(su.chunkSize), su.allocationObj.DataShards, su.allocationObj.ParityShards, su.encryptOnUpload, su.uploadMask, su.fileErasureEncoder, su.fileEncscheme, su.fileHasher, su.chunkNumber)

	if err != nil {
		return nil, err
	}

	su.chunkReader = cReader

	su.formBuilder = CreateChunkedUploadFormBuilder()

	su.isRepair = isRepair
	uploadWorker, uploadRequest := calculateWorkersAndRequests(su.allocationObj.DataShards, len(su.blobbers), su.chunkNumber)
	su.uploadChan = make(chan UploadData, uploadRequest)
	su.uploadWorkers = uploadWorker
	return su, nil
}

func calculateWorkersAndRequests(dataShards, totalShards, chunknumber int) (uploadWorkers int, uploadRequests int) {
	if totalShards < 4 {
		uploadWorkers = 4
	} else {
		switch CurrentMode {
		case UploadModeLow:
			uploadWorkers = 1
		case UploadModeMedium:
			uploadWorkers = 2
		case UploadModeHigh:
			uploadWorkers = HighModeWorkers
		}
	}

	if chunknumber*dataShards < 640 && !IsWasm {
		uploadRequests = 4
	} else {
		uploadRequests = 2
	}
	return
}

// progressID build local progress id with [allocationid]_[Hash(LocalPath+"_"+RemotePath)]_[RemoteName] format
func (su *ChunkedUpload) progressID() string {

	if len(su.allocationObj.ID) > 8 {
		return filepath.Join(su.workdir, "upload", "u"+su.allocationObj.ID[:8]+"_"+su.fileMeta.FileID())
	}

	return filepath.Join(su.workdir, "upload", su.allocationObj.ID+"_"+su.fileMeta.FileID())
}

// loadProgress load progress from ~/.zcn/upload/[progressID]
func (su *ChunkedUpload) loadProgress() {
	// ChunkIndex starts with 0, so default value should be -1
	su.progress.ChunkIndex = -1

	progressID := su.progressID()
	if shouldSaveProgress {
		progress := su.progressStorer.Load(progressID)

		if progress != nil {
			su.progress = *progress
			su.progress.ID = progressID
		}
	}
}

// saveProgress save progress to ~/.zcn/upload/[progressID]
func (su *ChunkedUpload) saveProgress() {
	if su.progressStorer != nil {
		su.progressStorer.Save(su.progress)
	}
}

// removeProgress remove progress info once it is done
func (su *ChunkedUpload) removeProgress() {
	if su.progressStorer != nil {
		su.progressStorer.Remove(su.progress.ID) //nolint
	}
}

func (su *ChunkedUpload) updateProgress(chunkIndex int, upMask zboxutil.Uint128) {
	if su.progressStorer != nil {
		if chunkIndex > su.progress.ChunkIndex {
			su.progressStorer.Update(su.progress.ID, chunkIndex, upMask)
		}
	}
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
		mnemonic := client.GetClient().Mnemonic
		if mnemonic == "" {
			return nil
		}
		privateKey, err := encscheme.Initialize(mnemonic)
		if err != nil {
			return nil
		}

		su.progress.EncryptPrivateKey = hex.EncodeToString(privateKey)
	}
	if len(su.progress.EncryptedKeyPoint) > 0 {
		err := encscheme.InitForEncryptionWithPoint("filetype:audio", su.progress.EncryptedKeyPoint)
		if err != nil {
			return nil
		}
	} else {
		encscheme.InitForEncryption("filetype:audio")
		su.progress.EncryptedKeyPoint = encscheme.GetEncryptedKeyPoint()
	}
	su.encryptedKey = encscheme.GetEncryptedKey()
	return encscheme
}

func (su *ChunkedUpload) process() error {
	if su.statusCallback != nil {
		su.statusCallback.Started(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, int(su.fileMeta.ActualSize)+int(su.fileMeta.ActualThumbnailSize))
	}
	su.startProcessor()
	defer su.chunkReader.Close()
	defer su.ctxCncl(nil)
	for {

		chunks, err := su.readChunks(su.chunkNumber)

		// chunk, err := su.chunkReader.Next()
		if err != nil {
			if su.statusCallback != nil {
				su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, err)
			}
			return err
		}
		//logger.Logger.Debug("Read chunk #", chunk.Index)

		su.shardUploadedSize += chunks.totalFragmentSize
		su.progress.ReadLength += chunks.totalReadSize

		if chunks.isFinal {
			if su.fileMeta.ActualHash == "" {
				su.fileMeta.ActualHash, err = su.chunkReader.GetFileHash()
				if err != nil {
					if su.statusCallback != nil {
						su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, err)
					}
					return err
				}
			}
			if su.fileMeta.ActualSize == 0 {
				su.fileMeta.ActualSize = su.progress.ReadLength
				su.shardSize = getShardSize(su.fileMeta.ActualSize, su.allocationObj.DataShards, su.encryptOnUpload)
			} else if su.fileMeta.ActualSize != su.progress.ReadLength && su.thumbnailBytes == nil {
				if su.statusCallback != nil {
					su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, thrown.New("upload_failed", "Upload failed. Uploaded size does not match with actual size: "+fmt.Sprintf("%d != %d", su.fileMeta.ActualSize, su.progress.ReadLength)))
				}
				return thrown.New("upload_failed", "Upload failed. Uploaded size does not match with actual size: "+fmt.Sprintf("%d != %d", su.fileMeta.ActualSize, su.progress.ReadLength))
			}
		}

		err = su.processUpload(
			chunks.chunkStartIndex, chunks.chunkEndIndex,
			chunks.fileShards, chunks.thumbnailShards,
			chunks.isFinal, chunks.totalReadSize,
		)
		if err != nil {
			if su.statusCallback != nil {
				su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, err)
			}
			return err
		}

		// last chunk might 0 with io.EOF
		// https://stackoverflow.com/questions/41208359/how-to-test-eof-on-io-reader-in-go
		if chunks.isFinal {
			break
		}
	}
	return nil
}

// Start start/resume upload
func (su *ChunkedUpload) Start() error {
	now := time.Now()

	err := su.process()
	if err != nil {
		return err
	}
	su.ctx, su.ctxCncl = context.WithCancelCause(su.allocationObj.ctx)
	defer su.ctxCncl(nil)
	elapsedProcess := time.Since(now)

	blobbers := make([]*blockchain.StorageNode, len(su.blobbers))
	for i, b := range su.blobbers {
		blobbers[i] = b.blobber
	}
	if su.addConsensus == int32(su.consensus.fullconsensus) {
		return thrown.New("upload_failed", "Duplicate upload detected")
	}

	err = su.writeMarkerMutex.Lock(
		su.ctx, &su.uploadMask, su.maskMu,
		blobbers, &su.consensus, int(su.addConsensus), su.uploadTimeOut,
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

	defer func() {
		elapsedProcessCommit := time.Since(now) - elapsedProcess - elapsedLock
		logger.Logger.Info("[ChunkedUpload - start] Timings:\n",
			fmt.Sprintf("allocation_id: %s", su.allocationObj.ID),
			fmt.Sprintf("process: %d ms", elapsedProcess.Milliseconds()),
			fmt.Sprintf("Lock: %d ms", elapsedLock.Milliseconds()),
			fmt.Sprintf("processCommit: %d ms", elapsedProcessCommit.Milliseconds()))
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

			data.thumbnailShards, err = su.chunkReader.Read(su.thumbnailBytes)
			if err != nil {
				return nil, err
			}
		}

		if data.fileShards == nil {
			data.fileShards = make([]blobberShards, len(chunk.Fragments))
		}

		// concact blobber's fragments
		if chunk.ReadSize > 0 {
			for i, v := range chunk.Fragments {
				//blobber i
				data.fileShards[i] = append(data.fileShards[i], v)
			}
		}

		if chunk.IsFinal {
			data.isFinal = true
			break
		}
	}
	su.chunkReader.Reset()
	return data, nil
}

// processCommit commit shard upload on its blobber
func (su *ChunkedUpload) processCommit() error {
	defer su.removeProgress()

	logger.Logger.Info("Submitting for commit")
	su.consensus.Reset()
	su.consensus.consensus = int(su.addConsensus)
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
	if dataSize == 0 {
		return 0
	}
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

func (su *ChunkedUpload) uploadProcessor() {
	for {
		select {
		case <-su.ctx.Done():
			return
		case uploadData, ok := <-su.uploadChan:
			if !ok {
				return
			}
			su.uploadToBlobbers(uploadData) //nolint:errcheck
			su.uploadWG.Done()
		}
	}
}

func (su *ChunkedUpload) uploadToBlobbers(uploadData UploadData) error {
	select {
	case <-su.ctx.Done():
		return context.Cause(su.ctx)
	default:
	}
	consensus := Consensus{
		RWMutex:         &sync.RWMutex{},
		consensusThresh: su.consensus.consensusThresh,
		fullconsensus:   su.consensus.fullconsensus,
	}

	wgErrors := make(chan error, len(su.blobbers))
	ctx, cancel := context.WithCancel(su.ctx)
	defer cancel()
	var pos uint64
	var errCount int32
	var wg sync.WaitGroup
	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		wg.Add(1)
		go func(pos uint64) {
			defer wg.Done()
			err := su.blobbers[pos].sendUploadRequest(ctx, su, uploadData.isFinal, su.encryptedKey, uploadData.uploadBody[pos].dataBuffers, uploadData.uploadBody[pos].formData, uploadData.uploadBody[pos].contentSlice, pos, &consensus)

			if err != nil {
				if strings.Contains(err.Error(), "duplicate") {
					su.consensus.Done()
					errC := atomic.AddInt32(&su.addConsensus, 1)
					if errC >= int32(su.consensus.consensusThresh) {
						wgErrors <- err
					}
					return
				}
				logger.Logger.Error("error during sendUploadRequest", err, " connectionID: ", su.progress.ConnectionID)
				errC := atomic.AddInt32(&errCount, 1)
				if errC > int32(su.allocationObj.ParityShards-1) { // If atleast data shards + 1 number of blobbers can process the upload, it can be repaired later
					wgErrors <- err
				}
			}
		}(pos)
	}
	wg.Wait()
	close(wgErrors)
	for err := range wgErrors {
		su.ctxCncl(thrown.New("upload_failed", fmt.Sprintf("Upload failed. %s", err)))
		return err
	}
	if !consensus.isConsensusOk() {
		err := thrown.New("consensus_not_met", fmt.Sprintf("Upload failed File not found for path %s. Required consensus atleast %d, got %d",
			su.fileMeta.RemotePath, consensus.consensusThresh, consensus.getConsensus()))
		su.ctxCncl(err)
		return err
	}
	if uploadData.uploadLength > 0 {
		index := uploadData.chunkEndIndex
		uploadLength := uploadData.uploadLength
		go su.updateProgress(index, su.uploadMask)
		if su.statusCallback != nil {
			su.statusCallback.InProgress(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, int(atomic.AddInt64(&su.progress.UploadLength, uploadLength)), nil)
		}
	}
	uploadData = UploadData{} // release memory
	return nil
}
