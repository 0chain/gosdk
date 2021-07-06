package sdk

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jtguibas/cinema"
)

// YoutubeDL wrap youtube-dl to download video from youtube
type YoutubeDL struct {
	fileReader  *os.File
	fileName    string
	err         error
	cmd         *exec.Cmd
	delay       float64
	clipsIndex  int
	clipsOffset float64
	offset      int64
}

// CreateYoutubeDL create a youtube-dl instance to download video file from youtube
func CreateYoutubeDL(feedURL string, format string, cacheDir string, proxy string, delay int) (*YoutubeDL, error) {

	args := []string{"--no-part", "-q", "-f", format}
	if len(proxy) > 0 {
		args = append(args, "--proxy", proxy)
	}

	cmdGetFileName := exec.Command("youtube-dl", append(args, "--get-filename", "--id", feedURL)...)
	buf, err := cmdGetFileName.Output()

	if err != nil {
		return nil, err
	}
	defer cmdGetFileName.Process.Kill()

	//remove line-break
	fileName := cacheDir + string(os.PathSeparator) + time.Now().Format("2006-01-02_15-04-05") + "_" + strings.Split(string(buf), "\n")[0]

	cmd := exec.Command("youtube-dl", append(args, "-o", fileName, feedURL)...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Start()

	if err != nil {
		return nil, err
	}

	dl := &YoutubeDL{
		fileName: fileName,
		delay:    float64(delay),
		cmd:      cmd,
	}

	go dl.Wait()
	go dl.splitClips()

	return dl, nil
}

// Wait wait youtube-dl to complete or crash
func (dl *YoutubeDL) Wait() {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		dl.Close()
	}()

	dl.err = dl.cmd.Wait()

}

func (dl *YoutubeDL) splitClips() {
	for {

		time.Sleep(5 * time.Second)

		video, err := cinema.Load(dl.fileName)

		if err == nil {

			if video.End().Seconds() > dl.clipsOffset+dl.delay {
				//video.Trim(start time.Duration, end time.Duration)
				video.SetStart(time.Duration(dl.clipsOffset) * time.Second)
				video.SetEnd(time.Duration(dl.delay) * time.Second)

				video.Render(dl.fileName + "." + strconv.Itoa(dl.clipsIndex))

				dl.clipsOffset += dl.delay
				dl.clipsIndex++
			}

		}

	}

}

// GetFileName get video file name
func (dl *YoutubeDL) GetFileName() string {
	return dl.fileName
}

// Read implements io.Raader
func (dl *YoutubeDL) Read(p []byte) (int, error) {

	if dl.fileReader == nil {
		var err error
		for {

			if dl.err != nil {
				return 0, dl.err
			}

			if dl.fileReader == nil {
				dl.fileReader, err = os.Open(dl.fileName)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) { //download is not started yet
						time.Sleep(1 * time.Second)
						continue
					}
				}
				// download is started, and bytes is ready
				break
			}
		}
	}

	wantRead := int64(len(p))

	//loop read bytes till ready
	for {

		if dl.err != nil {
			return 0, dl.err
		}

		fi, _ := dl.fileReader.Stat()

		log.Println(fi.Size() / 1024 / 1024)

		if dl.offset+wantRead < fi.Size() {
			dl.fileReader.Seek(dl.offset, 0)
			readLen, err := dl.fileReader.Read(p)

			dl.offset += int64(readLen)

			return readLen, err
		}

		time.Sleep(1 * time.Second)

	}

}

// Close implements io.Closer
func (dl *YoutubeDL) Close() error {

	if dl != nil {

		if dl.fileReader != nil {
			dl.fileReader.Close()
		}

		if dl.cmd != nil {
			dl.cmd.Process.Kill()
		}
	}

	return nil
}
