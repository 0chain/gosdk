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

type UploadResult struct {
	Filename   string `json:"filename"`
	ShardSize  int64  `json:"size"`
	Hash       string `json:"content_hash,omitempty"`
	MerkleRoot string `json:"merkle_root,omitempty"`
}
