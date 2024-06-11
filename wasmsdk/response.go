package main

import "github.com/0chain/gosdk/zboxcore/sdk"

type FileCommandResponse struct {
	CommandSuccess bool   `json:"commandSuccess,omitempty"`
	Error          string `json:"error,omitempty"`
}

type DownloadCommandResponse struct {
	CommandSuccess bool   `json:"commandSuccess,omitempty"`
	Error          string `json:"error,omitempty"`

	FileName string `json:"fileName,omitempty"`
	Url      string `json:"url,omitempty"`
}

type CheckStatusResult struct {
	Status        string              `json:"status"`
	Err           error               `json:"error"`
	BlobberStatus []sdk.BlobberStatus `json:"blobberStatus"`
}
