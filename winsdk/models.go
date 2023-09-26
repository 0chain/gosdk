//go:build windows
// +build windows

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

type SharedInfo struct {
	AllocationID string
	LookupHash   string
}
