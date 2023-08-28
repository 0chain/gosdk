package main

type UploadFile struct {
	Name          string
	Path          string
	ThumbnailPath string

	RemotePath     string
	Encrypt        bool
	IsUpdate       bool
	IsWebstreaming bool

	ChunkNumber int
}
