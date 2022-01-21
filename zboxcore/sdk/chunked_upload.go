package sdk

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	coreEncryption "github.com/0chain/gosdk/core/encryption"
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
)

// DefaultChunkSize default chunk size for file and thumbnail
const DefaultChunkSize = 64 * 1024

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
	if isRepair {
		found, repairRequired, _, err := allocationObj.RepairRequired(fileMeta.RemotePath)
		if err != nil {
			return nil, err
		}

		if !repairRequired {
			return nil, thrown.New("chunk_upload", "Repair not required")
		}

		uploadMask = found.Not().And(uploadMask)
	}

	su := &ChunkedUpload{
		// consensus: Consensus{
		// 	fullconsensus:          allocationObj.fullconsensus,
		// 	consensusThresh:        allocationObj.consensusThreshold,
		// 	consensusRequiredForOk: allocationObj.consensusOK,
		// },
		allocationObj: allocationObj,
		client:        zboxutil.Client,

		fileMeta:   fileMeta,
		fileReader: fileReader,

		uploadMask:      uploadMask,
		chunkSize:       DefaultChunkSize,
		encryptOnUpload: false,
	}

	if isUpdate {
		su.httpMethod = http.MethodPut
		su.buildChange = func(ref *fileref.FileRef) allocationchange.AllocationChange {
			change := &allocationchange.UpdateFileChange{}
			change.NewFile = ref
			change.NumBlocks = ref.NumBlocks
			change.Operation = constants.FileOperationUpdate
			change.Size = ref.Size
			change.NewFile.Attributes = ref.Attributes

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
			change.File.Attributes = ref.Attributes
			return change
		}
	}

	su.workdir = filepath.Join(workdir, ".zcn")

	//create upload folder to save progress
	err := FS.MkdirAll(filepath.Join(su.workdir, "upload"), 0744)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(su)
	}

	if su.createWriteMarkerLocker == nil {
		su.createWriteMarkerLocker = createWriteMarkerLocker
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

	}

	blobbers := su.allocationObj.Blobbers
	if len(blobbers) == 0 {
		return nil, thrown.New("no_blobbers", "Unable to find blobbers")
	}

	su.blobbers = make([]*ChunkedUploadBlobber, len(blobbers))

	for i := 0; i < len(blobbers); i++ {

		su.blobbers[i] = &ChunkedUploadBlobber{
			WriteMarkerLocker: su.createWriteMarkerLocker(filepath.Join(su.workdir, "blobber."+su.allocationObj.Blobbers[i].ID+".lock")),
			progress:          su.progress.Blobbers[i],
			blobber:           su.allocationObj.Blobbers[i],
			fileRef: &fileref.FileRef{
				Ref: fileref.Ref{
					Name:         su.fileMeta.RemoteName,
					Path:         su.fileMeta.RemotePath,
					Type:         fileref.FILE,
					AllocationID: su.allocationObj.ID,
				},
				Attributes: su.fileMeta.Attributes,
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

	allocationObj *Allocation

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

	// shardUploadedSize how much bytes a shard has. it is original size
	shardUploadedSize int64
	// shardUploadedThumbnailSize how much thumbnail bytes a shard has. it is original size
	shardUploadedThumbnailSize int64

	// statusCallback trigger progress on StatusCallback
	statusCallback StatusCallback

	blobbers                []*ChunkedUploadBlobber
	createWriteMarkerLocker func(file string) WriteMarkerLocker

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
	progressID := su.progressID()

	progress := su.progressStorer.Load(progressID)

	if progress != nil {
		su.progress = *progress
		su.progress.ID = progressID
	}

}

// saveProgress save progress to ~/.zcn/upload/[progressID]
func (su *ChunkedUpload) saveProgress() {
	su.progressStorer.Save(&su.progress)
}

// removeProgress remove progress info once it is done
func (su *ChunkedUpload) removeProgress() {
	su.progressStorer.Remove(su.progress.ID)
}

// createUploadProgress create a new UploadProgress
func (su *ChunkedUpload) createUploadProgress() UploadProgress {
	progress := UploadProgress{ConnectionID: zboxutil.NewConnectionId(),
		ChunkIndex:   0,
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
		chunk, err := su.chunkReader.Next()
		if err != nil {
			return err
		}
		logger.Logger.Info("Read chunk #", chunk.Index)

		su.shardUploadedSize += chunk.FragmentSize
		su.progress.UploadLength += chunk.ReadSize

		if chunk.Index == 0 && len(su.thumbnailBytes) > 0 {
			su.progress.UploadLength += int64(su.fileMeta.ActualThumbnailSize)
		}

		//skip chunk if it has been uploaded
		if chunk.Index < su.progress.ChunkIndex {
			continue
		}

		if chunk.IsFinal {
			su.fileMeta.ActualHash, err = su.fileHasher.GetFileHash()
			if err != nil {
				return err
			}

			if su.fileMeta.ActualSize == 0 {
				su.fileMeta.ActualSize = su.progress.UploadLength
			}
		}

		var thumbnailFragments [][]byte

		// upload entire thumbnail in first request only
		if chunk.Index == 0 && len(su.thumbnailBytes) > 0 {

			thumbnailFragments, err = su.chunkReader.Read(su.thumbnailBytes)
			if err != nil {
				return err
			}

		}

		err = su.processUpload(chunk.Index, chunk.Fragments, thumbnailFragments, chunk.IsFinal, chunk.ReadSize)
		if err != nil {
			return err
		}

		// last chunk might 0 with io.EOF
		// https://stackoverflow.com/questions/41208359/how-to-test-eof-on-io-reader-in-go
		if chunk.ReadSize > 0 {
			su.progress.ChunkIndex = chunk.Index
			su.saveProgress()

			if su.statusCallback != nil {
				su.statusCallback.InProgress(su.allocationObj.ID, su.fileMeta.RemotePath, OpUpload, int(su.progress.UploadLength), nil)
			}
		}

		if chunk.IsFinal {
			break
		}
	}

	if su.consensus.isConsensusOk() {
		logger.Logger.Info("Completed the upload. Submitting for commit")
		return su.processCommit()
	}

	err := fmt.Errorf("Upload failed: Consensus_rate:%f, expected:%f", su.consensus.getConsensusRate(), su.consensus.getConsensusRequiredForOk())
	if su.statusCallback != nil {
		su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.Path, OpUpload, err)
	}

	return err

}

//processUpload process upload fragment to its blobber
func (su *ChunkedUpload) processUpload(chunkIndex int, fileFragments [][]byte, thumbnailFragments [][]byte, isFinal bool, uploadLength int64) error {

	num := len(su.blobbers)
	if su.isRepair {
		num = len(su.blobbers) - su.uploadMask.TrailingZeros()
	} else {
		consensus := su.allocationObj.DataShards + su.allocationObj.ParityShards

		if num != consensus {
			return thrown.Throw(constants.ErrInvalidParameter, "len(su.blobbers) requires "+strconv.Itoa(num)+", not "+strconv.Itoa(len(su.blobbers)))
		}
	}

	wait := make(chan error, num)
	defer close(wait)

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	encryptedKey := ""
	if su.fileEncscheme != nil {
		encryptedKey = su.fileEncscheme.GetEncryptedKey()
	}

	var pos uint64
	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		blobber := su.blobbers[pos]
		blobber.progress.UploadLength += uploadLength

		var thumbnailBytes []byte
		var fileBytes []byte

		if len(thumbnailFragments) > 0 {
			thumbnailBytes = thumbnailFragments[pos]
		}

		if len(fileFragments) > 0 {
			fileBytes = fileFragments[pos]
		}

		body, formData, err := su.formBuilder.Build(&su.fileMeta, blobber.progress.Hasher, su.progress.ConnectionID, su.chunkSize, chunkIndex, isFinal, encryptedKey, fileBytes, thumbnailBytes)

		if err != nil {
			return err
		}

		go func(b *ChunkedUploadBlobber, buf *bytes.Buffer, form ChunkedUploadFormMetadata) {
			err := b.sendUploadRequest(ctx, su, chunkIndex, isFinal, encryptedKey, buf, form)

			// channel is not closed
			if ctx.Err() == nil {
				wait <- err
			}
		}(blobber, body, formData)
	}
	var err error
	for i := 0; i < num; i++ {
		err = <-wait
		if err != nil {
			return err
		}
	}

	return nil

}

// processCommit commit shard upload on its blobber
func (su *ChunkedUpload) processCommit() error {
	logger.Logger.Info("Submitting for commit")
	su.consensus.Reset()

	num := su.allocationObj.DataShards + su.allocationObj.ParityShards
	if su.isRepair {
		num = num - su.uploadMask.TrailingZeros()
	}

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

			// channel is not closed
			if ctx.Err() == nil {
				wait <- err
			}

		}(blobber)
	}

	var err error
	for i := 0; i < num; i++ {
		err = <-wait

		if err != nil {
			logger.Logger.Error("Commit: ", err)
			break
		}
	}

	if !su.consensus.isConsensusOk() {
		if su.consensus.getConsensus() != 0 {
			logger.Logger.Info("Commit consensus failed, Deleting remote file....")
			su.allocationObj.deleteFile(su.fileMeta.RemotePath, su.consensus.getConsensus(), su.consensus.getConsensus()) //nolint
		}
		if su.statusCallback != nil {
			su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, OpUpload, thrown.New("commit_consensus_failed", "Upload failed as there was no commit consensus"))
			return nil
		}
	}

	su.removeProgress()

	if su.statusCallback != nil {
		su.statusCallback.Completed(su.allocationObj.ID, su.fileMeta.RemotePath, su.fileMeta.RemoteName, su.fileMeta.MimeType, int(su.progress.UploadLength), OpUpload)
	}

	return nil
}
