//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"errors"
	"io"
	"io/fs"
	"syscall/js"

	"github.com/0chain/gosdk/core/common"
	"github.com/valyala/bytebufferpool"
)

type FileWriter struct {
	writableStream js.Value
	uint8Array     js.Value
	fileHandle     js.Value
	bufLen         int
	buf            []byte
	bufWriteOffset int
	writeError     bool
}

const writeBlocks = 10

// len(p) will always be <= 64KB
func (w *FileWriter) Write(p []byte) (int, error) {
	//init buffer if not initialized
	if len(w.buf) == 0 {
		w.buf = make([]byte, len(p)*writeBlocks)
	}

	//copy bytes to buf
	if w.bufWriteOffset+len(p) > len(w.buf) {
		w.writeError = true
		return 0, io.ErrShortWrite
	}
	n := copy(w.buf[w.bufWriteOffset:], p)
	w.bufWriteOffset += n
	if w.bufWriteOffset == len(w.buf) {
		//write to file
		if w.bufLen != len(w.buf) {
			w.bufLen = len(w.buf)
			w.uint8Array = js.Global().Get("Uint8Array").New(w.bufLen)
		}
		js.CopyBytesToJS(w.uint8Array, w.buf)
		_, err := Await(w.writableStream.Call("write", w.uint8Array))
		if len(err) > 0 && !err[0].IsNull() {
			w.writeError = true
			return 0, errors.New("file_writer: " + err[0].String())
		}
		//reset buffer
		w.bufWriteOffset = 0
	}
	return len(p), nil
}

// func (w *FileWriter) WriteAt(p []byte, offset int64) (int, error) {
// 	uint8Array := js.Global().Get("Uint8Array").New(len(p))
// 	js.CopyBytesToJS(uint8Array, p)
// 	options := js.Global().Get("Object").New()
// 	options.Set("type", "write")
// 	options.Set("position", offset)
// 	options.Set("data", uint8Array)
// 	options.Set("size", len(p))
// 	_, err := Await(w.writableStream.Call("write", options))
// 	if len(err) > 0 && !err[0].IsNull() {
// 		return 0, errors.New("file_writer: " + err[0].String())
// 	}
// 	return len(p), nil
// }

func (w *FileWriter) Close() error {

	if w.bufWriteOffset > 0 && !w.writeError {
		w.buf = w.buf[:w.bufWriteOffset]
		uint8Array := js.Global().Get("Uint8Array").New(len(w.buf))
		js.CopyBytesToJS(uint8Array, w.buf)
		_, err := Await(w.writableStream.Call("write", uint8Array))
		if len(err) > 0 && !err[0].IsNull() {
			return errors.New("file_writer: " + err[0].String())
		}
	}

	_, err := Await(w.writableStream.Call("close"))
	if len(err) > 0 && !err[0].IsNull() {
		return errors.New("file_writer: " + err[0].String())
	}
	return nil
}

func (w *FileWriter) Read(p []byte) (int, error) {
	return 0, errors.New("file_writer: not supported")
}

func (w *FileWriter) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (w *FileWriter) Sync() error {
	return nil
}

func (w *FileWriter) Stat() (fs.FileInfo, error) {
	return nil, nil
}

func NewFileWriter(filename string) (*FileWriter, error) {

	if !js.Global().Get("window").Get("showSaveFilePicker").Truthy() || !js.Global().Get("window").Get("WritableStream").Truthy() {
		return nil, errors.New("file_writer: not supported")
	}

	showSaveFilePicker := js.Global().Get("window").Get("showSaveFilePicker")
	//create options with suggested name
	options := js.Global().Get("Object").New()
	options.Set("suggestedName", filename)

	//request a file handle
	fileHandle, err := Await(showSaveFilePicker.Invoke(options))
	if len(err) > 0 && !err[0].IsNull() {
		return nil, errors.New("file_writer: " + err[0].String())
	}
	//create a writable stream
	writableStream, err := Await(fileHandle[0].Call("createWritable"))
	if len(err) > 0 && !err[0].IsNull() {
		return nil, errors.New("file_writer: " + err[0].String())
	}
	return &FileWriter{
		writableStream: writableStream[0],
		fileHandle:     fileHandle[0],
	}, nil
}

func NewFileWriterFromHandle(dirHandler js.Value, name string) (*FileWriter, error) {
	options := js.Global().Get("Object").New()
	options.Set("create", true)
	fileHandler, err := Await(dirHandler.Call("getFileHandle", name, options))
	if len(err) > 0 && !err[0].IsNull() {
		return nil, errors.New("dir_picker: " + err[0].String())
	}

	writableStream, err := Await(fileHandler[0].Call("createWritable"))
	if len(err) > 0 && !err[0].IsNull() {
		return nil, errors.New("file_writer: " + err[0].String())
	}
	return &FileWriter{
		writableStream: writableStream[0],
		fileHandle:     fileHandler[0],
	}, nil
}

type FileCallbackWriter struct {
	writeChunk js.Value
	buf        []byte
	offset     int64
}

const bufCallbackCap = 4 * 1024 * 1024 //4MB

func NewFileCallbackWriter(writeChunkFuncName string) *FileCallbackWriter {
	writeChunk := js.Global().Get(writeChunkFuncName)
	return &FileCallbackWriter{
		writeChunk: writeChunk,
	}
}

func (wc *FileCallbackWriter) Write(p []byte) (int, error) {
	if len(wc.buf) == 0 {
		buff := common.MemPool.Get()
		if cap(buff.B) < bufCallbackCap {
			buff.B = make([]byte, 0, bufCallbackCap)
		}
		wc.buf = buff.B
	}
	if len(wc.buf)+len(p) > cap(wc.buf) {
		uint8Array := js.Global().Get("Uint8Array").New(len(wc.buf))
		js.CopyBytesToJS(uint8Array, wc.buf)
		_, err := Await(wc.writeChunk.Invoke(uint8Array, wc.offset))
		if len(err) > 0 && !err[0].IsNull() {
			return 0, errors.New("file_writer: " + err[0].String())
		}
		wc.offset += int64(len(wc.buf))
		wc.buf = wc.buf[:0]
	}
	wc.buf = append(wc.buf, p...)
	return len(p), nil
}

func (wc *FileCallbackWriter) Close() error {
	if len(wc.buf) > 0 {
		uint8Array := js.Global().Get("Uint8Array").New(len(wc.buf))
		js.CopyBytesToJS(uint8Array, wc.buf)
		_, err := Await(wc.writeChunk.Invoke(uint8Array, wc.offset))
		if len(err) > 0 && !err[0].IsNull() {
			return errors.New("file_writer: " + err[0].String())
		}
		wc.offset += int64(len(wc.buf))
		wc.buf = wc.buf[:0]
	}
	buff := &bytebufferpool.ByteBuffer{
		B: wc.buf,
	}
	common.MemPool.Put(buff)
	return nil
}

func (wc *FileCallbackWriter) Read(p []byte) (int, error) {
	return 0, errors.New("file_writer: not supported")
}

func (wc *FileCallbackWriter) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (wc *FileCallbackWriter) Sync() error {
	return nil
}

func (wc *FileCallbackWriter) Stat() (fs.FileInfo, error) {
	return nil, nil
}
