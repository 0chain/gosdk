//go:build !js && !wasm
// +build !js,!wasm

package sdk

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

// createUploadProgress create a new UploadProgress
func (su *ChunkedUpload) createUploadProgress(connectionId string) {
	if su.progress.ChunkSize <= 0 {
		su.progress = UploadProgress{
			ConnectionID:      connectionId,
			ChunkIndex:        -1,
			ChunkSize:         su.chunkSize,
			EncryptOnUpload:   su.encryptOnUpload,
			EncryptedKeyPoint: su.encryptedKeyPoint,
			ActualSize:        su.fileMeta.ActualSize,
			ChunkNumber:       su.chunkNumber,
		}
	}
	su.progress.Blobbers = make([]*UploadBlobberStatus, su.allocationObj.DataShards+su.allocationObj.ParityShards)

	for i := 0; i < len(su.progress.Blobbers); i++ {
		su.progress.Blobbers[i] = &UploadBlobberStatus{
			Hasher: CreateHasher(su.shardSize),
		}
	}

	su.progress.ID = su.progressID()
	su.saveProgress()
}

// processUpload process upload fragment to its blobber
func (su *ChunkedUpload) processUpload(chunkStartIndex, chunkEndIndex int,
	fileShards []blobberShards, thumbnailShards blobberShards,
	isFinal bool, uploadLength int64) error {

	//chunk has not be uploaded yet
	if chunkEndIndex <= su.progress.ChunkIndex {
		// Write data to hashers
		for i, blobberShard := range fileShards {
			hasher := su.blobbers[i].progress.Hasher
			for _, chunkBytes := range blobberShard {
				err := hasher.WriteToBlockHasher(chunkBytes)
				if err != nil {
					if su.statusCallback != nil {
						su.statusCallback.Error(su.allocationObj.ID, su.fileMeta.RemotePath, su.opCode, err)
					}
					return err
				}
			}
		}
		return nil
	}

	var (
		errCount       int32
		finalBuffer    []blobberData
		pos            uint64
		wg             sync.WaitGroup
		lastBufferOnly bool
	)
	if isFinal {
		finalBuffer = make([]blobberData, len(su.blobbers))
	}
	blobberUpload := UploadData{
		chunkStartIndex: chunkStartIndex,
		chunkEndIndex:   chunkEndIndex,
		isFinal:         isFinal,
		uploadBody:      make([]blobberData, len(su.blobbers)),
		uploadLength:    uploadLength,
	}

	wgErrors := make(chan error, len(su.blobbers))
	if len(fileShards) == 0 {
		return thrown.New("upload_failed", "Upload failed. No data to upload")
	}

	for i := su.uploadMask; !i.Equals64(0); i = i.And(zboxutil.NewUint128(1).Lsh(pos).Not()) {
		pos = uint64(i.TrailingZeros())
		blobber := su.blobbers[pos]
		blobber.progress.UploadLength += uploadLength

		var thumbnailChunkData []byte

		if len(thumbnailShards) > 0 {
			thumbnailChunkData = thumbnailShards[pos]
		}

		wg.Add(1)
		go func(b *ChunkedUploadBlobber, thumbnailChunkData []byte, pos uint64) {
			defer wg.Done()
			uploadData, err := su.formBuilder.Build(
				&su.fileMeta, blobber.progress.Hasher, su.progress.ConnectionID,
				su.chunkSize, chunkStartIndex, chunkEndIndex, isFinal, su.encryptedKey, su.progress.EncryptedKeyPoint,
				fileShards[pos], thumbnailChunkData, su.shardSize)
			if err != nil {
				errC := atomic.AddInt32(&errCount, 1)
				if errC > int32(su.allocationObj.ParityShards-1) { // If atleast data shards + 1 number of blobbers can process the upload, it can be repaired later
					wgErrors <- err
				}
				return
			}
			if isFinal {
				finalBuffer[pos] = blobberData{
					dataBuffers:  uploadData.dataBuffers[len(uploadData.dataBuffers)-1:],
					formData:     uploadData.formData,
					contentSlice: uploadData.contentSlice[len(uploadData.contentSlice)-1:],
				}
				if len(uploadData.dataBuffers) == 1 {
					lastBufferOnly = true
					return
				}
				uploadData.dataBuffers = uploadData.dataBuffers[:len(uploadData.dataBuffers)-1]
			}
			blobberUpload.uploadBody[pos] = uploadData
		}(blobber, thumbnailChunkData, pos)
	}

	wg.Wait()
	close(wgErrors)
	fileShards = nil
	for err := range wgErrors {
		su.removeProgress()
		return thrown.New("upload_failed", fmt.Sprintf("Upload failed. %s", err))
	}
	if !lastBufferOnly {
		su.uploadWG.Add(1)
		select {
		case <-su.ctx.Done():
			return context.Cause(su.ctx)
		case su.uploadChan <- blobberUpload:
		}
	}

	if isFinal {
		close(su.uploadChan)
		su.uploadWG.Wait()
		select {
		case <-su.ctx.Done():
			return context.Cause(su.ctx)
		default:
		}
		blobberUpload.uploadBody = finalBuffer
		return su.uploadToBlobbers(blobberUpload)
	}
	return nil
}

func (su *ChunkedUpload) startProcessor() {
	for i := 0; i < su.uploadWorkers; i++ {
		go su.uploadProcessor()
	}
}
