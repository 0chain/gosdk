package sdk

import (
	"net/url"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

// FileMeta metadata of stream input/local
type FileMeta struct {
	// Path local path of source file
	Path string
	// Size total bytes of source file. it is 0 if input is live stream.
	Size int64
	// Mimetype mime type of source file
	MimeType string

	// ThumbnailPath local path of source thumbnail
	ThumbnailPath string
	// Size total bytes of source thumbnail
	ThumbnailSize int

	//RemoteName remote file name
	RemoteName string
	// RemotePath remote path
	RemotePath string
	// Attributes file attributes in blockchain
	Attributes fileref.Attributes
}

// UploadFormData form data of upload
type UploadFormData struct {

	// Name remote file name
	Name string `json:"name,omitempty"`
	// Path remote path
	Path string `json:"path,omitempty"`
	// MimeType the mime type of source file
	MimeType string `json:"mime_type,omitempty"`

	// Hash hash of current uploadFormFile
	Hash string `json:"hash,omitempty"`
	// Size total bytes of current uploadFormFile
	Size int64 `json:"size,omitempty"`
	// Hash hash of current uploadThumbnail
	ThumbnailHash string `json:"thumbnail_hash,omitempty"`
	// ThumbnailSize total bytes of current uploadThumbnail
	ThumbnailSize int `json:"thumbnail_size,omitempty"`

	// ActualHash merkle's root hash of a shard includes all chunks with encryption. it is only set in last chunk
	ActualHash string `json:"actual_hash,omitempty"`
	// ActualSize total bytes of a shard includes all chunks with encryption.
	ActualSize int64 `json:"actual_size,omitempty"`

	// AllocationID id of allocation
	AllocationID string `json:"allocation_id,omitempty"`
	// ConnectionID the connection_is used in resumable upload
	ConnectionID string `json:"connection_id,omitempty"`
	// CustomMeta custom meta in blockchain
	CustomMeta string `json:"custom_meta,omitempty"`
	//
	EncryptedKey string             `json:"encrypted_key,omitempty"`
	Attributes   fileref.Attributes `json:"attributes,omitempty"`

	IsFinal      bool  `json:"is_final,omitempty"`      // current chunk is last or not
	ChunkIndex   int   `json:"chunk_index,omitempty"`   // the seq of current chunk. all chunks MUST be uploaded one by one because of streaming merkle hash
	UploadOffset int64 `json:"upload_offset,omitempty"` // It is next position that new incoming chunk should be append to

}

// FileID generante id of progress on local cache
func (meta *FileMeta) FileID() string {
	return url.PathEscape(meta.Path) + "_" + url.PathEscape(meta.RemotePath)
}

// UploadProgress progress of upload
type UploadProgress struct {

	// ChunkSize size of chunk
	ChunkSize int `json:"chunk_size,omitempty"`
	// EncryptOnUpload encrypt data on upload or not
	EncryptOnUpload  bool `json:"is_encrypted,omitempty"`
	EncryptPrivteKey string

	// ConnectionID resumable upload connection_id
	ConnectionID string `json:"connection_id,omitempty"`
	// ChunkIndex index of last updated chunk
	ChunkIndex int `json:"chunk_index,omitempty"`
	// UploadLength total bytes that has been uploaded to blobbers
	UploadLength int64 `json:"upload_length,omitempty"`

	Blobbers []*UploadBlobberStatus `json:"merkle_hashers,omitempty"`
}

// UploadBlobberStatus the status of blobber's upload progress
type UploadBlobberStatus struct {
	// ActualSize total bytes of shard includes all chunks with encryption.
	ActualSize int64 `json:"actual_size,omitempty"`

	// MerkleHasher a stateful stream merkle tree for uploaded chunks
	MerkleHasher util.StreamMerkleHasher
}
