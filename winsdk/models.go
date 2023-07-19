package main

import "sync"

type StatusCallback struct {
	sync.Mutex
}

func (c *StatusCallback) Started(allocationId, filePath string, op int, totalBytes int) {
	//c.Callback.Started(allocationId, filePath, op, totalBytes)
}

func (c *StatusCallback) InProgress(allocationId, filePath string, op int, completedBytes int, data []byte) {
	//c.Callback.InProgress(allocationId, filePath, op, completedBytes, data)
}

func (c *StatusCallback) Error(allocationID string, filePath string, op int, err error) {
	//c.Callback.Error(allocationID, filePath, op, err)
}

func (c *StatusCallback) Completed(allocationId, filePath string, filename string, mimetype string, size int, op int) {
	//c.Callback.Completed(allocationId, filePath, filename, mimetype, size, op)
}

func (c *StatusCallback) CommitMetaCompleted(request, response string, err error) {
	//c.Callback.CommitMetaCompleted(request, response, err)
}

func (c *StatusCallback) RepairCompleted(filesRepaired int) {
	//c.Callback.RepairCompleted(filesRepaired)
}

type UploadFile struct {
	Name          string `json:"fileName,omitempty"`
	Path          string `json:"filePath,omitempty"`
	ThumbnailPath string `json:"thumbnailPath,omitempty"`

	RemotePath string `json:"remotePath,omitempty"`
	Encrypt    bool   `json:"encrypt,omitempty"`
	IsUpdate   bool   `json:"isUpdate,omitempty"`

	ChunkNumber int `json:"chunkNumber,omitempty"`
}
