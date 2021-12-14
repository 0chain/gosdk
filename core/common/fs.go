package common

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

	// Remove removes the named file or (empty) directory.
	// If there is an error, it will be of type *PathError.
	Remove(name string) error
}

type File interface {
	Stat() (fs.FileInfo, error)
	Read([]byte) (int, error)
	Write(p []byte) (n int, err error)

	Sync() error
	Seek(offset int64, whence int) (ret int64, err error)

	Close() error
}
