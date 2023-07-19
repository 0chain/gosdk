package sdk

import (
	"context"
	"encoding/hex"
	"math"
	"os"
	"time"

	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/klauspost/reedsolomon"
	"golang.org/x/crypto/sha3"
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

			thumbnailHasher := sha3.New256()
			thumbnailHasher.Write(buf)

			su.fileMeta.ActualThumbnailHash = hex.EncodeToString(thumbnailHasher.Sum(nil))

			su.thumbailErasureEncoder, _ = reedsolomon.New(su.allocationObj.DataShards, su.allocationObj.ParityShards)

		}
	}
}

// WithThumbnailFile add thumbnail from file. stream mode is unnecessary for thumbnail
func WithThumbnailFile(fileName string) ChunkedUploadOption {

	buf, _ := os.ReadFile(fileName)

	return WithThumbnail(buf)
}

// WithChunkNumber set the number of chunks should be upload in a request. ignore if size <=0
func WithChunkNumber(num int) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		if num > 0 {
			su.chunkNumber = num
		}
	}
}

// WithEncrypt turn on/off encrypt on upload. It is turn off as default.
func WithEncrypt(on bool) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.encryptOnUpload = on
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

func WithUploadTimeout(t time.Duration) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.uploadTimeOut = t
	}
}

func WithCommitTimeout(t time.Duration) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.commitTimeOut = t
	}
}

func WithMask(mask zboxutil.Uint128) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		su.uploadMask = mask
	}
}

func WithReaderContext(cancel context.Context) ChunkedUploadOption {
	return func(su *ChunkedUpload) {
		if cancel != nil {
			su.readerCtx = cancel
		}
	}
}
