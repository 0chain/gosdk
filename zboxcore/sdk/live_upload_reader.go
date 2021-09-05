package sdk

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/h2non/filetype"
)

var (
	// ErrClispIsNotReady clips file is not ready
	ErrClispIsNotReady = errors.New("live: clips is not ready")
)

// LiveUploadReader implements io.Reader and Size for live stream upload
type LiveUploadReader interface {
	io.Reader
	Size() int64
	GetClipsFile(clipsIndex int) string
	GetClipsFileName(cliipsIndex int) string
}

type liveUploadReaderBase struct {
	builder FileNameBuilder

	// delay segment time of output
	delay int

	// cmd ffmpeg command
	cmd *exec.Cmd
	// err last err
	err error

	// clipsIndex current clips index
	clipsIndex int
	// clipsReader file reader of current clips
	clipsReader *os.File
	// clipsOffset how much bytes is read
	clipsOffset int64
}

func (r *liveUploadReaderBase) wait() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		r.Close()
	}()

	r.err = r.cmd.Wait()
}

// GetClipsFile get clips file
func (r *liveUploadReaderBase) GetClipsFile(clipsIndex int) string {
	return r.builder.ClipsFile(clipsIndex)
}

// GetClipsFileName get clips file name
func (r *liveUploadReaderBase) GetClipsFileName(clipsIndex int) string {
	return r.builder.ClipsFileName(clipsIndex)
}

// Read implements io.Raader
func (r *liveUploadReaderBase) Read(p []byte) (int, error) {

	err := r.initClipsReader()

	if err != nil {
		return 0, err
	}

	for {

		if r.err != nil {
			return 0, r.err
		}

		fi, _ := r.clipsReader.Stat()

		if fi != nil {

			size := fi.Size()

			wantRead := int64(len(p))

			if r.clipsOffset+wantRead < size {
				readLen, err := r.clipsReader.Read(p)

				r.clipsOffset += int64(readLen)

				return readLen, err
			}

			readLen, err := r.clipsReader.Read(p)

			r.clipsReader.Close()
			r.clipsReader = nil
			r.clipsOffset = 0
			r.clipsIndex++

			return readLen, err
		}

		time.Sleep(1 * time.Second)

	}
}

// Close implements io.Closer
func (r *liveUploadReaderBase) Close() error {
	if r != nil {
		if r.cmd != nil {
			r.cmd.Process.Kill()
		}

		if r.clipsReader != nil {
			r.clipsReader.Close()
		}
	}

	return nil
}

// GetFileContentType get MIME type
func (r *liveUploadReaderBase) GetFileContentType() (string, error) {
	for {

		if r.err != nil {
			return "", r.err
		}

		currentClips := r.GetClipsFile(r.clipsIndex)
		reader, err := os.Open(currentClips)

		if err == nil {
			defer reader.Close()

			for {
				fi, _ := reader.Stat()

				if fi.Size() > 261 {
					buffer := make([]byte, 261)
					_, err := reader.Read(buffer)

					if err != nil {
						return "", err
					}

					kind, _ := filetype.Match(buffer)
					if kind == filetype.Unknown {
						return "application/octet-stream", nil
					}

					return kind.MIME.Value, nil
				}

				time.Sleep(1 * time.Second)
			}

		}

	}

}

// Size get current clips size
func (r *liveUploadReaderBase) Size() int64 {
	err := r.initClipsReader()

	if err != nil {
		return 0
	}

	for {

		fi, _ := r.clipsReader.Stat()

		if fi != nil {
			return fi.Size()
		}

		time.Sleep(1 * time.Second)
	}

}

func (r *liveUploadReaderBase) initClipsReader() error {

	if r.clipsReader == nil {

		nextClips := r.GetClipsFile(r.clipsIndex + 1)

		for {

			if r.err != nil {
				return r.err
			}

			// file content is less than bytes want to read, check whether current clips file is ended
			_, err := os.Stat(nextClips)

			if err == nil {
				if r.clipsReader == nil {
					r.clipsReader, err = os.Open(r.GetClipsFile(r.clipsIndex))

					if err != nil {
						return err
					}

					return nil
				}
			}

			time.Sleep(1 * time.Second)
		}
	}

	return nil
}
