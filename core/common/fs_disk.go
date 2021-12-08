package common

import (
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
func (fs *DiskFS) Open(name string) (File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

func (fs *DiskFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return os.OpenFile(name, flag, perm)

}

// Remove removes the named file or (empty) directory.
// If there is an error, it will be of type *PathError.
func (fs *DiskFS) Remove(name string) error {
	return os.Remove(name)
}
