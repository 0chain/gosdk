package model

// UploadFormData form data of upload
type UploadFormData struct {
	ConnectionID string `json:"connection_id,omitempty"`
	// Filename remote file name
	Filename string `json:"filename,omitempty"`
	// Path remote path
	Path string `json:"filepath,omitempty"`

	// ValidationRoot of shard data (encoded,encrypted) where leaf is sha256 hash of 64KB data
	ValidationRoot string `json:"validation_root,omitempty"`
	// Hash hash of shard thumbnail  (encoded,encrypted)
	ThumbnailContentHash string `json:"thumbnail_content_hash,omitempty"`

	// FixedMerkleRoot merkle root of shard data (encoded, encrypted)
	FixedMerkleRoot string `json:"fixed_merkle_root,omitempty"`

	// ActualHash hash of orignial file (unencoded, unencrypted)
	ActualHash string `json:"actual_hash,omitempty"`
	// ActualSize total bytes of  orignial file (unencoded, unencrypted)
	ActualSize int64 `json:"actual_size,omitempty"`
	// ActualThumbnailSize total bytes of orignial thumbnail (unencoded, unencrypted)
	ActualThumbSize int64 `json:"actual_thumb_size,omitempty"`
	// ActualThumbnailHash hash of orignial thumbnail (unencoded, unencrypted)
	ActualThumbHash string `json:"actual_thumb_hash,omitempty"`

	MimeType     string `json:"mimetype,omitempty"`
	CustomMeta   string `json:"custom_meta,omitempty"`
	EncryptedKey string `json:"encrypted_key,omitempty"`

	IsFinal      bool   `json:"is_final,omitempty"`      // current chunk is last or not
	ChunkHash    string `json:"chunk_hash"`              // hash of current chunk
	ChunkIndex   int    `json:"chunk_index,omitempty"`   // the seq of current chunk. all chunks MUST be uploaded one by one because of streaming merkle hash
	ChunkSize    int64  `json:"chunk_size,omitempty"`    // the size of a chunk. 64*1024 is default
	UploadOffset int64  `json:"upload_offset,omitempty"` // It is next position that new incoming chunk should be append to

}
