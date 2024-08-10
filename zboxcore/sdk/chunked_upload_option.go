package sdk

import (
	"crypto/md5"
	"encoding/hex"
	"math"
	"os"
	"time"

	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/klauspost/reedsolomon"
)

// ChunkedUploadOption Generic type for chunked upload option functions 
type ChunkedUploadOption func(su *ChunkedUpload)

// WithThumbnail add thumbnail. stream mode is unnecessary for thumbnail
func WithThumbnail(buf []byte) ChunkedUploadOption {
	return func(su *ChunkedUpload) {

		size := len(buf)

		if size > 0 {
			su.shardUploadedThumbnailSize = int64(math.Ceil(float64(size) / float64(su.allocationObj.DataShards)))

			su.thumbnailBytes = buf
			su.fileMeta.ActualThumbnailSize = int64(len(buf))

			thumbnailHasher := md5.New()
			thumbnailHasher.Write(buf)

			su.fileMeta.ActualThumbnailHash = hex.EncodeToString(thumbnailHasher.Sum(nil))

			su.thumbailErasureEncoder, _ = reedsolomon.New(su.allocationObj.DataShards, su.allocationObj.ParityShards)

		}
	}
}

// WithThumbnailFile add thumbnail from file. stream mode is unnecessary for thumbnail.
// 		- fileName: file name of the thumbnail, which will be read and uploaded
func WithThumbnailFile(fileName string) ChunkedUploadOption {

	buf, _ := os.ReadFile(fileName)

	return WithThumbnail(buf)
}

// WithChunkNumber set the number of chunks should be upload in a request. ignore if size <=0
// 		- num: number of chunks
func WithChunkNumber(num int) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		if num > 0 {
			su.chunkNumber = num
		}
	}
}

// WithEncrypt turn on/off encrypt on upload. It is turn off as default.
// 		- on: true to turn on, false to turn off
func WithEncrypt(on bool) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.encryptOnUpload = on
	}
}

// WithStatusCallback return a wrapper option function to set status callback of the chunked upload instance, which is used to track upload progress
// 		- callback: StatusCallback instance
func WithStatusCallback(callback StatusCallback) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.statusCallback = callback
	}
}

// WithProgressCallback return a wrapper option function to set progress callback of the chunked upload instance
func WithProgressStorer(progressStorer ChunkedUploadProgressStorer) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.progressStorer = progressStorer
	}
}

// WithUploadTimeout return a wrapper option function to set upload timeout of the chunked upload instance
func WithUploadTimeout(t time.Duration) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.uploadTimeOut = t
	}
}

// WithCommitTimeout return a wrapper option function to set commit timeout of the chunked upload instance
func WithCommitTimeout(t time.Duration) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.commitTimeOut = t
	}
}

// WithUploadMask return a wrapper option function to set upload mask of the chunked upload instance
func WithMask(mask zboxutil.Uint128) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.uploadMask = mask
	}
}

// WithEncryptedKeyPoint return a wrapper option function to set encrypted key point of the chunked upload instance
func WithEncryptedPoint(point string) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.encryptedKeyPoint = point
	}
}

// WithActualHash return a wrapper option function to set actual hash of the chunked upload instance
func WithActualHash(hash string) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.fileMeta.ActualHash = hash
	}
}

// WithActualSize return a wrapper option function to set the file hasher used in the chunked upload instance
func WithFileHasher(h Hasher) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.fileHasher = h
	}
}
