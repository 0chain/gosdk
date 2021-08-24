package sdk

import (
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

// FileMeta metadata of stream input/local
type FileMeta struct {
	// Mimetype mime type of source file
	MimeType string

	// Path local path of source file
	Path string
	// ThumbnailPath local path of source thumbnail
	ThumbnailPath string

	// ActualHash hash of orignial file (unencoded, unencrypted)
	ActualHash string
	// ActualSize total bytes of  orignial file (unencoded, unencrypted).  it is 0 if input is live stream.
	ActualSize int64
	// ActualThumbnailSize total bytes of orignial thumbnail (unencoded, unencrypted)
	ActualThumbnailSize int64
	// ActualThumbnailHash hash of orignial thumbnail (unencoded, unencrypted)
	ActualThumbnailHash string

	//RemoteName remote file name
	RemoteName string
	// RemotePath remote path
	RemotePath string
	// Attributes file attributes in blockchain
	Attributes fileref.Attributes
}

// FileID generante id of progress on local cache
func (meta *FileMeta) FileID() string {
	return encryption.Hash(meta.Path+"_"+meta.RemotePath) + "_" + meta.RemoteName
}

// UploadFormData form data of upload
type UploadFormData struct {
	ConnectionID string `json:"connection_id,omitempty"`
	// Filename remote file name
	Filename string `json:"filename,omitempty"`
	// Path remote path
	Path string `json:"filepath,omitempty"`

	// ContentHash hash of chunk data (encoded,encrypted)
	ContentHash string `json:"content_hash,omitempty"`
	// Hash hash of shard thumbnail  (encoded,encrypted)
	ThumbnailContentHash string `json:"thumbnail_content_hash,omitempty"`

	// MerkleRoot merkle's root hash of shard data (encoded, encrypted)
	MerkleRoot string `json:"merkle_root,omitempty"`

	// ActualHash hash of orignial file (unencoded, unencrypted)
	ActualHash string `json:"actual_hash,omitempty"`
	// ActualSize total bytes of  orignial file (unencoded, unencrypted)
	ActualSize int64 `json:"actual_size,omitempty"`
	// ActualThumbnailSize total bytes of orignial thumbnail (unencoded, unencrypted)
	ActualThumbSize int64 `json:"actual_thumb_size,omitempty"`
	// ActualThumbnailHash hash of orignial thumbnail (unencoded, unencrypted)
	ActualThumbHash string `json:"actual_thumb_hash,omitempty"`

	MimeType     string             `json:"mimetype,omitempty"`
	CustomMeta   string             `json:"custom_meta,omitempty"`
	EncryptedKey string             `json:"encrypted_key,omitempty"`
	Attributes   fileref.Attributes `json:"attributes,omitempty"`

	IsFinal      bool  `json:"is_final,omitempty"`      // current chunk is last or not
	ChunkIndex   int   `json:"chunk_index,omitempty"`   // the seq of current chunk. all chunks MUST be uploaded one by one because of streaming merkle hash
	ChunkSize    int64 `json:"chunk_size,omitempty"`    // the size of achunk. 64*1024 is default
	UploadOffset int64 `json:"upload_offset,omitempty"` // It is next position that new incoming chunk should be append to

}

// UploadProgress progress of upload
type UploadProgress struct {
	ID string `json:"id"`

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
	UploadLength int64 `json:"-"`

	Blobbers []*UploadBlobberStatus `json:"merkle_hashers,omitempty"`
}

// UploadBlobberStatus the status of blobber's upload progress
type UploadBlobberStatus struct {
	FixedMerkleTree *util.FixedMerkleTree `json:"trusted_content_hasher"`

	// ShardHasher hash.Hash `json:"-"`

	// UploadLength total bytes that has been uploaded to blobbers
	UploadLength int64 `json:"upload_length,omitempty"`
	// MerkleHasher a stateful stream merkle tree for uploaded chunks
	//MerkleHasher util.CompactMerkleTree `json:"merkle_hasher,omitempty"`
}

// getMerkelRoot see section 1.8 Oursourcing Attack Protection on Whitepaper
func (status *UploadBlobberStatus) getMerkelRoot() string {
	return status.FixedMerkleTree.GetMerkleRoot()
}
