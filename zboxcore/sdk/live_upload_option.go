package sdk

// LiveUploadOption set live upload option
type LiveUploadOption func(lu *LiveUpload)

// WithLiveDelay set number of seconds to wait before starting the live upload. Ignore if less than 0.
// 		- delaySeconds: number of seconds
func WithLiveDelay(delaySeconds int) LiveUploadOption {
	return func(lu *LiveUpload) {
		if delaySeconds > 0 {
			lu.delay = delaySeconds
		}
	}
}

// WithLiveChunkNumber set the number of chunks should be upload in a live upload request. Ignore if size <=0.
// 		- num: number of chunks
func WithLiveChunkNumber(num int) LiveUploadOption {
	return func(lu *LiveUpload) {
		if num > 0 {
			lu.chunkNumber = num
		}
	}
}

// WithLiveEncrypt trun on/off encrypt on upload. It is turned off by default.
// 		- status: true to turn on, false to turn off
func WithLiveEncrypt(status bool) LiveUploadOption {
	return func(lu *LiveUpload) {
		lu.encryptOnUpload = status
	}
}

// WithLiveStatusCallback register StatusCallback instance to track live upload progress
// 		- callback: StatusCallback instance
func WithLiveStatusCallback(callback func() StatusCallback) LiveUploadOption {
	return func(lu *LiveUpload) {
		lu.statusCallback = callback
	}
}
