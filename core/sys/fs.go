package sys

import (
	"io/fs"
	"os"
)

// FS An FS provides access to a hierarchical file system.
type FS interface {

	// Open opens the named file for reading. If successful, methods on
	// the returned file can be used for reading; the associated file
	// descriptor has mode O_RDONLY.
	// If there is an error, it will be of type *PathError.
	Open(name string) (File, error)

	// OpenFile open a file
	OpenFile(name string, flag int, perm os.FileMode) (File, error)

	// ReadFile reads the file named by filename and returns the contents.
	ReadFile(name string) ([]byte, error)

	// WriteFile writes data to a file named by filename.
	WriteFile(name string, data []byte, perm fs.FileMode) error

	Stat(name string) (fs.FileInfo, error)

	// Remove removes the named file or (empty) directory.
	// If there is an error, it will be of type *PathError.
	Remove(name string) error

	//MkdirAll creates a directory named path
	MkdirAll(path string, perm os.FileMode) error

	// LoadProgress load progress
	LoadProgress(progressID string) ([]byte, error)

	// SaveProgress save progress
	SaveProgress(progressID string, data []byte, perm fs.FileMode) error

	// RemoveProgress remove progress
	RemoveProgress(progressID string) error

	// Create Directory
	CreateDirectory(dirID string) error

	// GetFileHandler
	GetFileHandler(dirID, name string) (File, error)

	// Remove all created directories(used in download directory)
	RemoveAllDirectories()
}

type File interface {
	Stat() (fs.FileInfo, error)
	Read([]byte) (int, error)
	Write(p []byte) (n int, err error)

	Sync() error
	Seek(offset int64, whence int) (ret int64, err error)

	Close() error
}
