package sdk

// LiveUploadOption set live upload option
type LiveUploadOption func(lu *LiveUpload)

// WithLiveDelay set delayed . ignore if delayed <=0
func WithLiveDelay(delaySeconds int) LiveUploadOption {
	return func(lu *LiveUpload) {
		if delaySeconds > 0 {
			lu.delay = delaySeconds
		}
	}
}

// WithLiveChunkNumber set the number of chunks should be upload in a request. ignore if size <=0
func WithLiveChunkNumber(num int) LiveUploadOption {
	return func(lu *LiveUpload) {
		if num > 0 {
			lu.chunkNumber = num
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
