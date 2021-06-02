package sdk

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/0chain/gosdk/core/common"
	coreEncryption "github.com/0chain/gosdk/core/encryption"
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
)

// DefaultChunkSize default chunk size for file and thumbnail
const DefaultChunkSize = 64 * 1024

// CreateStreamUpload create a StreamUpload instance
func CreateStreamUpload(allocationObj *Allocation, fileMeta FileMeta, fileReader io.Reader, opts ...StreamUploadOption) *StreamUpload {

	su := &StreamUpload{
		allocationObj: allocationObj,
		fileMeta:      fileMeta,
		fileReader:    fileReader,

		uploadMask:      zboxutil.NewUint128(1).Lsh(uint64(len(allocationObj.Blobbers))).Sub64(1),
		chunkSize:       DefaultChunkSize,
		encryptOnUpload: false,
	}

	su.loadProgress()

	for _, opt := range opts {
		opt(su)
	}

	// encrypt option has been chaned.upload it from scratch
	if su.progress.EncryptOnUpload != su.encryptOnUpload {
		su.progress = su.createUploadProgress()
	}

	// chunkSize has been changed. upload it from scratch
	if su.progress.ChunkSize != su.chunkSize {
		su.progress = su.createUploadProgress()
	}

	su.fileErasureEncoder, _ = reedsolomon.New(su.allocationObj.DataShards, su.allocationObj.ParityShards, reedsolomon.WithAutoGoroutines(su.chunkSize))

	if su.encryptOnUpload {
		su.fileEncscheme = su.createEncscheme()

	}

	su.blobbers = make([]*StreamUploadBobbler, len(su.allocationObj.Blobbers))

	for i := 0; i < len(su.allocationObj.Blobbers); i++ {
		su.blobbers[i] = &StreamUploadBobbler{
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

	return su

}

// StreamUpload upload manager with resumable upload feature
type StreamUpload struct {
	Consensus

	allocationObj *Allocation

	progress   UploadProgress
	uploadMask zboxutil.Uint128

	fileMeta           FileMeta
	fileReader         io.Reader
	fileErasureEncoder reedsolomon.Encoder
	fileEncscheme      encryption.EncryptionScheme

	thumbnailBytes         []byte
	thumbailErasureEncoder reedsolomon.Encoder

	// encryptOnUpload encrypt data on upload or not.
	encryptOnUpload bool
	// chunkSize how much bytes a chunk has. 64KB is default value.
	chunkSize int

	// statusCallback trigger progress on StatusCallback
	statusCallback StatusCallback

	blobbers []*StreamUploadBobbler
}

// progressID build local progress id with [allocationid]_[encodeURI(localpath)]_[encodeURI(remotepath)] format
func (su *StreamUpload) progressID() string {
	return "~/.zcn/upload/" + su.allocationObj.ID + "_" + su.fileMeta.FileID()
}

// loadProgress load progress from ~/.zcn/upload/[progressID]
func (su *StreamUpload) loadProgress() {
	progressID := su.progressID()

	buf, err := ioutil.ReadFile(progressID)

	if errors.Is(err, os.ErrNotExist) {
		logger.Logger.Info("[upload] init progress: ", progressID)
		su.progress = su.createUploadProgress()
		return
	}

	progress := UploadProgress{}
	if err := json.Unmarshal(buf, &progress); err != nil {
		logger.Logger.Info("[upload] init progress failed: ", err, ", upload it from scratch")
		su.progress = su.createUploadProgress()
		return
	}

	for _, b := range progress.Blobbers {
		b.MerkleHasher.Hash = DefaultHashFunc
	}

	su.progress = progress
}

// saveProgress save progress to ~/.zcn/upload/[progressID]
func (su *StreamUpload) saveProgress() {
	buf, err := json.Marshal(su.progress)
	if err != nil {
		logger.Logger.Error("[upload] save progress: ", err)
	}

	progressID := su.progressID()
	err = ioutil.WriteFile(progressID, buf, 0644)
	if err != nil {
		logger.Logger.Error("[upload] save progress: ", err)
		return
	}

	logger.Logger.Info("[upload] save progress: ", progressID)
}

// createUploadProgress create a new UploadProgress
func (su *StreamUpload) createUploadProgress() UploadProgress {
	progress := UploadProgress{ConnectionID: zboxutil.NewConnectionId(),
		ChunkIndex:   0,
		ChunkSize:    su.chunkSize,
		UploadLength: 0,
		Blobbers:     make([]*UploadBlobberStatus, su.allocationObj.DataShards+su.allocationObj.ParityShards),
	}

	for i := 0; i < len(progress.Blobbers); i++ {
		progress.Blobbers[i] = &UploadBlobberStatus{
			MerkleHasher: util.StreamMerkleHasher{
				Hash: DefaultHashFunc,
			},
		}
	}

	return progress
}

func (su *StreamUpload) createEncscheme() encryption.EncryptionScheme {
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
func (su *StreamUpload) Start() error {

	if su.statusCallback != nil {
		su.statusCallback.Started(su.allocationObj.ID, su.fileMeta.RemotePath, OpUpload, int(su.fileMeta.Size)+su.fileMeta.ThumbnailSize)
	}

	for i := 0; ; i++ {
		fileShards, readLen, isFinal, err := su.readNextShards(i < su.progress.ChunkIndex)
		if err != nil {
			return err
		}

		//skip chunk if it has been uploaded
		if i < su.progress.ChunkIndex {
			continue
		}

		// upload entire thumbnail in first reqeust only
		if i == 0 {
			thumbnailShards, err := su.readThumbnailShards()
			if err != nil {
				return err
			}

			su.processUpload(i, fileShards, thumbnailShards, isFinal)

			su.progress.UploadLength += int64(su.fileMeta.ThumbnailSize)
		} else {
			su.processUpload(i, fileShards, nil, isFinal)
		}

		su.progress.ChunkIndex = i
		su.progress.UploadLength += int64(readLen)
		su.saveProgress()

		if su.statusCallback != nil {
			su.statusCallback.InProgress(su.allocationObj.ID, su.fileMeta.RemotePath, OpUpload, int(su.progress.UploadLength), nil)
		}

		if isFinal {
			break
		}
	}

	return su.processCommit()
}

// readThumbnailShards encode and encrypt thumbnail
func (su *StreamUpload) readThumbnailShards() ([][]byte, error) {

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

	var c, pos uint64 = 0, 0
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

		c, pos = 0, 0
	}

	return shards, nil
}

func (su *StreamUpload) readNextShards(uploaded bool) ([][]byte, int, bool, error) {

	chunkSize := su.chunkSize

	if su.encryptOnUpload {
		chunkSize -= 16
		chunkSize -= 2 * 1024
	}

	isFinal := false
	chunkBytes := make([]byte, chunkSize*(su.allocationObj.DataShards+su.allocationObj.ParityShards))
	readLen, err := su.fileReader.Read(chunkBytes)

	if err != nil {
		// all bytes are read
		if errors.Is(err, io.EOF) {
			isFinal = true
		} else {
			return nil, readLen, isFinal, err
		}
	}

	if su.progress.UploadLength+int64(readLen) == su.fileMeta.Size {
		isFinal = true
	}

	shards, err := su.thumbailErasureEncoder.Split(su.thumbnailBytes)
	if err != nil {
		logger.Logger.Error("[upload] Erasure coding on thumbnail failed:", err.Error())
		return nil, readLen, isFinal, err
	}

	err = su.thumbailErasureEncoder.Encode(shards)
	if err != nil {
		logger.Logger.Error("[upload] Erasure coding on thumbnail failed:", err.Error())
		return nil, readLen, isFinal, err
	}

	var pos uint64
	if su.encryptOnUpload {
		for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
			pos = uint64(i.TrailingZeros())
			encMsg, err := su.fileEncscheme.Encrypt(shards[pos])
			if err != nil {
				logger.Logger.Error("[upload] Encryption on thumbnail failed:", err.Error())
				return nil, readLen, isFinal, err
			}
			header := make([]byte, 2*1024)
			copy(header[:], encMsg.MessageChecksum+","+encMsg.OverallChecksum)
			shards[pos] = append(header, encMsg.EncryptedData...)
		}
	}

	return shards, readLen, isFinal, nil

}

//processUpload process upload shard to its blobber
func (su *StreamUpload) processUpload(chunkIndex int, fileShards [][]byte, thumbnailShards [][]byte, isFinal bool) {
	threads := su.allocationObj.DataShards + su.allocationObj.ParityShards

	wg := &sync.WaitGroup{}
	wg.Add(threads)

	var pos uint64
	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		blobber := su.blobbers[pos]

		if len(thumbnailShards) > 0 {
			go blobber.processUpload(su, chunkIndex, fileShards[pos], thumbnailShards[pos], isFinal, wg)
		} else {
			go blobber.processUpload(su, chunkIndex, fileShards[pos], nil, isFinal, wg)
		}
	}

	wg.Wait()

}

// processCommit commit shard upload on its blobber
func (su *StreamUpload) processCommit() error {
	logger.Logger.Info("Closed all the channels. Submitting for commit")
	su.consensus = 0
	wg := &sync.WaitGroup{}
	ones := su.uploadMask.CountOnes()
	wg.Add(ones)

	var pos uint64
	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		blobber := su.blobbers[pos]

		newChange := &allocationchange.NewFileChange{}
		newChange.File = blobber.fileRef
		newChange.NumBlocks = int64(su.progress.ChunkIndex)
		newChange.Operation = allocationchange.INSERT_OPERATION
		newChange.Size = blobber.fileRef.Size
		newChange.File.Attributes = blobber.fileRef.Attributes
		blobber.commitChanges = append(blobber.commitChanges, newChange)

		go blobber.processCommit(su, wg)
	}
	wg.Wait()

	if !su.isConsensusOk() {
		if su.consensus != 0 {
			logger.Logger.Info("Commit consensus failed, Deleting remote file....")
			su.allocationObj.deleteFile(su.fileMeta.RemotePath, su.consensus, su.consensus)
		}
		if su.statusCallback != nil {
			su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, OpUpload, common.NewError("commit_consensus_failed", "Upload failed as there was no commit consensus"))
			return nil
		}
	}

	if su.statusCallback != nil {
		su.statusCallback.Completed(su.allocationObj.ID, su.fileMeta.RemotePath, su.fileMeta.RemoteName, su.fileMeta.MimeType, int(su.progress.UploadLength), OpUpload)
	}

	return nil
}
