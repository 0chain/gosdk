package sdk

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

// YoutubeDL wrap youtube-dl to download video from youtube
type YoutubeDL struct {
	liveUploadReaderBase

	// cmdYoutubeDL youtube-dl command
	cmdYoutubeDL *exec.Cmd

	// cmdFfmpeg ffmpeg command
	cmdFfmpeg *exec.Cmd
}

// CreateYoutubeDL create a youtube-dl instance to download video file from youtube
func CreateYoutubeDL(localPath string, feedURL string, format string, proxy string, delay int) (*YoutubeDL, error) {

	//youtube-dl -f best https://www.youtube.com/watch?v=qjNQfSobVwE --proxy http://127.0.0.1:8000 -o - | ffmpeg -i - -flags +cgop -g 30 -hls_time 5 youtube.m3u8

	builder := createFileNameBuilder(localPath)

	argsYoutubeDL := []string{
		//"-q",
		"-f", format,
		feedURL,
		"-o", "-"} //output to stdout

	if len(proxy) > 0 {
		argsYoutubeDL = append(argsYoutubeDL, "--proxy", proxy)
	}

	argsYoutubeDL = append(argsYoutubeDL, "|", "ffmpeg")

	fmt.Println("[cmd]", "youtube-dl", strings.Join(argsYoutubeDL, " "))

	r, w := io.Pipe()

	cmdYoutubeDL := exec.Command("youtube-dl", argsYoutubeDL...)
	cmdYoutubeDL.Stderr = os.Stderr
	cmdYoutubeDL.Stdout = w

	argsFfmpeg := []string{
		"-i", "-",
		"-flags", "+cgop",
		"-g", "30",
		"-hls_time", strconv.Itoa(delay),
		builder.OutFile(),
	}

	fmt.Println("ffmpeg", strings.Join(argsFfmpeg, " "))
	cmdFfmpeg := exec.Command("ffmpeg", argsFfmpeg...)
	cmdFfmpeg.Stderr = os.Stderr
	cmdFfmpeg.Stdin = r

	err := cmdYoutubeDL.Start()
	if err != nil {
		return nil, err
	}

	err = cmdFfmpeg.Start()
	if err != nil {
		return nil, err
	}

	dl := &YoutubeDL{

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
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
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
			r.cmd.Process.Kill()
		}

		if r.cmdYoutubeDL != nil {
			r.cmdYoutubeDL.Process.Kill()
		}

		if r.cmdFfmpeg != nil {
			r.cmdFfmpeg.Process.Kill()
		}

		if r.clipsReader != nil {
			r.clipsReader.Close()
		}
	}

	return nil
}
