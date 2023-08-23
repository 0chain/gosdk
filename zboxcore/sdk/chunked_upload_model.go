package sdk

import (
	"encoding/json"
	"hash/fnv"
	"strconv"

	"golang.org/x/crypto/sha3"
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

	// ValidationRoot is merkle root of sha256 of 64KB as leaf
	ValidationRoot          string `json:"validation_root,omitempty"`
	ValidationRootSignature string `json:"validation_root_signature,omitempty"`
	// Hash hash of shard thumbnail  (encoded,encrypted)
	ThumbnailContentHash string `json:"thumbnail_content_hash,omitempty"`

	// ChallengeHash challenge hash of shard data (encoded, encrypted)
	FixedMerkleRoot string `json:"fixed_merkle_root,omitempty"`

	// ActualHash hash of original file (un-encoded, un-encrypted)
	ActualHash              string `json:"actual_hash,omitempty"`
	ActualFileHashSignature string `json:"actual_file_hash_signature,omitempty"`
	// ActualSize total bytes of original file (un-encoded, un-encrypted)
	ActualSize int64 `json:"actual_size,omitempty"`
	// ActualThumbnailSize total bytes of original thumbnail (un-encoded, un-encrypted)
	ActualThumbSize int64 `json:"actual_thumb_size,omitempty"`
	// ActualThumbnailHash hash of original thumbnail (un-encoded, un-encrypted)
	ActualThumbHash string `json:"actual_thumb_hash,omitempty"`

	MimeType          string `json:"mimetype,omitempty"`
	CustomMeta        string `json:"custom_meta,omitempty"`
	EncryptedKey      string `json:"encrypted_key,omitempty"`
	EncryptedKeyPoint string `json:"encrypted_key_point,omitempty"`

	IsFinal         bool  `json:"is_final,omitempty"`          // all of chunks are uploaded
	ChunkStartIndex int   `json:"chunk_start_index,omitempty"` // start index of chunks.
	ChunkEndIndex   int   `json:"chunk_end_index,omitempty"`   // end index of chunks. all chunks MUST be uploaded one by one because of streaming merkle hash
	ChunkSize       int64 `json:"chunk_size,omitempty"`        // the size of a chunk. 64*1024 is default
	UploadOffset    int64 `json:"upload_offset,omitempty"`     // It is next position that new incoming chunk should be append to
	Size            int64 `json:"size"`                        // total size of shard

}

// UploadProgress progress of upload
type UploadProgress struct {
	ID string `json:"id"`

	// ChunkSize size of chunk
	ChunkSize int64 `json:"chunk_size,omitempty"`
	// EncryptOnUpload encrypt data on upload or not
	EncryptOnUpload   bool   `json:"is_encrypted,omitempty"`
	EncryptPrivateKey string `json:"-"`
	EncryptedKeyPoint string `json:"encrypted_key_point,omitempty"`

	// ConnectionID chunked upload connection_id
	ConnectionID string `json:"connection_id,omitempty"`
	// ChunkIndex index of last updated chunk
	ChunkIndex int `json:"chunk_index,omitempty"`
	// UploadLength total bytes that has been uploaded to blobbers
	UploadLength int64 `json:"-"`

	Blobbers []*UploadBlobberStatus `json:"-"`
}

// UploadBlobberStatus the status of blobber's upload progress
type UploadBlobberStatus struct {
	Hasher Hasher

	// UploadLength total bytes that has been uploaded to blobbers
	UploadLength int64 `json:"upload_length,omitempty"`
}

type status struct {
	Hasher       hasher
	UploadLength int64 `json:"upload_length,omitempty"`
}

func (s *UploadBlobberStatus) UnmarshalJSON(b []byte) error {
	if s == nil {
		return nil
	}
	//fixed Hasher doesn't work in UnmarshalJSON
	status := &status{}

	if err := json.Unmarshal(b, status); err != nil {
		return err
	}

	status.Hasher.File = sha3.New256()

	s.Hasher = &status.Hasher
	s.UploadLength = status.UploadLength

	return nil
}

type blobberShards [][]byte

// batchChunksData chunks data
type batchChunksData struct {
	// chunkStartIndex start index of chunks
	chunkStartIndex int
	// chunkEndIndex end index of chunks
	chunkEndIndex int
	// isFinal last chunk or not
	isFinal bool
	// ReadSize total size read from original reader (un-encoded, un-encrypted)
	totalReadSize int64
	// FragmentSize total fragment size for a blobber (un-encrypted)
	totalFragmentSize int64

	fileShards      []blobberShards
	thumbnailShards blobberShards
}
