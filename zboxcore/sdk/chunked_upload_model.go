package sdk

import (
	"hash/fnv"
	"strconv"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/rogpeppe/go-internal/lockedfile"
)

// FileMeta metadata of stream input/local
type FileMeta struct {
	// Mimetype mime type of source file
	MimeType string

	// Path local path of source file
	Path string
	// ThumbnailPath local path of source thumbnail
	ThumbnailPath string

	// ActualHash hash of original file (un-encoded, un-encrypted)
	ActualHash string
	// ActualSize total bytes of  original file (unencoded, un-encrypted).  it is 0 if input is live stream.
	ActualSize int64
	// ActualThumbnailSize total bytes of original thumbnail (un-encoded, un-encrypted)
	ActualThumbnailSize int64
	// ActualThumbnailHash hash of original thumbnail (un-encoded, un-encrypted)
	ActualThumbnailHash string

	//RemoteName remote file name
	RemoteName string
	// RemotePath remote path
	RemotePath string
	// Attributes file attributes in blockchain
	Attributes fileref.Attributes
}

// FileID generate id of progress on local cache
func (meta *FileMeta) FileID() string {

	hash := fnv.New64a()
	hash.Write([]byte(meta.Path + "_" + meta.RemotePath))

	return strconv.FormatUint(hash.Sum64(), 36) + "_" + meta.RemoteName
}

// UploadFormData form data of upload
type UploadFormData struct {
	ConnectionID string `json:"connection_id,omitempty"`
	// Filename remote file name
	Filename string `json:"filename,omitempty"`
	// Path remote path
	Path string `json:"filepath,omitempty"`

	// ContentHash hash of shard data (encoded,encrypted) when it is last chunk. it is ChunkHash if it is not last chunk.
	ContentHash string `json:"content_hash,omitempty"`
	// Hash hash of shard thumbnail  (encoded,encrypted)
	ThumbnailContentHash string `json:"thumbnail_content_hash,omitempty"`

	// ChallengeHash challenge hash of shard data (encoded, encrypted)
	ChallengeHash string `json:"merkle_root,omitempty"`

	// ActualHash hash of original file (un-encoded, un-encrypted)
	ActualHash string `json:"actual_hash,omitempty"`
	// ActualSize total bytes of original file (un-encoded, un-encrypted)
	ActualSize int64 `json:"actual_size,omitempty"`
	// ActualThumbnailSize total bytes of original thumbnail (un-encoded, un-encrypted)
	ActualThumbSize int64 `json:"actual_thumb_size,omitempty"`
	// ActualThumbnailHash hash of original thumbnail (un-encoded, un-encrypted)
	ActualThumbHash string `json:"actual_thumb_hash,omitempty"`

	MimeType     string             `json:"mimetype,omitempty"`
	CustomMeta   string             `json:"custom_meta,omitempty"`
	EncryptedKey string             `json:"encrypted_key,omitempty"`
	Attributes   fileref.Attributes `json:"attributes,omitempty"`

	IsFinal      bool   `json:"is_final,omitempty"`      // current chunk is last or not
	ChunkHash    string `json:"chunk_hash"`              // hash of current chunk
	ChunkIndex   int    `json:"chunk_index,omitempty"`   // the seq of current chunk. all chunks MUST be uploaded one by one because of streaming merkle hash
	ChunkSize    int64  `json:"chunk_size,omitempty"`    // the size of a chunk. 64*1024 is default
	UploadOffset int64  `json:"upload_offset,omitempty"` // It is next position that new incoming chunk should be append to

}

// UploadProgress progress of upload
type UploadProgress struct {
	ID string `json:"id"`

	// ChunkSize size of chunk
	ChunkSize int64 `json:"chunk_size,omitempty"`
	// EncryptOnUpload encrypt data on upload or not
	EncryptOnUpload   bool   `json:"is_encrypted,omitempty"`
	EncryptPrivateKey string `json:"-"`

	// ConnectionID chunked upload connection_id
	ConnectionID string `json:"connection_id,omitempty"`
	// ChunkIndex index of last updated chunk
	ChunkIndex int `json:"chunk_index,omitempty"`
	// UploadLength total bytes that has been uploaded to blobbers
	UploadLength int64 `json:"-"`

	Blobbers []*UploadBlobberStatus `json:"merkle_hashers,omitempty"`
}

// UploadBlobberStatus the status of blobber's upload progress
type UploadBlobberStatus struct {
	Hasher Hasher

	// UploadLength total bytes that has been uploaded to blobbers
	UploadLength int64 `json:"upload_length,omitempty"`
}

// TODO: copy lockedfile from https://cs.opensource.google/go/go/+/refs/tags/go1.17.1:src/cmd/go/internal/lockedfile/internal/filelock/
// see more detail on

// - https://github.com/golang/go/issues/33974
// - https://go.googlesource.com/proposal/+/master/design/33974-add-public-lockedfile-pkg.md

// We should replaced it with official package if it is released as public
type FLock struct {
	sync.Mutex
	file string

	fileMutex  *lockedfile.Mutex
	fileUnlock func()
}

func createFLock(file string) *FLock {
	return &FLock{
		file: file,
	}
}

func (f *FLock) Lock() error {
	if f == nil {
		return errors.Throw(constants.ErrInvalidParameter, "f")
	}

	f.Mutex.Lock()
	defer f.Mutex.Unlock()

	if f.fileMutex == nil {
		// // open a new os.File instance
		// // create it if it doesn't exist, and open the file read-only.
		// flags := os.O_CREATE
		// if runtime.GOOS == "aix" {
		// 	// AIX cannot preform write-lock (ie exclusive) on a
		// 	// read-only file.
		// 	flags |= os.O_RDWR
		// } else {
		// 	flags |= os.O_RDONLY
		// }
		// fh, err := os.OpenFile(f.file, flags, os.FileMode(0600))
		// if err != nil {
		// 	return err
		// }

		// f.fh = fh
		f.fileMutex = lockedfile.MutexAt(f.file)
	}

	fileUnlock, err := f.fileMutex.Lock()
	if err != nil {
		return err
	}

	f.fileUnlock = fileUnlock

	return nil
}

func (f *FLock) Unlock() {

	if f.fileUnlock != nil {
		f.fileUnlock()
	}

	f.fileUnlock = nil
}
