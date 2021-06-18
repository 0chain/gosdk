package sdk

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"math"
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
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/sha3"
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
		fileHasher: util.NewStreamMerkleHasher(func(left, right string) string {
			return coreEncryption.Hash(left + right)
		}),

		uploadMask:      zboxutil.NewUint128(1).Lsh(uint64(len(allocationObj.Blobbers))).Sub64(1),
		chunkSize:       DefaultChunkSize,
		encryptOnUpload: false,
	}

	home, _ := homedir.Dir()

	su.configDir = home + string(os.PathSeparator) + ".zcn"

	//create upload folder to save progress
	os.MkdirAll(su.configDir+"/upload", os.ModePerm)

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

	configDir string

	allocationObj *Allocation

	progress   UploadProgress
	uploadMask zboxutil.Uint128

	fileMeta           FileMeta
	fileReader         io.Reader
	fileErasureEncoder reedsolomon.Encoder
	fileEncscheme      encryption.EncryptionScheme
	fileHasher         *util.StreamMerkleHasher

	thumbnailBytes         []byte
	thumbailErasureEncoder reedsolomon.Encoder

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

	blobbers []*StreamUploadBobbler
}

// progressID build local progress id with [allocationid]_[Hash(LocalPath+"_"+RemotePath)]_[RemoteName] format
func (su *StreamUpload) progressID() string {

	return su.configDir + "/upload/" + su.allocationObj.ID + "_" + su.fileMeta.FileID()
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
		//b.MerkleHasher.Hash = DefaultHashFunc
		b.MerkleHashes = make([]hash.Hash, 1024)
		for idx := range b.MerkleHashes {
			b.MerkleHashes[idx] = sha3.New256()
		}
		b.ShardHasher = sha1.New()
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

// removeProgress remove progress info once it is done
func (su *StreamUpload) removeProgress() {

	os.Remove(su.progressID())
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
			MerkleHashes: make([]hash.Hash, 1024),
			ShardHasher:  sha1.New(),
		}

		for idx := range progress.Blobbers[i].MerkleHashes {
			progress.Blobbers[i].MerkleHashes[idx] = sha3.New256()
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
		su.statusCallback.Started(su.allocationObj.ID, su.fileMeta.RemotePath, OpUpload, int(su.fileMeta.ActualSize)+int(su.fileMeta.ActualThumbnailSize))
	}

	for i := 0; ; i++ {
		fileShards, readLen, chunkSize, isFinal, err := su.readNextChunks(i)
		if err != nil {
			return err
		}

		su.shardUploadedSize += chunkSize

		//skip chunk if it has been uploaded
		if i < su.progress.ChunkIndex {
			continue
		}

		if isFinal {
			su.fileMeta.ActualHash = su.fileHasher.GetMerkleRoot()
		}

		// upload entire thumbnail in first reqeust only
		if i == 0 && len(su.thumbnailBytes) > 0 {

			thumbnailShards, err := su.readThumbnailShards()
			if err != nil {
				return err
			}

			su.processUpload(i, fileShards, thumbnailShards, isFinal, readLen)

			su.progress.UploadLength += int64(su.fileMeta.ActualThumbnailSize) + readLen
		} else {
			su.processUpload(i, fileShards, nil, isFinal, readLen)
		}

		su.progress.ChunkIndex = i
		su.progress.UploadLength += int64(readLen)
		su.saveProgress()

		if su.statusCallback != nil {
			su.statusCallback.InProgress(su.allocationObj.ID, su.fileMeta.RemotePath, OpUpload, int(su.progress.UploadLength), nil)
		}

		if isFinal {
			su.removeProgress()
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

func (su *StreamUpload) readNextChunks(chunkIndex int) ([][]byte, int64, int64, bool, error) {

	chunkSize := su.chunkSize

	if su.encryptOnUpload {
		chunkSize -= 16
		chunkSize -= 2 * 1024
	}

	shardSize := chunkSize * su.allocationObj.DataShards

	isFinal := false
	chunkBytes := make([]byte, shardSize)
	readLen, err := su.fileReader.Read(chunkBytes)

	if readLen > 0 {
		hash := sha1.New()
		hash.Write(chunkBytes[:readLen])
		leafHash := hex.EncodeToString(hash.Sum(nil))
		su.fileHasher.Push(leafHash, chunkIndex)
	}

	if readLen < shardSize {
		chunkSize = int(math.Ceil(float64(readLen / su.allocationObj.DataShards)))
		chunkBytes = chunkBytes[:readLen]
	}

	if err != nil {
		// all bytes are read
		if errors.Is(err, io.EOF) {
			isFinal = true
		} else {
			return nil, int64(readLen), int64(chunkSize), isFinal, err
		}
	}

	if su.fileMeta.ActualSize > 0 && su.progress.UploadLength+int64(readLen) >= su.fileMeta.ActualSize {
		isFinal = true
	}

	shards, err := su.fileErasureEncoder.Split(chunkBytes)
	if err != nil {
		logger.Logger.Error("[upload] Erasure coding on thumbnail failed:", err.Error())
		return nil, int64(readLen), int64(chunkSize), isFinal, err
	}

	err = su.fileErasureEncoder.Encode(shards)
	if err != nil {
		logger.Logger.Error("[upload] Erasure coding on thumbnail failed:", err.Error())
		return nil, int64(readLen), int64(chunkSize), isFinal, err
	}

	var pos uint64
	if su.encryptOnUpload {
		for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
			pos = uint64(i.TrailingZeros())
			encMsg, err := su.fileEncscheme.Encrypt(shards[pos])
			if err != nil {
				logger.Logger.Error("[upload] Encryption on thumbnail failed:", err.Error())
				return nil, int64(readLen), int64(chunkSize), isFinal, err
			}
			header := make([]byte, 2*1024)
			copy(header[:], encMsg.MessageChecksum+","+encMsg.OverallChecksum)
			shards[pos] = append(header, encMsg.EncryptedData...)
		}
	}

	return shards, int64(readLen), int64(chunkSize), isFinal, nil

}

//processUpload process upload shard to its blobber
func (su *StreamUpload) processUpload(chunkIndex int, fileShards [][]byte, thumbnailShards [][]byte, isFinal bool, uploadLenght int64) {
	threads := su.allocationObj.DataShards + su.allocationObj.ParityShards

	wg := &sync.WaitGroup{}
	wg.Add(threads)

	var pos uint64
	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())

		blobber := su.blobbers[pos]
		blobber.progress.UploadLength += uploadLenght

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
		blobber.fileRef.NumBlocks = int64(su.progress.ChunkIndex + 1)

		newChange := &allocationchange.NewFileChange{}
		newChange.File = blobber.fileRef

		newChange.NumBlocks = blobber.fileRef.NumBlocks
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

	su.removeProgress()

	if su.statusCallback != nil {
		su.statusCallback.Completed(su.allocationObj.ID, su.fileMeta.RemotePath, su.fileMeta.RemoteName, su.fileMeta.MimeType, int(su.progress.UploadLength), OpUpload)
	}

	return nil
}
