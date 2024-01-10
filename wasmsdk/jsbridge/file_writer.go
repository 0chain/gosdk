//go:build js && wasm
// +build js,wasm

package jsbridge

import (
	"errors"
	"io/fs"
	"sync"
	"syscall/js"
)

var jsFileWriterMutex sync.Mutex

type FileWriter struct {
	writableStream js.Value
	uint8Array     js.Value
	fileHandle     js.Value
	bufLen         int
	written        int
	totalWritten   int
}

func (w *FileWriter) Write(p []byte) (int, error) {
	//js.Value doesn't work in parallel invoke
	jsFileWriterMutex.Lock()
	defer jsFileWriterMutex.Unlock()

	if w.bufLen != len(p) {
		w.bufLen = len(p)
		w.uint8Array = js.Global().Get("Uint8Array").New(w.bufLen)
	}
	js.CopyBytesToJS(w.uint8Array, p)
	w.writableStream.Call("write", w.uint8Array)
	w.written++
	w.totalWritten += len(p)
	if w.written >= 3 {
		w.written = 0
		w.writableStream.Call("flush")
	}
	return len(p), nil
}

func (w *FileWriter) Close() error {
	w.writableStream.Call("close")
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
	writableStream, err := Await(fileHandle[0].Call("createSyncAccessHandle"))
	if len(err) > 0 && !err[0].IsNull() {
		return nil, errors.New("file_writer: " + err[0].String())
	}
	return &FileWriter{
		writableStream: writableStream[0],
		fileHandle:     fileHandle[0],
	}, nil
}

// func (w *FileWriter) SetNewWriter(offset int) error {
// 	writableStream, err := Await(w.fileHandle.Call("createWritable"))
// 	if len(err) > 0 && !err[0].IsNull() {
// 		return errors.New("file_writer: " + err[0].String())
// 	}
// 	w.writableStream = writableStream[0]
// 	_, err = Await(w.writableStream.Call("seek", offset))
// 	if len(err) > 0 && !err[0].IsNull() {
// 		return errors.New("file_writer: " + err[0].String())
// 	}
// 	return nil
// }
