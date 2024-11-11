package sys

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/0chain/gosdk/core/common"
	"github.com/valyala/bytebufferpool"
)

// MemFile represents a file totally loaded in memory
// Aware of the file size, so it can seek and truncate.
type MemFile struct {
	Name    string
	Buffer  []byte      // file content
	Mode    fs.FileMode // FileInfo.Mode
	ModTime time.Time   // FileInfo.ModTime
	Sys     interface{} // FileInfo.Sys
	reader  io.Reader
}

// Stat returns the file information
func (f *MemFile) Stat() (fs.FileInfo, error) {
	return &MemFileInfo{name: f.Name, f: f}, nil
}

// Read reads data from the file
func (f *MemFile) Read(p []byte) (int, error) {
	if f.reader == nil {
		f.reader = bytes.NewReader(f.Buffer)
	}
	return f.reader.Read(p)

}

// Write writes data to the file
func (f *MemFile) Write(p []byte) (n int, err error) {
	f.Buffer = append(f.Buffer, p...)
	return len(p), nil
}

// WriteAt writes data to the file at a specific offset
func (f *MemFile) WriteAt(p []byte, offset int64) (n int, err error) {
	if offset < 0 || offset > int64(len(f.Buffer)) || len(p) > len(f.Buffer)-int(offset) {
		return 0, io.ErrShortWrite
	}

	copy(f.Buffer[offset:], p)

	return len(p), nil
}

// InitBuffer initializes the buffer with a specific size
func (f *MemFile) InitBuffer(size int) {
	buff := common.MemPool.Get()
	if cap(buff.B) < size {
		buff.B = make([]byte, size)
	}
	f.Buffer = buff.B[:size]
}

// Sync not implemented
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

// MemFileInfo represents file information
type MemFileInfo struct {
	name string
	f    *MemFile
}

// Name returns the base name of the file
func (i *MemFileInfo) Name() string {
	return i.name
}

// Size returns the size of the file
func (i *MemFileInfo) Size() int64 {
	return int64(len(i.f.Buffer))
}

// Mode returns the file mode bits
func (i *MemFileInfo) Mode() fs.FileMode {
	return i.f.Mode
}

// Type returns the file mode type
func (i *MemFileInfo) Type() fs.FileMode {
	return i.f.Mode.Type()
}

// ModTime returns the modification time of the file
func (i *MemFileInfo) ModTime() time.Time {
	return i.f.ModTime
}

// IsDir returns true if the file is a directory
func (i *MemFileInfo) IsDir() bool {
	return i.f.Mode&fs.ModeDir != 0
}

// Sys returns the underlying data source (can return nil)
func (i *MemFileInfo) Sys() interface{} {
	return i.f.Sys
}

// Info returns the file information
func (i *MemFileInfo) Info() (fs.FileInfo, error) {
	return i, nil
}

// MemChanFile used to read or write file content sequentially through a buffer channel.
// Not aware of the file size, so it can't seek or truncate.
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

// Stat returns the file information
func (f *MemChanFile) Stat() (fs.FileInfo, error) {
	return &MemFileChanInfo{name: f.Name, f: f}, nil
}

// Read reads data from the file through the buffer channel
// It returns io.EOF when the buffer channel is closed.
//   - p: file in bytes loaded from the buffer channel
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

// Write writes data to the file through the buffer channel
// It writes the data to the buffer channel in chunks of ChunkWriteSize.
// If ChunkWriteSize is 0, it writes the data as a whole.
//   - p: file in bytes to write to the buffer channel
func (f *MemChanFile) Write(p []byte) (n int, err error) {
	if f.ChunkWriteSize == 0 {
		data := make([]byte, len(p))
		copy(data, p)
		f.Buffer <- data
	} else {
		if cap(f.data) == 0 {
			bbuf := common.MemPool.Get()
			if cap(bbuf.B) < len(p) {
				bbuf.B = make([]byte, 0, len(p))
			}
			f.data = bbuf.B
		}
		f.data = append(f.data, p...)
	}
	return len(p), nil
}

// Sync write the data chunk to the buffer channel
// It writes the data to the buffer channel in chunks of ChunkWriteSize.
// If ChunkWriteSize is 0, it writes the data as a whole.
func (f *MemChanFile) Sync() error {
	current := 0
	for ; current < len(f.data); current += f.ChunkWriteSize {
		end := current + f.ChunkWriteSize
		if end > len(f.data) {
			end = len(f.data)
		}
		f.Buffer <- f.data[current:end]
	}
	f.data = f.data[:0]
	return nil
}

// Seek not implemented
func (f *MemChanFile) Seek(offset int64, whence int) (ret int64, err error) {
	return 0, nil
}

// Close closes the buffer channel
func (f *MemChanFile) Close() error {
	close(f.Buffer)
	if cap(f.data) > 0 {
		bbuf := &bytebufferpool.ByteBuffer{
			B: f.data,
		}
		common.MemPool.Put(bbuf)
	}
	return nil
}

// MemFileChanInfo represents file information
type MemFileChanInfo struct {
	name string
	f    *MemChanFile
}

// Name returns the base name of the file
func (i *MemFileChanInfo) Name() string {
	return i.name
}

// Size not implemented
func (i *MemFileChanInfo) Size() int64 {
	return 0
}

// Mode returns the file mode bits
func (i *MemFileChanInfo) Mode() fs.FileMode {
	return i.f.Mode
}

// Type returns the file mode type
func (i *MemFileChanInfo) Type() fs.FileMode {
	return i.f.Mode.Type()
}

// ModTime returns the modification time of the file
func (i *MemFileChanInfo) ModTime() time.Time {
	return i.f.ModTime
}

// IsDir returns true if the file is a directory
func (i *MemFileChanInfo) IsDir() bool {
	return i.f.Mode&fs.ModeDir != 0
}

// Sys returns the underlying data source (can return nil)
func (i *MemFileChanInfo) Sys() interface{} {
	return i.f.Sys
}

// Info returns the file information
func (i *MemFileChanInfo) Info() (fs.FileInfo, error) {
	return i, nil
}
