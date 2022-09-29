package main

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
