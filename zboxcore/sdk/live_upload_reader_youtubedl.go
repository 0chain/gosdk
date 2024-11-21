package sdk

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// YoutubeDL wrap youtube-dl to download video from youtube
type YoutubeDL struct {
	ctx context.Context

	liveUploadReaderBase

	// cmdYoutubeDL youtube-dl command
	cmdYoutubeDL *exec.Cmd

	// cmdFfmpeg ffmpeg command
	cmdFfmpeg *exec.Cmd
}

// CreateYoutubeDL create a youtube-dl instance to download video file from youtube
//   - localPath: output file path
//   - feedURL: youtube video url
//   - downloadArgs: youtube-dl download arguments
//   - ffmpegArgs: ffmpeg arguments
//   - delay: delay in seconds
func CreateYoutubeDL(ctx context.Context, localPath string, feedURL string, downloadArgs []string, ffmpegArgs []string, delay int) (*YoutubeDL, error) {

	//youtube-dl -f best https://www.youtube.com/watch?v=RfUVIwnsvS8 --proxy http://127.0.0.1:8000 -o - | ffmpeg -i - -flags +cgop -g 30 -hls_time 5 youtube.m3u8

	builder := createFileNameBuilder(localPath)

	argsYoutubeDL := append(downloadArgs,
		"-o", "-",
		feedURL) //output to stdout

	//argsYoutubeDL = append(argsYoutubeDL)

	fmt.Println("[cmd]", "youtube-dl", strings.Join(argsYoutubeDL, " "))

	r, w := io.Pipe()

	cmdYoutubeDL := exec.Command("youtube-dl", argsYoutubeDL...)
	cmdYoutubeDL.Stderr = os.Stderr
	cmdYoutubeDL.Stdout = w

	argsFfmpeg := []string{"-i", "-"}

	argsFfmpeg = append(argsFfmpeg, ffmpegArgs...)
	argsFfmpeg = append(argsFfmpeg,
		"-flags", "+cgop",
		"-g", "30",
		"-hls_time", strconv.Itoa(delay),
		builder.OutFile())

	fmt.Println("[cmd]ffmpeg", strings.Join(argsFfmpeg, " "))
	cmdFfmpeg := exec.Command("ffmpeg", argsFfmpeg...)
	cmdFfmpeg.Stderr = os.Stderr
	cmdFfmpeg.Stdin = r
	cmdFfmpeg.Stdout = os.Stdout

	err := cmdYoutubeDL.Start()
	if err != nil {
		return nil, err
	}

	err = cmdFfmpeg.Start()
	if err != nil {
		return nil, err
	}

	dl := &YoutubeDL{
		ctx: ctx,
		liveUploadReaderBase: liveUploadReaderBase{
			builder:    builder,
			delay:      delay,
			clipsIndex: 0,
		},
		cmdYoutubeDL: cmdYoutubeDL,
		cmdFfmpeg:    cmdFfmpeg,
	}

	go dl.wait()

	return dl, nil
}

func (r *YoutubeDL) wait() {

	go func() {
		<-r.ctx.Done()
		r.Close()
	}()

	go func() {
		r.err = r.cmdFfmpeg.Wait()
	}()

	r.err = r.cmdYoutubeDL.Wait()
}

// Close implements io.Closer
func (r *YoutubeDL) Close() error {
	if r != nil {
		if r.cmd != nil {
			r.cmd.Process.Kill() //nolint
		}

		if r.cmdYoutubeDL != nil {
			r.cmdYoutubeDL.Process.Kill() //nolint
		}

		if r.cmdFfmpeg != nil {
			r.cmdFfmpeg.Process.Kill() //nolint
		}

		if r.clipsReader != nil {
			r.clipsReader.Close()
		}
	}

	return nil
}
