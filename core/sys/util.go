package sys

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"time"
)

type MemFile struct {
	Name    string
	Buffer  []byte      // file content
	Mode    fs.FileMode // FileInfo.Mode
	ModTime time.Time   // FileInfo.ModTime
	Sys     interface{} // FileInfo.Sys
	reader  io.Reader
}

func (f *MemFile) Stat() (fs.FileInfo, error) {
	return &MemFileInfo{name: f.Name, f: f}, nil
}
func (f *MemFile) Read(p []byte) (int, error) {
	if f.reader == nil {
		f.reader = bytes.NewReader(f.Buffer)
	}
	return f.reader.Read(p)

}
func (f *MemFile) Write(p []byte) (n int, err error) {
	f.Buffer = append(f.Buffer, p...)
	return len(p), nil
}

func (f *MemFile) WriteAt(p []byte, offset int64) (n int, err error) {
	if offset < 0 || offset > int64(len(f.Buffer)) {
		return 0, io.ErrShortWrite
	}

	copy(f.Buffer[offset:], p)

	return len(p), nil
}

func (f *MemFile) InitBuffer(size int) {
	f.Buffer = make([]byte, size)
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
	case offset > int64(len(f.Buffer)):
		return 0, io.EOF
	default:
		f.reader = bytes.NewReader(f.Buffer[offset:])
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
	return int64(len(i.f.Buffer))
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
	ErrChan        chan error
	data           []byte
}

func (f *MemChanFile) Stat() (fs.FileInfo, error) {
	return &MemFileChanInfo{name: f.Name, f: f}, nil
}
func (f *MemChanFile) Read(p []byte) (int, error) {
	select {
	case err := <-f.ErrChan:
		return 0, err
	case recieveData, ok := <-f.Buffer:
		if !ok {
			return 0, io.EOF
		}
		if len(recieveData) > len(p) {
			return 0, io.ErrShortBuffer
		}
		n := copy(p, recieveData)
		return n, nil
	}
}

func (f *MemChanFile) Write(p []byte) (n int, err error) {
	if f.ChunkWriteSize == 0 {
		data := make([]byte, len(p))
		copy(data, p)
		f.Buffer <- data
	} else {
		if cap(f.data) == 0 {
			f.data = make([]byte, 0, len(p))
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
	f.data = nil
	return nil
}
func (f *MemChanFile) Seek(offset int64, whence int) (ret int64, err error) {
	return 0, nil
}

func (f *MemChanFile) Close() error {
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
