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

type lazybuf struct {
	path       string
	buf        []byte
	w          int
	volAndPath string
	volLen     int
}

func (b *lazybuf) index(i int) byte {
	if b.buf != nil {
		return b.buf[i]
	}
	return b.path[i]
}

func (b *lazybuf) append(c byte) {
	if b.buf == nil {
		if b.w < len(b.path) && b.path[b.w] == c {
			b.w++
			return
		}
		b.buf = make([]byte, len(b.path))
		copy(b.buf, b.path[:b.w])
	}
	b.buf[b.w] = c
	b.w++
}

func (b *lazybuf) string() string {
	if b.buf == nil {
		return b.volAndPath[:b.volLen+b.w]
	}
	return b.volAndPath[:b.volLen] + string(b.buf[:b.w])
}

func GetFileContentType(out *os.File) (string, error) {

	buffer := make([]byte, 261)
	_, err := out.Read(buffer)
	defer out.Seek(0, 0)
	if err != nil {
		return "", err
	}

	kind, _ := filetype.Match(buffer)
	if kind == filetype.Unknown {
		return "application/octet-stream", nil
	}

	return kind.MIME.Value, nil
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

func IsRemoteAbs(path string) bool {
	return strings.HasPrefix(path, "/")
}

func RemoteClean(path string) string {
	originalPath := path
	volLen := 0 //volumeNameLen(path)
	path = path[volLen:]
	if path == "" {
		if volLen > 1 && originalPath[1] != ':' {
			// should be UNC
			return path //FromSlash(originalPath)
		}
		return originalPath + "."
	}
	rooted := path[0] == '/' //os.IsPathSeparator(path[0])
	// Invariants:
	//	reading from path; r is index of next byte to process.
	//	writing to buf; w is index of next byte to write.
	//	dotdot is index in buf where .. must stop, either because
	//		it is the leading slash or it is a leading ../../.. prefix.
	n := len(path)
	out := lazybuf{path: path, volAndPath: originalPath, volLen: volLen}
	r, dotdot := 0, 0
	if rooted {
		out.append('/') //(Separator)
		r, dotdot = 1, 1
	}
	for r < n {
		switch {
		case path[r] == '/': //os.IsPathSeparator(path[r]):
			// empty path element
			r++
		case path[r] == '.' && (r+1 == n || path[r+1]=='/'): //os.IsPathSeparator(path[r+1])):
			// . element
			r++
		case path[r] == '.' && path[r+1] == '.' && (r+2 == n || path[r+2]=='/'): //os.IsPathSeparator(path[r+2])):
			// .. element: remove to last separator
			r += 2
			switch {
			case out.w > dotdot:
				// can backtrack
				out.w--
				for out.w > dotdot && !((out.index(out.w))=='/') { //!os.IsPathSeparator(out.index(out.w)) {
					out.w--
				}
			case !rooted:
				// cannot backtrack, but not rooted, so append .. element.
				if out.w > 0 {
					out.append('/') //Separator)
				}
				out.append('.')
				out.append('.')
				dotdot = out.w
			}
		default:
			// real path element.
			// add slash if needed
			if rooted && out.w != 1 || !rooted && out.w != 0 {
				out.append('/') //(Separator)
			}
			// copy element
			for ; r < n && !(path[r]=='/'); r++ { //!os.IsPathSeparator(path[r]); r++ {
				out.append(path[r])
			}
		}
	}
	// Turn empty string into "."
	if out.w == 0 {
		out.append('.')
	}
	return out.string() //(FromSlash(out.string())
}
