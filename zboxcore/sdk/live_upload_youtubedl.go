package sdk

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// YoutubeDL wrap youtube-dl to download video from youtube
type YoutubeDL struct {
	fileReader *os.File
	fileName   string
	err        error
	cmd        *exec.Cmd
}

// CreateYoutubeDL create a youtube-dl instance to download video file from youtube
func CreateYoutubeDL(feedURL string, format string, cacheDir string, proxy string) (*YoutubeDL, error) {

	args := []string{"--no-part", "-f", format}
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
		cmd:      cmd,
	}

	go dl.Wait()

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

	n := len(p)

	readLen := 0

	for i := 0; i < n; i++ {
		//loop read bytes till ready
		for {

			if dl.err != nil {
				return readLen, dl.err
			}

			if dl.fileReader != nil {
				buf := make([]byte, 1)
				m, _ := dl.fileReader.Read(buf)

				if m == 1 {
					readLen++
					p[i] = buf[0]
					break
				}
			}
			time.Sleep(1 * time.Second)
		}

	}

	return readLen, nil
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
