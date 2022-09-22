package sdk

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"errors"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	coreEncryption "github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/klauspost/reedsolomon"
)

var (
	// DefaultHashFunc default hash method for stream merkle tree
	DefaultHashFunc = func(left, right string) string {
		return coreEncryption.Hash(left + right)
	}

	ErrInvalidChunkSize = errors.New("chunk: chunk size is too small. it must greater than 272 if file is uploaded with encryption")

	ErrCommitConsensusFailed = thrown.New("commit_consensus_failed", "Upload failed as there was no commit consensus")
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
func CreateChunkedUpload(workdir string, allocationObj *Allocation, fileMeta FileMeta, fileReader io.Reader, isUpdate, isRepair bool, opts ...ChunkedUploadOption) (*ChunkedUpload, error) {
	if allocationObj == nil {
		return nil, thrown.Throw(constants.ErrInvalidParameter, "allocationObj")
	}

	var uploadMask zboxutil.Uint128 = zboxutil.NewUint128(1).Lsh(uint64(len(allocationObj.Blobbers))).Sub64(1)
	fullConsensus, threshConsensus, consensusOK := allocationObj.getConsensuses()
	if isRepair {
		found, repairRequired, _, err := allocationObj.RepairRequired(fileMeta.RemotePath)
		if err != nil {
			return nil, err
		}

		if !repairRequired {
			return nil, thrown.New("chunk_upload", "Repair not required")
		}

		uploadMask = found.Not().And(uploadMask)

		fullConsensus = float32(uploadMask.CountOnes())
		threshConsensus = 100
		consensusOK = 100
	}

	su := &ChunkedUpload{
		allocationObj: allocationObj,
		client:        zboxutil.Client,

		fileMeta:   fileMeta,
		fileReader: fileReader,

		uploadMask:      uploadMask,
		chunkSize:       DefaultChunkSize,
		chunkNumber:     1,
		encryptOnUpload: false,
	}

	su.consensus.Init(threshConsensus, fullConsensus, consensusOK)

	if isUpdate {
		su.httpMethod = http.MethodPut
		su.buildChange = func(ref *fileref.FileRef) allocationchange.AllocationChange {
			change := &allocationchange.UpdateFileChange{}
			change.NewFile = ref
			change.NumBlocks = ref.NumBlocks
			change.Operation = constants.FileOperationUpdate
			change.Size = ref.Size

			return change
		}
	} else {
		su.httpMethod = http.MethodPost
		su.buildChange = func(ref *fileref.FileRef) allocationchange.AllocationChange {
			change := &allocationchange.NewFileChange{}
			change.File = ref
			change.NumBlocks = ref.NumBlocks
			change.Operation = constants.FileOperationInsert
			change.Size = ref.Size
			return change
		}
	}

	su.workdir = filepath.Join(workdir, ".zcn")

	//create upload folder to save progress
	err := sys.Files.MkdirAll(filepath.Join(su.workdir, "upload"), 0744)
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

	su.fileHasher = CreateHasher(int(su.chunkSize))

	// encrypt option has been chaned.upload it from scratch
	// chunkSize has been changed. upload it from scratch
	if su.progress.EncryptOnUpload != su.encryptOnUpload || su.progress.ChunkSize != su.chunkSize {
		su.progress = su.createUploadProgress()
	}

	su.fileErasureEncoder, _ = reedsolomon.New(su.allocationObj.DataShards, su.allocationObj.ParityShards, reedsolomon.WithAutoGoroutines(int(su.chunkSize)))

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
	client         zboxutil.HttpClient

	uploadMask zboxutil.Uint128

	// httpMethod POST = Upload File / PUT = Update file
	httpMethod  string
	buildChange func(ref *fileref.FileRef) allocationchange.AllocationChange

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
func (su *ChunkedUpload) createUploadProgress() UploadProgress {
	progress := UploadProgress{ConnectionID: zboxutil.NewConnectionId(),
		ChunkIndex:   -1,
		ChunkSize:    su.chunkSize,
		UploadLength: 0,
		Blobbers:     make([]*UploadBlobberStatus, su.allocationObj.DataShards+su.allocationObj.ParityShards),
	}

	for i := 0; i < len(progress.Blobbers); i++ {
		progress.Blobbers[i] = &UploadBlobberStatus{
			Hasher: CreateHasher(int(su.chunkSize)),
		}
	}

	progress.ID = su.progressID()
	return progress
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

// Start start/resume upload
func (su *ChunkedUpload) Start() error {

	if su.statusCallback != nil {
		su.statusCallback.Started(su.allocationObj.ID, su.fileMeta.RemotePath, OpUpload, int(su.fileMeta.ActualSize)+int(su.fileMeta.ActualThumbnailSize))
	}

	for {

		chunks, err := su.readChunks(su.chunkNumber)

		// chunk, err := su.chunkReader.Next()
		if err != nil {
			if su.statusCallback != nil {
				su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.Path, OpUpload, err)
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
					su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.Path, OpUpload, err)
				}
				return err
			}

			if su.fileMeta.ActualSize == 0 {
				su.fileMeta.ActualSize = su.progress.UploadLength
			}
		}

		//chunk has not be uploaded yet
		if chunks.chunkEndIndex > su.progress.ChunkIndex {

			err = su.processUpload(chunks.chunkStartIndex, chunks.chunkEndIndex, chunks.fileShards, chunks.thumbnailShards, chunks.isFinal, chunks.totalReadSize)
			if err != nil {
				if su.statusCallback != nil {
					su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.Path, OpUpload, err)
				}
				return err
			}
		}

		// last chunk might 0 with io.EOF
		// https://stackoverflow.com/questions/41208359/how-to-test-eof-on-io-reader-in-go
		if chunks.totalReadSize > 0 {
			su.progress.ChunkIndex = chunks.chunkEndIndex
			su.saveProgress()

			if su.statusCallback != nil {
				su.statusCallback.InProgress(su.allocationObj.ID, su.fileMeta.RemotePath, OpUpload, int(su.progress.UploadLength), nil)
			}
		}

		if chunks.isFinal {
			break
		}
	}

	logger.Logger.Info("Completed the upload. Submitting for commit")

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

		// concat blobber's fragments
		for i, v := range chunk.Fragments {
			//blobber i
			data.fileShards[i] = append(data.fileShards[i], v)
		}

		if chunk.IsFinal {
			data.isFinal = true
			break
		}
	}

	return data, nil
}

//processUpload process upload fragment to its blobber
func (su *ChunkedUpload) processUpload(chunkStartIndex, chunkEndIndex int, fileShards []blobberShards, thumbnailShards blobberShards, isFinal bool, uploadLength int64) error {

	num := su.uploadMask.CountOnes()
	wait := make(chan UploadError, num)
	defer close(wait)

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	su.consensus.Reset()

	encryptedKey := ""
	if su.fileEncscheme != nil {
		encryptedKey = su.fileEncscheme.GetEncryptedKey()
	}

	var pos uint64
	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		blobber := su.blobbers[pos]
		blobber.progress.UploadLength += uploadLength

		var thumbnailChunkData []byte

		if len(thumbnailShards) > 0 {
			thumbnailChunkData = thumbnailShards[pos]
		}

		body, formData, err := su.formBuilder.Build(&su.fileMeta, blobber.progress.Hasher, su.progress.ConnectionID, su.chunkSize, chunkStartIndex, chunkEndIndex, isFinal, encryptedKey, fileShards[pos], thumbnailChunkData)

		if err != nil {
			return err
		}

		go func(idx uint64, b *ChunkedUploadBlobber, buf *bytes.Buffer, form ChunkedUploadFormMetadata) {
			err := b.sendUploadRequest(ctx, su, chunkEndIndex, isFinal, encryptedKey, buf, form)

			util.WithRecover(func() {
				wait <- UploadError{
					Error:      err,
					BlobberIdx: pos,
				}
			})

		}(pos, blobber, body, formData)
	}

	var opError error
	for i := 0; i < num; i++ {
		err, ok := <-wait
		//channel is closed
		if !ok {
			break
		}
		if err.Error != nil {
			logger.Logger.Error("Upload: ", err.Error)
			//stop to upload new chunks to failed blobber
			su.uploadMask = su.uploadMask.Sub64(err.BlobberIdx)
			opError = err.Error
		}
	}

	if su.consensus.isConsensusOk() {
		return nil
	}

	// all of blobber are failed. it should be rejected by rule
	if su.consensus.getConsensus() == 0 {
		return opError
	}

	errConsensus := fmt.Errorf("Upload failed: Consensus_rate:%f, expected:%f", su.consensus.getConsensusRate(), su.consensus.getConsensusRequiredForOk())
	if su.statusCallback != nil {
		su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.Path, OpUpload, errConsensus)
	}

	return errConsensus

}

// processCommit commit shard upload on its blobber
func (su *ChunkedUpload) processCommit() error {
	shouldUnlock := true
	err := su.writeMarkerMutex.Lock(context.TODO(), su.progress.ConnectionID)
	defer func() {
		if shouldUnlock {
			su.writeMarkerMutex.Unlock(context.TODO(), su.progress.ConnectionID) //nolint: errcheck
		}
	}()

	if err != nil {
		if su.statusCallback != nil {
			su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.Path, OpUpload, err)
		}
		return err
	}

	defer su.removeProgress()

	logger.Logger.Info("Submitting for commit")
	su.consensus.Reset()

	num := su.uploadMask.CountOnes()

	wait := make(chan error, num)
	defer close(wait)

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	var pos uint64
	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		blobber := su.blobbers[pos]
		//fixed numBlocks
		blobber.fileRef.ChunkSize = su.chunkSize
		blobber.fileRef.NumBlocks = int64(su.progress.ChunkIndex + 1)

		blobber.commitChanges = append(blobber.commitChanges, su.buildChange(blobber.fileRef))

		go func(b *ChunkedUploadBlobber) {
			err := b.processCommit(ctx, su)

			if err != nil {
				b.commitResult = ErrorCommitResult(err.Error())
			}

			util.WithRecover(func() {
				wait <- err
			})

		}(blobber)
	}

	for i := 0; i < num; i++ {
		err, ok := <-wait
		if !ok {
			break
		}

		if err != nil {
			logger.Logger.Error("Commit: ", err)
		}
	}

	if !su.consensus.isConsensusOk() {
		if su.consensus.getConsensus() != 0 {
			logger.Logger.Info("Commit consensus failed, Deleting remote file....")
			su.writeMarkerMutex.Unlock(context.TODO(), su.progress.ConnectionID) //nolint: errcheck
			shouldUnlock = false
			su.allocationObj.deleteFile(su.fileMeta.RemotePath, su.consensus.getConsensus(), su.consensus.getConsensus()) //nolint
		}
		if su.statusCallback != nil {
			su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, OpUpload, ErrCommitConsensusFailed)
			return nil
		}
	}

	if su.statusCallback != nil {
		su.statusCallback.Completed(su.allocationObj.ID, su.fileMeta.RemotePath, su.fileMeta.RemoteName, su.fileMeta.MimeType, int(su.progress.UploadLength), OpUpload)
	}

	return nil
}

type UploadError struct {
	BlobberIdx uint64
	Error      error
}
