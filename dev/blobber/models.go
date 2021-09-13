package blobber

import (
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/fileref"
)

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

	// MerkleRoot challenge hash of shard data (encoded, encrypted)
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

	IsFinal      bool   `json:"is_final,omitempty"`      // current chunk is last or not
	ChunkHash    string `json:"chunk_hash"`              // hash of current chunk
	ChunkIndex   int    `json:"chunk_index,omitempty"`   // the seq of current chunk. all chunks MUST be uploaded one by one because of streaming merkle hash
	ChunkSize    int64  `json:"chunk_size,omitempty"`    // the size of a chunk. 64*1024 is default
	UploadOffset int64  `json:"upload_offset,omitempty"` // It is next position that new incoming chunk should be append to

}

type UploadResult struct {
	Filename   string `json:"filename"`
	ShardSize  int64  `json:"size"`
	Hash       string `json:"content_hash,omitempty"`
	MerkleRoot string `json:"merkle_root,omitempty"`
}

type WriteMarker struct {
	AllocationRoot         string           `gorm:"column:allocation_root;primary_key" json:"allocation_root"`
	PreviousAllocationRoot string           `gorm:"column:prev_allocation_root" json:"prev_allocation_root"`
	AllocationID           string           `gorm:"column:allocation_id" json:"allocation_id"`
	Size                   int64            `gorm:"column:size" json:"size"`
	BlobberID              string           `gorm:"column:blobber_id" json:"blobber_id"`
	Timestamp              common.Timestamp `gorm:"column:timestamp" json:"timestamp"`
	ClientID               string           `gorm:"column:client_id" json:"client_id"`
	Signature              string           `gorm:"column:signature" json:"signature"`
}

type CommitResult struct {
	AllocationRoot string       `json:"allocation_root"`
	WriteMarker    *WriteMarker `json:"write_marker"`
	Success        bool         `json:"success"`
	ErrorMessage   string       `json:"error_msg,omitempty"`
	//	Changes        []*allocation.AllocationChange `json:"-"`
	//Result         []*UploadResult         `json:"result"`
}
