package sdk

import "time"

// LiveUploadOption set live upload option
type LiveUploadOption func(lu *LiveUpload)

// WithLiveClipsSize set clipsSize . ignore if clipsSize <=0
func WithLiveClipsSize(clipsSize int) LiveUploadOption {
	return func(lu *LiveUpload) {
		if clipsSize > 0 {
			lu.clipsSize = clipsSize
		}
	}
}

// WithLiveDelay set delayed . ignore if delayed <=0
func WithLiveDelay(delaySeconds int) LiveUploadOption {
	return func(lu *LiveUpload) {
		if delaySeconds > 0 {
			lu.delay = time.Duration(delaySeconds) * time.Second
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
