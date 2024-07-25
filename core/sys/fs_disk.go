package sys

import (
	"io/fs"
	"os"
)

// DiskFS implement file system on disk
type DiskFS struct {
}

// NewDiskFS create DiskFS instance
func NewDiskFS() FS {
	return &DiskFS{}
}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *PathError.
func (dfs *DiskFS) Open(name string) (File, error) {
	return dfs.OpenFile(name, os.O_RDONLY, 0)
}

func (dfs *DiskFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return os.OpenFile(name, flag, perm)
}

// ReadFile reads the file named by filename and returns the contents.
func (dfs *DiskFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// WriteFile writes data to a file named by filename.
func (dfs *DiskFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// Remove removes the named file or (empty) directory.
// If there is an error, it will be of type *PathError.
func (dfs *DiskFS) Remove(name string) error {
	return os.Remove(name)
}

// MkdirAll creates a directory named path
func (dfs *DiskFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Stat returns a FileInfo describing the named file.
// If there is an error, it will be of type *PathError.
func (dfs *DiskFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (dfs *DiskFS) LoadProgress(progressID string) ([]byte, error) {
	return dfs.ReadFile(progressID)
}

func (dfs *DiskFS) SaveProgress(progressID string, data []byte, perm fs.FileMode) error {
	return dfs.WriteFile(progressID, data, perm)
}

func (dfs *DiskFS) RemoveProgress(progressID string) error {
	return dfs.Remove(progressID)
}
