package sdk

import (
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
			chunkSize := int(math.Ceil(float64(size) / float64(su.allocationObj.DataShards)))

			paddingBytes := make([]byte, size-chunkSize*su.allocationObj.DataShards)

			//padding data to make that shard has equally sized data
			su.thumbnailBytes = append(buf, paddingBytes...)
			su.fileMeta.ThumbnailSize = len(buf)
			su.thumbailErasureEncoder, _ = reedsolomon.New(su.allocationObj.DataShards, su.allocationObj.ParityShards, reedsolomon.WithAutoGoroutines(chunkSize))

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
