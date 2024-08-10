//go:build !js && !wasm
// +build !js,!wasm

package sdk

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// FfmpegRecorder wrap ffmpeg command to capture video and audio from local camera and microphone
type FfmpegRecorder struct {
	liveUploadReaderBase
}

// CreateFfmpegRecorder create a ffmpeg commander to capture video and audio  local camera and microphone
//   - file: output file path
//   - delay: delay in seconds
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
		liveUploadReaderBase: liveUploadReaderBase{
			builder:    builder,
			delay:      delay,
			cmd:        cmd,
			clipsIndex: 0,
		},
	}

	go fr.wait()

	return fr, nil
}
