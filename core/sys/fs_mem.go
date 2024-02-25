package sys

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

// ReadFile reads the file named by filename and returns the contents.
func (mfs *MemFS) ReadFile(name string) ([]byte, error) {
	file, ok := mfs.files[name]
	if ok {
		return file.Buffer.Bytes(), nil
	}

	return nil, os.ErrNotExist
}

// WriteFile writes data to a file named by filename.
func (mfs *MemFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	fileName := filepath.Base(name)
	file := &MemFile{Name: fileName, Buffer: new(bytes.Buffer), Mode: perm, ModTime: time.Now()}

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

// MemFS implement file system on memory
type MemChanFS struct {
	files map[string]*MemChanFile
}

// NewMemFS create MemFS instance
func NewMemChanFS() FS {
	return &MemChanFS{
		files: make(map[string]*MemChanFile),
	}
}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *PathError.
func (mfs *MemChanFS) Open(name string) (File, error) {
	file := mfs.files[name]
	if file != nil {
		return file, nil
	}

	fileName := filepath.Base(name)

	file = &MemChanFile{Name: fileName, Buffer: make(chan []byte), Mode: fs.ModePerm, ModTime: time.Now()}

	mfs.files[name] = file

	return file, nil
}

func (mfs *MemChanFS) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	file := mfs.files[name]
	if file != nil {
		return file, nil
	}

	fileName := filepath.Base(name)
	file = &MemChanFile{Name: fileName, Buffer: make(chan []byte), Mode: perm, ModTime: time.Now()}

	mfs.files[name] = file

	return file, nil

}

// ReadFile reads the file named by filename and returns the contents.
func (mfs *MemChanFS) ReadFile(name string) ([]byte, error) {
	file, ok := mfs.files[name]
	if ok {
		return <-file.Buffer, nil
	}

	return nil, os.ErrNotExist
}

// WriteFile writes data to a file named by filename.
func (mfs *MemChanFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	fileName := filepath.Base(name)
	file := &MemChanFile{Name: fileName, Buffer: make(chan []byte), Mode: perm, ModTime: time.Now()}

	mfs.files[name] = file

	return nil
}

// Remove removes the named file or (empty) directory.
// If there is an error, it will be of type *PathError.
func (mfs *MemChanFS) Remove(name string) error {
	delete(mfs.files, name)
	return nil
}

// MkdirAll creates a directory named path
func (mfs *MemChanFS) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

// Stat returns a FileInfo describing the named file.
// If there is an error, it will be of type *PathError.
func (mfs *MemChanFS) Stat(name string) (fs.FileInfo, error) {
	file, ok := mfs.files[name]
	if ok {
		return file.Stat()
	}

	return nil, os.ErrNotExist
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
	if whence != io.SeekStart {
		return 0, os.ErrInvalid
	}
	switch {
	case offset < 0:
		return 0, os.ErrInvalid
	case offset > int64(f.Buffer.Len()):
		return 0, io.EOF
	default:
		f.reader = bytes.NewReader(f.Buffer.Bytes()[offset:])
		return offset, nil
	}
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

type MemChanFile struct {
	Name           string
	Buffer         chan []byte // file content
	Mode           fs.FileMode // FileInfo.Mode
	ModTime        time.Time   // FileInfo.ModTime
	ChunkWriteSize int         //  0 value means no limit
	Sys            interface{} // FileInfo.Sys
	reader         io.Reader
	data           []byte
}

func (f *MemChanFile) Stat() (fs.FileInfo, error) {
	return &MemFileChanInfo{name: f.Name, f: f}, nil
}
func (f *MemChanFile) Read(p []byte) (int, error) {
	recieveData, ok := <-f.Buffer
	if !ok {
		return 0, io.EOF
	}
	if len(recieveData) > len(p) {
		return 0, io.ErrShortBuffer
	}
	n := copy(p, recieveData)
	return n, nil
}

func (f *MemChanFile) Write(p []byte) (n int, err error) {
	if f.ChunkWriteSize == 0 {
		f.Buffer <- p
	} else {
		if cap(f.data) == 0 {
			f.data = make([]byte, 0, f.ChunkWriteSize)
		}
		f.data = append(f.data, p...)
	}
	return len(p), nil
}

func (f *MemChanFile) Sync() error {
	current := 0
	for ; current < len(f.data); current += f.ChunkWriteSize {
		end := current + f.ChunkWriteSize
		if end > len(f.data) {
			end = len(f.data)
		}
		f.Buffer <- f.data[current:end]
	}
	f.data = make([]byte, 0, f.ChunkWriteSize)
	return nil
}
func (f *MemChanFile) Seek(offset int64, whence int) (ret int64, err error) {
	return 0, nil
}

func (f *MemChanFile) Close() error {
	f.reader = nil
	close(f.Buffer)
	return nil
}

type MemFileChanInfo struct {
	name string
	f    *MemChanFile
}

func (i *MemFileChanInfo) Name() string {
	return i.name
}
func (i *MemFileChanInfo) Size() int64 {
	return 0
}

func (i *MemFileChanInfo) Mode() fs.FileMode {
	return i.f.Mode
}

func (i *MemFileChanInfo) Type() fs.FileMode {
	return i.f.Mode.Type()
}

func (i *MemFileChanInfo) ModTime() time.Time {
	return i.f.ModTime
}

func (i *MemFileChanInfo) IsDir() bool {
	return i.f.Mode&fs.ModeDir != 0
}

func (i *MemFileChanInfo) Sys() interface{} {
	return i.f.Sys
}

func (i *MemFileChanInfo) Info() (fs.FileInfo, error) {
	return i, nil
}
