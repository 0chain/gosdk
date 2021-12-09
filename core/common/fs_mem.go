package common

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

	file = &MemFile{Name: fileName, Buffer: new(bytes.Buffer), Mode: fs.ModePerm, ModTime: time.Now()}

	mfs.files[name] = file

	return file, nil
}

func (mfs *MemFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	file := mfs.files[name]
	if file != nil {
		return file, nil
	}

	fileName := filepath.Base(name)
	file = &MemFile{Name: fileName, Buffer: new(bytes.Buffer), Mode: perm, ModTime: time.Now()}

	mfs.files[name] = file

	return file, nil

}

// Remove removes the named file or (empty) directory.
// If there is an error, it will be of type *PathError.
func (mfs *MemFS) Remove(name string) error {
	delete(mfs.files, name)
	return nil
}

type MemFile struct {
	Name    string
	Buffer  *bytes.Buffer // file content
	Mode    fs.FileMode   // FileInfo.Mode
	ModTime time.Time     // FileInfo.ModTime
	Sys     interface{}   // FileInfo.Sys
	reader  io.Reader
}

func (f *MemFile) Stat() (fs.FileInfo, error) {
	return &MemFileInfo{name: f.Name, f: f}, nil
}
func (f *MemFile) Read(p []byte) (int, error) {
	if f.reader == nil {
		f.reader = bytes.NewReader(f.Buffer.Bytes())
	}
	return f.reader.Read(p)

}
func (f *MemFile) Write(p []byte) (n int, err error) {
	return f.Buffer.Write(p)

}

func (f *MemFile) Sync() error {
	return nil
}
func (f *MemFile) Seek(offset int64, whence int) (ret int64, err error) {

	// always reset it from beginning, it only work for wasm download
	f.reader = bytes.NewReader(f.Buffer.Bytes())

	return 0, nil
}

func (f *MemFile) Close() error {
	f.reader = nil
	return nil
}

type MemFileInfo struct {
	name string
	f    *MemFile
}

func (i *MemFileInfo) Name() string {
	return i.name
}
func (i *MemFileInfo) Size() int64 {
	return int64(i.f.Buffer.Len())
}

func (i *MemFileInfo) Mode() fs.FileMode {
	return i.f.Mode
}

func (i *MemFileInfo) Type() fs.FileMode {
	return i.f.Mode.Type()
}

func (i *MemFileInfo) ModTime() time.Time {
	return i.f.ModTime
}

func (i *MemFileInfo) IsDir() bool {
	return i.f.Mode&fs.ModeDir != 0
}

func (i *MemFileInfo) Sys() interface{} {
	return i.f.Sys
}

func (i *MemFileInfo) Info() (fs.FileInfo, error) {
	return i, nil
}
