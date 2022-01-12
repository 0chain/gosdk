package zboxutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"

	"errors"

	"github.com/h2non/filetype"
	"github.com/lithammer/shortuuid/v3"
	"golang.org/x/crypto/sha3"
)

const EncryptedFolderName = "encrypted"

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

func GetFileContentType(out io.ReadSeeker) (string, error) {
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

// NewConnectionId generate new connection id
func NewConnectionId() string {
	return shortuuid.New()
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
		case path[r] == '/' || path[r] == '\\': //os.IsPathSeparator(path[r]):
			// empty path element
			r++
		case path[r] == '.' && (r+1 == n || path[r+1] == '/'): //os.IsPathSeparator(path[r+1])):
			// . element
			r++
		case path[r] == '.' && path[r+1] == '.' && (r+2 == n || path[r+2] == '/'): //os.IsPathSeparator(path[r+2])):
			// .. element: remove to last separator
			r += 2
			switch {
			case out.w > dotdot:
				// can backtrack
				out.w--
				for out.w > dotdot && !((out.index(out.w)) == '/') { //!os.IsPathSeparator(out.index(out.w)) {
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
			for ; r < n && !(path[r] == '/' || path[r] == '\\'); r++ { //!os.IsPathSeparator(path[r]); r++ {
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

func Encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

func Decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func calculateMinRequired(minRequired, percent float64) int {
	return int(math.Ceil(minRequired * percent))
}

func Join(a, b string) string {
	return strings.ReplaceAll(filepath.Join(a, b), "\\", "/")
}

func GetRefsHash(r []byte) string {
	hash := sha3.New256()
	hash.Write(r)
	var buf []byte
	buf = hash.Sum(buf)
	return string(buf)
}
