package common

import (
	"io/fs"
	"os"
)

// MemFS implement file system on memory
type MemFS struct {
	files map[string]MemFile
}

// NewMemFS create MemFS instance
func NewMemFS() FS {
	return &MemFS{
		files: make(map[string]MemFile),
	}
}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *PathError.
func (fs *MemFS) Open(name string) (File, error) {
	return nil, nil
}

func (fs *MemFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return nil, nil
}

// Remove removes the named file or (empty) directory.
// If there is an error, it will be of type *PathError.
func (fs *MemFS) Remove(name string) error {
	return os.Remove(name)
}

type MemFile struct {
}

func (f *MemFile) Stat() (fs.FileInfo, error) {
	return nil, nil
}
func (f *MemFile) Read([]byte) (int, error) {
	return 0, nil
}
func (f *MemFile) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (f *MemFile) Sync() error {
	return nil
}
func (f *MemFile) Seek(offset int64, whence int) (ret int64, err error) {
	return 0, nil
}

func (f *MemFile) Close() error {
	return nil
}
