package sdk

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path"
	"strings"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/logger"
)

// Converting the video file to fmp4 format for web streaming
func TranscodeWebStreaming(fileReader io.Reader, fileMeta FileMeta) (io.Reader, *FileMeta, error) {

	mimeTypeSlice := strings.Split(fileMeta.MimeType, "/")
	if mimeTypeSlice[0] != "video" {
		return nil, nil, thrown.New("Transcoding Failed", fmt.Sprintf("Format Invalid %s", fileMeta.MimeType))
	}

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	args := []string{"-i", "-", "-g", "30", "-f", "mp4", "-movflags", "frag_keyframe+empty_moov", "pipe:1"}
	cmdFfmpeg := exec.Command("ffmpeg", args...)

	cmdFfmpeg.Stdin = fileReader
	cmdFfmpeg.Stdout = bufio.NewWriter(&stdOut)
	cmdFfmpeg.Stderr = bufio.NewWriter(&stdErr)

	err := cmdFfmpeg.Run()

	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, err
	}

	trascodedBufSlice := stdOut.Bytes()
	transcodedFileReader := bytes.NewReader(trascodedBufSlice)

	remoteName, remotePath := getRemoteNameAndRemotePath(fileMeta.RemoteName, fileMeta.RemotePath)

	transcodedFileMeta := &FileMeta{
		MimeType:            "video/fmp4",
		Path:                fileMeta.Path,
		ThumbnailPath:       fileMeta.ThumbnailPath,
		ActualHash:          fileMeta.ActualHash,
		ActualSize:          int64(len(trascodedBufSlice)),
		ActualThumbnailSize: fileMeta.ActualThumbnailSize,
		ActualThumbnailHash: fileMeta.ActualThumbnailHash,
		RemoteName:          remoteName,
		RemotePath:          remotePath,
	}

	return transcodedFileReader, transcodedFileMeta, nil
}

func getRemoteNameAndRemotePath(remoteName string, remotePath string) (string, string) {
	newRemotePath, newRemoteName := path.Split(remotePath)
	newRemoteNameSlice := strings.Split(newRemoteName, ".")
	if len(newRemoteNameSlice) > 0 {
		newRemoteNameSlice = newRemoteNameSlice[:len(newRemoteNameSlice)-1]
	}
	newRemoteNameWithoutType := strings.Join(newRemoteNameSlice, ".")
	newRemoteName = "raw." + newRemoteNameWithoutType + ".fmp4"
	newRemotePath = newRemotePath + newRemoteName
	return newRemoteName, newRemotePath
}

