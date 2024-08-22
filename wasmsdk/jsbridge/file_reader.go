//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"errors"
	"io"
	"syscall/js"
)

type FileReader struct {
	size      int64
	offset    int64
	readChunk js.Value
	buf       []byte
	bufOffset int
	endOfFile bool
}

const (
	bufferSize = 16 * 1024 * 1024 //16MB
)

func NewFileReader(readChunkFuncName string, fileSize, chunkReadSize int64) (*FileReader, error) {
	readChunk := js.Global().Get(readChunkFuncName)
	var buf []byte
	if bufferSize > fileSize {
		buf = make([]byte, fileSize)
	} else {
		bufSize := (chunkReadSize * (bufferSize / chunkReadSize))
		buf = make([]byte, bufSize)
	}
	result, err := Await(readChunk.Invoke(0, len(buf)))
	if len(err) > 0 && !err[0].IsNull() {
		return nil, errors.New("file_reader: " + err[0].String())
	}
	chunk := result[0]
	n := js.CopyBytesToGo(buf, chunk)
	if n < len(buf) {
		return nil, errors.New("file_reader: failed to read first chunk")
	}
	return &FileReader{
		size:      fileSize,
		offset:    int64(n),
		readChunk: readChunk,
		buf:       buf,
		endOfFile: n == int(fileSize),
	}, nil
}

func (r *FileReader) Read(p []byte) (int, error) {
	//js.Value doesn't work in parallel invoke
	size := len(p)

	if len(r.buf)-r.bufOffset < size && !r.endOfFile {
		r.bufOffset = 0 //reset buffer offset
		result, err := Await(r.readChunk.Invoke(r.offset, len(r.buf)))

		if len(err) > 0 && !err[0].IsNull() {
			return 0, errors.New("file_reader: " + err[0].String())
		}

		chunk := result[0]

		n := js.CopyBytesToGo(r.buf, chunk)
		r.offset += int64(n)
		if n < len(r.buf) {
			r.buf = r.buf[:n]
			r.endOfFile = true
		}
	}

	n := copy(p, r.buf[r.bufOffset:])
	r.bufOffset += n
	if r.endOfFile && r.bufOffset == len(r.buf) {
		return n, io.EOF
	}

	return n, nil
}

func (r *FileReader) Seek(offset int64, whence int) (int64, error) {

	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = r.offset + offset
	case io.SeekEnd:
		abs = r.size + offset
	default:
		return 0, errors.New("FileReader.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("FileReader.Seek: negative position")
	}
	if abs > int64(len(r.buf)) {
		return 0, errors.New("FileReader.Seek: position out of bounds")
	}
	r.bufOffset = int(abs)
	return abs, nil
}
