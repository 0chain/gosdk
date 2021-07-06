package sdk

import (
	"io"
	"time"
)

// LiveUploadReader wrap io.Reader with delay feature
type LiveUploadReader struct {
	reader    io.Reader
	delay     time.Duration
	clipsSize int
	since     time.Time
	readSize  int
}

func createLiveUploadReader(reader io.Reader, delay time.Duration, clipsSize int) *LiveUploadReader {
	return &LiveUploadReader{
		reader:    reader,
		delay:     delay,
		clipsSize: clipsSize,
		since:     time.Now(),
		readSize:  0,
	}
}

// Read implements io.Reader
func (r *LiveUploadReader) Read(p []byte) (int, error) {

	i, err := r.reader.Read(p)

	if err != nil {
		return i, err
	}

	if r.delay > 0 {
		now := time.Now()
		if now.Sub(r.since) > r.delay {
			r.since = now

			return i, io.EOF
		}
	}

	if r.clipsSize > 0 {
		r.readSize += i
		if r.readSize >= r.clipsSize {
			r.readSize = 0
			return i, io.EOF
		}
	}

	return i, nil

}
