package sdk

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/h2non/filetype"
)

// FfmpegRecorder wrap ffmpeg command to capture video and audio from local camera and microphone
type FfmpegRecorder struct {
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

// CreateFfmpegRecorder create a ffmpeg commander to capture video and audio  local camera and microphone
func CreateFfmpegRecorder(file string, delay int) (*FfmpegRecorder, error) {

	builder := createFileNameBuilder(file)

	args := buildFfmpegArgs(builder.OutFile(), delay)

	fmt.Println("ffmpeg", strings.Join(args, " "))

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Start()

	if err != nil {
		return nil, err
	}

	fr := &FfmpegRecorder{
		builder:    builder,
		delay:      delay,
		cmd:        cmd,
		clipsIndex: 0,
	}

	go fr.wait()

	return fr, nil
}

func (fr *FfmpegRecorder) wait() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fr.Close()
	}()

	fr.err = fr.cmd.Wait()
}

// GetClipsFile get clips file
func (fr *FfmpegRecorder) GetClipsFile(clipsIndex int) string {
	return fr.builder.ClipsFile(clipsIndex)
}

// GetFileName get clips file name
func (fr *FfmpegRecorder) GetClipsFileName(clipsIndex int) string {
	return fr.builder.ClipsFileName(clipsIndex)
}

// Read implements io.Raader
func (fr *FfmpegRecorder) Read(p []byte) (int, error) {

	err := fr.initClipsReader()

	if err != nil {
		return 0, err
	}

	for {

		if fr.err != nil {
			return 0, fr.err
		}

		fi, _ := fr.clipsReader.Stat()

		if fi != nil {

			size := fi.Size()

			wantRead := int64(len(p))

			if fr.clipsOffset+wantRead < size {
				readLen, err := fr.clipsReader.Read(p)

				fr.clipsOffset += int64(readLen)

				return readLen, err
			}

			readLen, err := fr.clipsReader.Read(p)

			fr.clipsReader.Close()
			fr.clipsReader = nil
			fr.clipsOffset = 0
			fr.clipsIndex++

			return readLen, err
		}

		time.Sleep(1 * time.Second)

	}
}

// Close implements io.Closer
func (fr *FfmpegRecorder) Close() error {
	if fr != nil {
		if fr.cmd != nil {
			fr.cmd.Process.Kill()
		}

		if fr.clipsReader != nil {
			fr.clipsReader.Close()
		}
	}

	return nil
}

// GetFileContentType get MIME type
func (fr *FfmpegRecorder) GetFileContentType() (string, error) {
	for {

		if fr.err != nil {
			return "", fr.err
		}

		reader, err := os.Open(fr.GetClipsFile(fr.clipsIndex))

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
func (fr *FfmpegRecorder) Size() int64 {
	err := fr.initClipsReader()

	if err != nil {
		return 0
	}

	for {

		fi, _ := fr.clipsReader.Stat()

		if fi != nil {
			return fi.Size()
		}

		time.Sleep(1 * time.Second)
	}

}

func (fr *FfmpegRecorder) initClipsReader() error {

	if fr.clipsReader == nil {

		nextClips := fr.GetClipsFile(fr.clipsIndex + 1)

		for {

			if fr.err != nil {
				return fr.err
			}

			// file content is less than bytes want to read, check whether current clips file is ended
			_, err := os.Stat(nextClips)

			if err == nil {
				if fr.clipsReader == nil {
					fr.clipsReader, err = os.Open(fr.GetClipsFile(fr.clipsIndex))

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
