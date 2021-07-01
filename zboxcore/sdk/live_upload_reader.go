package sdk

import (
	"io"
	"time"
)

// LiveUploadReader wrap io.Reader with delay feature
type LiveUploadReader struct {
	reader io.Reader
	delay  time.Duration
	since  time.Time
}

func createLiveUploadReader(reader io.Reader, delay time.Duration) *LiveUploadReader {
	return &LiveUploadReader{
		reader: reader,
		delay:  delay,
		since:  time.Now(),
	}
}

// Read implements io.Reader
func (r *LiveUploadReader) Read(p []byte) (int, error) {

	i, err := r.reader.Read(p)

	if err != nil {
		return i, err
	}

	now := time.Now()
	if now.Sub(r.since) > r.delay {
		r.since = now

		return i, io.EOF
	}

	return i, nil

}
