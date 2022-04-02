package sdk

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"math"

	"github.com/klauspost/reedsolomon"
)

// ChunkedUploadOption set stream option
type ChunkedUploadOption func(su *ChunkedUpload)

// WithThumbnail add thumbnail. stream mode is unnecessary for thumbnail
func WithThumbnail(buf []byte) ChunkedUploadOption {
	return func(su *ChunkedUpload) {

		size := len(buf)

		if size > 0 {
			su.shardUploadedThumbnailSize = int64(math.Ceil(float64(size) / float64(su.allocationObj.DataShards)))

			su.thumbnailBytes = buf
			su.fileMeta.ActualThumbnailSize = int64(len(buf))

			thumbnailHasher := sha256.New()
			thumbnailHasher.Write(buf)

			su.fileMeta.ActualThumbnailHash = hex.EncodeToString(thumbnailHasher.Sum(nil))

			su.thumbailErasureEncoder, _ = reedsolomon.New(su.allocationObj.DataShards, su.allocationObj.ParityShards)

		}
	}
}

// WithThumbnailFile add thumbnail from file. stream mode is unnecessary for thumbnail
func WithThumbnailFile(fileName string) ChunkedUploadOption {

	buf, _ := ioutil.ReadFile(fileName)

	return WithThumbnail(buf)
}

// WithChunkSize set custom chunk size. ignore if size <=0
func WithChunkSize(size int64) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		if size > 0 {
			su.chunkSize = size
		}
	}
}

// WithChunkNumber set the number of chunks should be upload in a request. ignore if size <=0
func WithChunkNumber(num int) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		if num > 0 {
			su.chunkNumber = num
		}
	}
}

// WithEncrypt trun on/off encrypt on upload. It is turn off as default.
func WithEncrypt(status bool) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.encryptOnUpload = status
	}
}

// WithStatusCallback register StatusCallback instance
func WithStatusCallback(callback StatusCallback) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.statusCallback = callback
	}
}

func WithProgressStorer(progressStorer ChunkedUploadProgressStorer) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.progressStorer = progressStorer
	}
}
