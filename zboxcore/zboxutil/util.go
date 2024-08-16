package zboxutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/bits"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"errors"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/zboxcore/allocationchange"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/h2non/filetype"
	"github.com/hitenjain14/fasthttp"
	"github.com/lithammer/shortuuid/v3"
	"github.com/minio/sha256-simd"
	"github.com/valyala/bytebufferpool"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/scrypt"
)

const EncryptedFolderName = "encrypted"

var BufferPool bytebufferpool.Pool

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

// GetFileContentType returns the content type of the file based on reading the first 10KB of the file
//   - ext is the extension of the file, shouldn't be empty
//   - out is the file content
func GetFileContentType(ext string, out io.ReadSeeker) (string, error) {

	if ext != "" {
		if content, ok := mimeDB[strings.TrimPrefix(ext, ".")]; ok {
			return content.ContentType, nil
		}
	}

	buffer := make([]byte, 10240)
	n, err := out.Read(buffer)
	defer out.Seek(0, 0) //nolint

	if err != nil && err != io.EOF {
		return "", err
	}
	buffer = buffer[:n]

	kind, _ := filetype.Match(buffer)
	if kind == filetype.Unknown {
		return "application/octet-stream", nil
	}

	return kind.MIME.Value, nil
}

// GetFullRemotePath returns the full remote path by combining the local path and remote path
//   - localPath is the local path of the file
//   - remotePath is the remote path of the file
func GetFullRemotePath(localPath, remotePath string) string {
	if remotePath == "" || strings.HasSuffix(remotePath, "/") {
		remotePath = strings.TrimRight(remotePath, "/")
		_, fileName := filepath.Split(localPath)
		remotePath = fmt.Sprintf("%s/%s", remotePath, fileName)
	}
	return remotePath
}

// NewConnectionId generate new connection id.
// Connection is used to track the upload/download progress and redeem the cost of the operation from the network.
// It's in the short uuid format. Check here for more on this format: https://pkg.go.dev/github.com/lithammer/shortuuid/v3@v3.0.7
func NewConnectionId() string {
	return shortuuid.New()
}

// IsRemoteAbs returns true if the path is remote absolute path
//   - path is the path to check
func IsRemoteAbs(path string) bool {
	return strings.HasPrefix(path, "/")
}

// RemoteClean returns the cleaned remote path
//   - path is the path to clean
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

func GetRefsHash(r []byte) string {
	hash := sha256.New()
	hash.Write(r)
	var buf []byte
	buf = hash.Sum(buf)
	return hex.EncodeToString(buf)
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

func GetFastRateLimitValue(r *fasthttp.Response) (int, error) {
	rlStr := r.Header.Peek("X-Rate-Limit-Limit")
	durStr := r.Header.Peek("X-Rate-Limit-Duration")

	rl, err := strconv.ParseFloat(string(rlStr), 64)
	if err != nil {
		return 0, err
	}

	dur, err := strconv.ParseFloat(string(durStr), 64)
	if err != nil {
		return 0, err
	}

	return int(math.Ceil(rl / dur)), nil
}

func MajorError(errors []error) error {
	countError := make(map[error]int)
	for _, value := range errors {
		if value != nil {
			countError[value] += 1
		}
	}
	maxFreq := 0
	var maxKey error
	for key, value := range countError {
		if value > maxFreq {
			maxKey = key
			maxFreq = value
		}
	}
	return maxKey
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

func ScryptEncrypt(key, text []byte) ([]byte, error) {
	if len(key) != keySize {
		return nil, errors.New("scrypt: invalid key size" + strconv.Itoa(len(key)))
	}
	if len(text) == 0 {
		return nil, errors.New("scrypt: plaintext cannot be empty")
	}
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

func ScryptDecrypt(key, ciphertext []byte) ([]byte, error) {
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

// Returns the error message code, message should be strictly of the
// format: ".... err: {"code" : <return_this>, ...}, ..."
func GetErrorMessageCode(errorMsg string) (string, error) {
	// find index of "err"
	targetWord := `err:`
	idx := strings.Index(errorMsg, targetWord)
	if idx == -1 {
		return "", thrown.New("invalid_params", "message doesn't contain `err` field")

	}
	var a = make(map[string]string)
	if idx+5 >= len(errorMsg) {
		return "", thrown.New("invalid_format", "err field is not proper json")
	}
	err := json.Unmarshal([]byte(errorMsg[idx+5:]), &a)
	if err != nil {
		return "", thrown.New("invalid_format", "err field is not proper json")
	}
	return a["code"], nil

}

// Returns transpose of 2-D slice
// Example: Given matrix [[a, b], [c, d], [e, f]] returns [[a, c, e], [b, d, f]]
func Transpose(matrix [][]allocationchange.AllocationChange) [][]allocationchange.AllocationChange {
	rowLength := len(matrix)
	if rowLength == 0 {
		return matrix
	}
	columnLength := len(matrix[0])
	transposedMatrix := make([][]allocationchange.AllocationChange, columnLength)
	for i := range transposedMatrix {
		transposedMatrix[i] = make([]allocationchange.AllocationChange, rowLength)
	}
	for i := 0; i < columnLength; i++ {
		for j := 0; j < rowLength; j++ {
			transposedMatrix[i][j] = matrix[j][i]
		}
	}
	return transposedMatrix
}
