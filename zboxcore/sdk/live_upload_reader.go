package sdk

import (
	"errors"
	"io"
)

var (
	// ErrClispIsNotReady clips file is not ready
	ErrClispIsNotReady = errors.New("live: clips is not ready")
)

// LiveUploadReader implements io.Reader and Size for live stream upload
type LiveUploadReader interface {
	io.Reader
	Size() int64
	GetFileName(clipsIndex int) string
}

// // liveUploadReader wrap io.Reader with delay feature
// type liveUploadReader struct {
// 	reader    io.Reader
// 	delay     int
// 	clipsSize int
// 	since     time.Time
// 	readSize  int
// }

// func createLiveUploadReader(reader io.Reader, delay, clipsSize int) *LiveUploadReader {
// 	return &liveUploadReader{
// 		reader:    reader,
// 		delay:     delay,
// 		clipsSize: clipsSize,
// 		since:     time.Now(),
// 		readSize:  0,
// 	}
// }

// // Read implements io.Reader
// func (r *liveUploadReader) Read(p []byte) (int, error) {

// 	i, err := r.reader.Read(p)

// 	if err != nil {
// 		return i, err
// 	}

// }
