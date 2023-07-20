package main

type UploadFile struct {
	Name          string `json:"name,omitempty"`
	Path          string `json:"path,omitempty"`
	ThumbnailPath string `json:"thumbnailPath,omitempty"`

	RemotePath string `json:"remotePath,omitempty"`
	Encrypt    bool   `json:"encrypt,omitempty"`
	IsUpdate   bool   `json:"isUpdate,omitempty"`

	ChunkNumber int `json:"chunkNumber,omitempty"`
}
