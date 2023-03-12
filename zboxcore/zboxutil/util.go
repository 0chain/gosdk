package zboxutil

import (
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"math/bits"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"errors"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/h2non/filetype"
	"github.com/lithammer/shortuuid/v3"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/scrypt"
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

const (
	keySize      = 32
	nonceSize    = 12
	saltSize     = 32
	tagSize      = 16
	scryptN      = 32768
	scryptR      = 8
	scryptP      = 1
	scryptKeyLen = 32
)

func Encrypt(key, text []byte) ([]byte, error) {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	derivedKey, err := scrypt.Key(key, salt, scryptN, scryptR, scryptP, scryptKeyLen)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	aead, err := chacha20poly1305.New(derivedKey)
	if err != nil {
		return nil, err
	}

	ciphertext := aead.Seal(nil, nonce, text, nil)
	ciphertext = append(salt, ciphertext...)
	ciphertext = append(nonce, ciphertext...)

	return ciphertext, nil
}

func Decrypt(key, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < saltSize+nonceSize+tagSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	salt := ciphertext[nonceSize : nonceSize+saltSize]
	text := ciphertext[saltSize+nonceSize:]

	derivedKey, err := scrypt.Key(key, salt, scryptN, scryptR, scryptP, scryptKeyLen)
	if err != nil {
		return nil, err
	}
	aead, err := chacha20poly1305.New(derivedKey)
	if err != nil {
		return nil, err
	}
	plaintext, err := aead.Open(nil, nonce, text, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
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

func GetActiveBlobbers(dirMask uint32, blobbers []*blockchain.StorageNode) []*blockchain.StorageNode {
	var c, pos int
	var r []*blockchain.StorageNode
	for i := dirMask; i != 0; i &= ^(1 << pos) {
		pos = bits.TrailingZeros32(i)
		r = append(r, blobbers[pos])
		c++
	}

	return r
}

func GetRateLimitValue(r *http.Response) (int, error) {
	rlStr := r.Header.Get("X-Rate-Limit-Limit")
	durStr := r.Header.Get("X-Rate-Limit-Duration")

	rl, err := strconv.ParseFloat(rlStr, 64)
	if err != nil {
		return 0, err
	}

	dur, err := strconv.ParseFloat(durStr, 64)
	if err != nil {
		return 0, err
	}

	return int(math.Ceil(rl / dur)), nil
}
