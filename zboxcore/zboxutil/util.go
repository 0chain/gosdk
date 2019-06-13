package zboxutil

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
)

func GetFileContentType(out *os.File) (string, error) {

	buffer := make([]byte, 261)
	_, err := out.Read(buffer)

	if err != nil {
		return "", err
	}

	kind, _ := filetype.Match(buffer)
	if kind == filetype.Unknown {
		return "application/octet-stream", nil
	}
	out.Seek(0, 0)
	return kind.MIME.Value, nil

	// // Only the first 512 bytes are used to sniff the content type.
	// buffer := make([]byte, 512)

	// _, err := out.Read(buffer)
	// if err != nil {
	// 	return "", err
	// }

	// // Use the net/http package's handy DectectContentType function. Always returns a valid
	// // content-type by returning "application/octet-stream" if no others seemed to match.
	// contentType := http.DetectContentType(buffer)
	// fmt.Println("Found content type : " + contentType)
	// out.Seek(0, 0)

	// return contentType, nil
}

func GetFullRemotePath(localPath, remotePath string) string {
	if remotePath == "" || strings.HasSuffix(remotePath, "/") {
		remotePath = strings.TrimRight(remotePath, "/")
		_, fileName := filepath.Split(localPath)
		remotePath = fmt.Sprintf("%s/%s", remotePath, fileName)
	}
	return remotePath
}

func NewConnectionId() string {
	nBig, err := rand.Int(rand.Reader, big.NewInt(0xffffffff))
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%d", nBig.Int64())
}
