package sdk

import "time"

// LiveUploadOption set live upload option
type LiveUploadOption func(lu *LiveUpload)

// WithLiveDelay set delayed . ignore if delayed <=0
func WithLiveDelay(delay time.Duration) LiveUploadOption {
	return func(lu *LiveUpload) {
		if delay > 0 {
			lu.delay = delay
		}
	}
}

// WithLiveChunkSize set custom chunk size. ignore if size <=0
func WithLiveChunkSize(size int) LiveUploadOption {
	return func(lu *LiveUpload) {
		if size > 0 {
			lu.chunkSize = size
		}
	}
}

// WithLiveEncrypt trun on/off encrypt on upload. It is turn off as default.
func WithLiveEncrypt(status bool) LiveUploadOption {
	return func(lu *LiveUpload) {
		lu.encryptOnUpload = status
	}
}

// WithLiveStatusCallback register StatusCallback instance
func WithLiveStatusCallback(callback func() StatusCallback) LiveUploadOption {
	return func(lu *LiveUpload) {
		lu.statusCallback = callback
	}
}
