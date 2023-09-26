package sdk

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/logger"
)

// Converting the video file to fmp4 format for web streaming
func TranscodeWebStreaming(workdir string, fileReader io.Reader, fileMeta FileMeta) (io.Reader, *FileMeta, string, error) {
	var stdErr bytes.Buffer

	outDir := filepath.Join(workdir, ".zcn", "transcode")
	// create ./zcn/transcode folder if it doesn't exists
	os.MkdirAll(outDir, 0644) //nolint: errcheck
	fileName := filepath.Join(outDir, fileMeta.RemoteName)

	logger.Logger.Info("transcode: ", fileName)

	args := []string{"-i", fileMeta.Path, "-f", "mp4", "-movflags", "frag_keyframe+empty_moov+default_base_moof", fileName, "-y"}
	cmd := exec.Command(CmdFFmpeg, args...)
	cmd.Stderr = bufio.NewWriter(&stdErr)
	cmd.SysProcAttr = sysProcAttr
	err := cmd.Run()
	defer cmd.Process.Kill()

	if err != nil {
		logger.Logger.Error(err, stdErr.String())
		return nil, nil, "", thrown.New("Transcoding Failed: ", err.Error())
	}

	// open file reader with readonly
	r, err := os.Open(fileName)

	if err != nil {
		return nil, nil, fileName, err
	}

	fi, err := r.Stat()
	if err != nil {
		return nil, nil, fileName, err
	}

	fm := &FileMeta{
		MimeType:            "video/mp4",
		Path:                fileMeta.Path,
		ThumbnailPath:       fileMeta.ThumbnailPath,
		ActualHash:          fileMeta.ActualHash,
		ActualSize:          fi.Size(),
		ActualThumbnailSize: fileMeta.ActualThumbnailSize,
		ActualThumbnailHash: fileMeta.ActualThumbnailHash,
		RemoteName:          fileMeta.RemoteName,
		RemotePath:          fileMeta.RemotePath,
	}

	return r, fm, fileName, nil
}

func getRemoteNameAndRemotePath(remoteName string, remotePath string) (string, string) {
	newRemotePath, newRemoteName := path.Split(remotePath)
	newRemoteNameSlice := strings.Split(newRemoteName, ".")
	if len(newRemoteNameSlice) > 0 {
		newRemoteNameSlice = newRemoteNameSlice[:len(newRemoteNameSlice)-1]
	}
	newRemoteNameWithoutType := strings.Join(newRemoteNameSlice, ".")
	newRemoteName = "raw." + newRemoteNameWithoutType + ".mp4"
	newRemotePath = newRemotePath + newRemoteName
	return newRemoteName, newRemotePath
}
