package sdk

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	coreEncryption "github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/gofrs/flock"
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

// CreateChunkedUpload create a ChunkedUpload instance
func CreateChunkedUpload(workdir string, allocationObj *Allocation, fileMeta FileMeta, fileReader io.Reader, isUpdate bool, opts ...ChunkedUploadOption) (*ChunkedUpload, error) {

	if allocationObj == nil {
		return nil, thrown.Throw(constants.ErrInvalidParameter, "allocationObj")
	}

	su := &ChunkedUpload{
		allocationObj: allocationObj,
		client: &http.Client{
			Transport: zboxutil.DefaultTransport,
		},

		fileMeta:   fileMeta,
		fileReader: fileReader,

		fileHasher: createHasher(),

		progressStorer: createFsChunkedUploadProgress(context.Background()),

		uploadMask:      zboxutil.NewUint128(1).Lsh(uint64(len(allocationObj.Blobbers))).Sub64(1),
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
	err := os.MkdirAll(filepath.Join(su.workdir, "upload"), 0744)
	if err != nil {
		return nil, err
	}

	su.loadProgress()

	for _, opt := range opts {
		opt(su)
	}

	// encrypt option has been chaned.upload it from scratch
	// chunkSize has been changed. upload it from scratch
	if su.progress.EncryptOnUpload != su.encryptOnUpload || su.progress.ChunkSize != su.chunkSize {
		su.progress = su.createUploadProgress()
	}

	su.fileErasureEncoder, _ = reedsolomon.New(su.allocationObj.DataShards, su.allocationObj.ParityShards, reedsolomon.WithAutoGoroutines(su.chunkSize))

	if su.encryptOnUpload {
		su.fileEncscheme = su.createEncscheme()

	}

	su.blobbers = make([]*ChunkedUploadBobbler, len(su.allocationObj.Blobbers))

	for i := 0; i < len(su.allocationObj.Blobbers); i++ {

		su.blobbers[i] = &ChunkedUploadBobbler{
			flock:    flock.New(filepath.Join(su.workdir, "blobber."+su.allocationObj.Blobbers[i].ID+".lock")),
			progress: su.progress.Blobbers[i],
			blobber:  su.allocationObj.Blobbers[i],
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

	cReader, err := createChunkReader(su.fileReader, int64(su.chunkSize), su.allocationObj.DataShards, su.encryptOnUpload, su.uploadMask, su.fileErasureEncoder, su.fileEncscheme, su.fileHasher)

	if err != nil {
		return nil, err
	}

	su.chunkReader = cReader

	return su, nil

}

// ChunkedUpload upload manager with chunked upload feature
type ChunkedUpload struct {
	Consensus

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

	chunkReader ChunkReader

	// encryptOnUpload encrypt data on upload or not.
	encryptOnUpload bool
	// chunkSize how much bytes a chunk has. 64KB is default value.
	chunkSize int

	// shardUploadedSize how much bytes a shard has. it is original size
	shardUploadedSize int64
	// shardUploadedThumbnailSize how much thumbnail bytes a shard has. it is original size
	shardUploadedThumbnailSize int64

	// statusCallback trigger progress on StatusCallback
	statusCallback StatusCallback

	blobbers []*ChunkedUploadBobbler
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
			ChallengeHasher: &util.FixedMerkleTree{ChunkSize: su.chunkSize},
			ContentHasher: util.NewCompactMerkleTree(func(left, right string) string {
				return coreEncryption.Hash(left + right)
			}),
		}
	}

	progress.ID = su.progressID()
	return progress
}

func (su *ChunkedUpload) createEncscheme() encryption.EncryptionScheme {
	encscheme := encryption.NewEncryptionScheme()

	if len(su.progress.EncryptPrivteKey) > 0 {

		privateKey, _ := hex.DecodeString(su.progress.EncryptPrivteKey)

		err := encscheme.InitializeWithPrivateKey(privateKey)
		if err != nil {
			return nil
		}
	} else {
		privateKey, err := encscheme.Initialize(client.GetClient().Mnemonic)
		if err != nil {
			return nil
		}

		su.progress.EncryptPrivteKey = hex.EncodeToString(privateKey)
	}

	encscheme.InitForEncryption("filetype:audio")

	return encscheme
}

// Start start/resume upload
func (su *ChunkedUpload) Start() error {

	if su.statusCallback != nil {
		su.statusCallback.Started(su.allocationObj.ID, su.fileMeta.RemotePath, OpUpload, int(su.fileMeta.ActualSize)+int(su.fileMeta.ActualThumbnailSize))
	}

	for i := 0; ; i++ {

		chunk, err := su.chunkReader.Next()

		if err != nil {
			return err
		}

		fileShards := chunk.Fragments

		readLen := chunk.Size
		chunkSize := chunk.FragmentSize
		isFinal := chunk.IsFinal

		su.shardUploadedSize += chunkSize
		su.progress.UploadLength += int64(readLen)

		if i == 0 && len(su.thumbnailBytes) > 0 {
			su.progress.UploadLength += int64(su.fileMeta.ActualThumbnailSize)
		}

		//skip chunk if it has been uploaded
		if i < su.progress.ChunkIndex {
			continue
		}

		if isFinal {
			su.fileMeta.ActualHash, err = su.fileHasher.GetFileHash()
			if err != nil {
				return err
			}

			if su.fileMeta.ActualSize == 0 {
				su.fileMeta.ActualSize = su.progress.UploadLength
			}
		}

		// upload entire thumbnail in first reqeust only
		if i == 0 && len(su.thumbnailBytes) > 0 {

			thumbnailShards, err := su.readThumbnailShards()
			if err != nil {
				return err
			}

			su.processUpload(i, fileShards, thumbnailShards, isFinal, readLen)

		} else {
			su.processUpload(i, fileShards, nil, isFinal, readLen)
		}

		// last chunk might 0 with io.EOF
		// https://stackoverflow.com/questions/41208359/how-to-test-eof-on-io-reader-in-go
		if readLen > 0 {
			su.progress.ChunkIndex = i
			su.saveProgress()

			if su.statusCallback != nil {
				su.statusCallback.InProgress(su.allocationObj.ID, su.fileMeta.RemotePath, OpUpload, int(su.progress.UploadLength), nil)
			}
		}

		if isFinal {
			break
		}
	}

	if su.isConsensusOk() {
		logger.Logger.Info("Completed the upload. Submitting for commit")
		return su.processCommit()
	}

	err := fmt.Errorf("Upload failed: Consensus_rate:%f, expected:%f", su.getConsensusRate(), su.getConsensusRequiredForOk())
	if su.statusCallback != nil {
		su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.Path, OpUpload, err)
	}

	return err

}

// readThumbnailShards encode and encrypt thumbnail
func (su *ChunkedUpload) readThumbnailShards() ([][]byte, error) {

	shards, err := su.thumbailErasureEncoder.Split(su.thumbnailBytes)
	if err != nil {
		logger.Logger.Error("[upload] Erasure coding on thumbnail failed:", err.Error())
		return nil, err
	}

	err = su.thumbailErasureEncoder.Encode(shards)
	if err != nil {
		logger.Logger.Error("[upload] Erasure coding on thumbnail failed:", err.Error())
		return nil, err
	}

	var c, pos uint64
	if su.encryptOnUpload {
		for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
			pos = uint64(i.TrailingZeros())
			encMsg, err := su.fileEncscheme.Encrypt(shards[pos])
			if err != nil {
				logger.Logger.Error("[upload] Encryption on thumbnail failed:", err.Error())
				return nil, err
			}
			header := make([]byte, 2*1024)
			copy(header[:], encMsg.MessageChecksum+","+encMsg.OverallChecksum)
			shards[pos] = append(header, encMsg.EncryptedData...)
			c++
		}

		//c, pos = 0, 0
	}

	return shards, nil
}

//processUpload process upload shard to its blobber
func (su *ChunkedUpload) processUpload(chunkIndex int, fileShards [][]byte, thumbnailShards [][]byte, isFinal bool, uploadLength int64) {
	threads := su.allocationObj.DataShards + su.allocationObj.ParityShards

	wg := &sync.WaitGroup{}
	wg.Add(threads)

	var pos uint64
	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		blobber := su.blobbers[pos]
		blobber.progress.UploadLength += uploadLength

		if len(thumbnailShards) > 0 {
			go blobber.processUpload(su, chunkIndex, fileShards[pos], thumbnailShards[pos], isFinal, wg)
		} else {
			go blobber.processUpload(su, chunkIndex, fileShards[pos], nil, isFinal, wg)
		}
	}

	wg.Wait()

}

// processCommit commit shard upload on its blobber
func (su *ChunkedUpload) processCommit() error {
	logger.Logger.Info("Submitting for commit")
	su.consensus = 0
	wg := &sync.WaitGroup{}
	ones := su.uploadMask.CountOnes()

	wg.Add(ones)

	var pos uint64
	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		blobber := su.blobbers[pos]

		//fixed numBlocks
		blobber.fileRef.ChunkSize = su.chunkSize
		blobber.fileRef.NumBlocks = int64(su.progress.ChunkIndex + 1)

		blobber.commitChanges = append(blobber.commitChanges, su.buildChange(blobber.fileRef))

		go blobber.processCommit(su, wg)
	}
	wg.Wait()

	if !su.isConsensusOk() {
		if su.consensus != 0 {
			logger.Logger.Info("Commit consensus failed, Deleting remote file....")
			su.allocationObj.deleteFile(su.fileMeta.RemotePath, su.consensus, su.consensus)
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
