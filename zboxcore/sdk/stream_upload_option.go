package sdk

import (
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"math"

	"github.com/klauspost/reedsolomon"
)

// StreamUploadOption set stream option
type StreamUploadOption func(su *StreamUpload)

// WithThumbnail add thumbnail. stream mode is unnecessary for thumbnail
func WithThumbnail(buf []byte) StreamUploadOption {
	return func(su *StreamUpload) {

		size := len(buf)

		if size > 0 {
			su.shardThumbnailSize = int64(math.Ceil(float64(size) / float64(su.allocationObj.DataShards)))

			su.thumbnailBytes = buf
			su.fileMeta.ActualThumbnailSize = int64(len(buf))

			thumbnailHasher := sha1.New()
			thumbnailHasher.Write(buf)

			su.fileMeta.ActualThumbnailHash = hex.EncodeToString(thumbnailHasher.Sum(nil))

			su.thumbailErasureEncoder, _ = reedsolomon.New(su.allocationObj.DataShards, su.allocationObj.ParityShards)

		}
	}
}

// WithThumbnailFile add thumbnail from file. stream mode is unnecessary for thumbnail
func WithThumbnailFile(fileName string) StreamUploadOption {

	buf, _ := ioutil.ReadFile(fileName)

	return WithThumbnail(buf)
}

// WithChunkSize set custom chunk size. ignore if size <=0
func WithChunkSize(size int) StreamUploadOption {
	return func(su *StreamUpload) {
		if size > 0 {
			su.chunkSize = size
		}
	}
}

// WithEncrypt trun on/off encrypt on upload. It is turn off as default.
func WithEncrypt(status bool) StreamUploadOption {
	return func(su *StreamUpload) {
		su.encryptOnUpload = status
	}
}

// WithStatusCallback register StatusCallback instance
func WithStatusCallback(callback StatusCallback) StreamUploadOption {
	return func(su *StreamUpload) {
		su.statusCallback = callback
	}
}
