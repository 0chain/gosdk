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
}

func NewFileReader(readChunkFuncName string, fileSize int64) *FileReader {
	readChunk := js.Global().Get(readChunkFuncName)

	return &FileReader{
		size:      fileSize,
		offset:    0,
		readChunk: readChunk,
	}
}

func (r *FileReader) Read(p []byte) (int, error) {

	size := len(p)

	result, err := Await(r.readChunk.Invoke(r.offset, size))

	if len(err) > 0 && !err[0].IsNull() {
		return 0, errors.New("read_chunk: " + err[0].String())
	}

	chunk := result[0]

	n := js.CopyBytesToGo(p, chunk)
	r.offset += int64(n)

	if n < size {
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
	r.offset = abs
	return abs, nil
}
