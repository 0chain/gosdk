package sdk

type UploadFileMeta struct {
	// Name remote file name
	Name string
	// Path remote path
	Path string
	// Hash hash of entire source file
	Hash     string
	MimeType string
	// Size total bytes of entire source file
	Size int64

	// ThumbnailSize total bytes of entire thumbnail
	ThumbnailSize int64
	// ThumbnailHash hash code of entire thumbnail
	ThumbnailHash string
}

type uploadFormData struct {
	ConnectionID string `json:"connection_id"`
	// Filename remote file name
	Filename string `json:"filename"`
	// Path remote path
	Path string `json:"filepath"`

	// Hash hash of shard data (encoded, encrypted)
	Hash string `json:"content_hash,omitempty"`
	// Hash hash of shard thumbnail (encoded, encrypted)
	ThumbnailHash string `json:"thumbnail_content_hash,omitempty"`

	// MerkleRoot merkle's root hash of shard data (encoded, encrypted)
	MerkleRoot string `json:"merkle_root,omitempty"`

	// ActualHash hash of orignial file (unencoded, unencrypted)
	ActualHash string `json:"actual_hash"`
	// ActualSize total bytes of orignial file (unencoded, unencrypted)
	ActualSize int64 `json:"actual_size"`
	// ActualThumbnailSize total bytes of orignial thumbnail (unencoded, unencrypted)
	ActualThumbnailSize int64 `json:"actual_thumb_size"`
	// ActualThumbnailHash hash of orignial thumbnail (unencoded, unencrypted)
	ActualThumbnailHash string `json:"actual_thumb_hash"`

	MimeType     string `json:"mimetype"`
	CustomMeta   string `json:"custom_meta,omitempty"`
	EncryptedKey string `json:"encrypted_key,omitempty"`
}

type UploadResult struct {
	Filename   string `json:"filename"`
	ShardSize  int64  `json:"size"`
	Hash       string `json:"content_hash,omitempty"`
	MerkleRoot string `json:"merkle_root,omitempty"`
}
