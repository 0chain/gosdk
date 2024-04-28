//go:build js && wasm
// +build js,wasm

package sys

import (
	"io/fs"
	"os"
	"path/filepath"
	"syscall/js"
	"time"
)

// MemFS implement file system on memory
type MemFS struct {
	files map[string]*MemFile
}

// NewMemFS create MemFS instance
func NewMemFS() FS {
	return &MemFS{
		files: make(map[string]*MemFile),
	}
}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *PathError.
func (mfs *MemFS) Open(name string) (File, error) {
	file := mfs.files[name]
	if file != nil {
		return file, nil
	}

	fileName := filepath.Base(name)

	file = &MemFile{Name: fileName, Mode: fs.ModePerm, ModTime: time.Now()}

	mfs.files[name] = file

	return file, nil
}

func (mfs *MemFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	file := mfs.files[name]
	if file != nil {
		return file, nil
	}

	fileName := filepath.Base(name)
	file = &MemFile{Name: fileName, Mode: perm, ModTime: time.Now()}

	mfs.files[name] = file

	return file, nil

}

// ReadFile reads the file named by filename and returns the contents.
func (mfs *MemFS) ReadFile(name string) ([]byte, error) {
	file, ok := mfs.files[name]
	if ok {
		return file.Buffer, nil
	}

	return nil, os.ErrNotExist
}

// WriteFile writes data to a file named by filename.
func (mfs *MemFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	fileName := filepath.Base(name)
	file := &MemFile{Name: fileName, Mode: perm, ModTime: time.Now()}

	mfs.files[name] = file

	return nil
}

// Remove removes the named file or (empty) directory.
// If there is an error, it will be of type *PathError.
func (mfs *MemFS) Remove(name string) error {
	delete(mfs.files, name)
	return nil
}

// MkdirAll creates a directory named path
func (mfs *MemFS) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

// Stat returns a FileInfo describing the named file.
// If there is an error, it will be of type *PathError.
func (mfs *MemFS) Stat(name string) (fs.FileInfo, error) {
	file, ok := mfs.files[name]
	if ok {
		return file.Stat()
	}

	return nil, os.ErrNotExist
}

func (mfs *MemFS) LoadProgress(progressID string) ([]byte, error) {
	key := filepath.Base(progressID)
	val := js.Global().Get("localStorage").Call("getItem", key)
	if val.Truthy() {
		return []byte(val.String()), nil
	}
	return nil, os.ErrNotExist
}

func (mfs *MemFS) SaveProgress(progressID string, data []byte, _ fs.FileMode) error {
	key := filepath.Base(progressID)
	js.Global().Get("localStorage").Call("setItem", key, string(data))
	return nil
}

func (mfs *MemFS) RemoveProgress(progressID string) error {
	key := filepath.Base(progressID)
	js.Global().Get("localStorage").Call("removeItem", key)
	return nil
}
