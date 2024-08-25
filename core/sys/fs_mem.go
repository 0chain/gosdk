//go:build js && wasm
// +build js,wasm

package sys

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall/js"
	"time"

	"github.com/0chain/gosdk/wasmsdk/jsbridge"
)

// MemFS implement file system on memory
type MemFS struct {
	files map[string]*MemFile
	dirs  map[string]js.Value
}

// NewMemFS create MemFS instance
func NewMemFS() FS {
	return &MemFS{
		files: make(map[string]*MemFile),
		dirs:  make(map[string]js.Value),
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

func (mfs *MemFS) CreateDirectory(dirID string) error {
	if !js.Global().Get("showDirectoryPicker").Truthy() || !js.Global().Get("WritableStream").Truthy() {
		return errors.New("dir_picker: not supported")
	}
	showDirectoryPicker := js.Global().Get("showDirectoryPicker")
	dirHandle, err := jsbridge.Await(showDirectoryPicker.Invoke())
	if len(err) > 0 && !err[0].IsNull() {
		return errors.New("dir_picker: " + err[0].String())
	}
	mfs.dirs[dirID] = dirHandle[0]
	return nil
}

func (mfs *MemFS) GetFileHandler(dirID, path string) (File, error) {
	dirHandler, ok := mfs.dirs[dirID]
	if !ok {
		return nil, errors.New("dir_picker: directory not found")
	}
	currHandler, err := mfs.mkdir(dirHandler, filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	return jsbridge.NewFileWriterFromHandle(currHandler, filepath.Base(path))
}

func (mfs *MemFS) RemoveAllDirectories() {
	for k := range mfs.dirs {
		delete(mfs.dirs, k)
	}
}

func (mfs *MemFS) mkdir(dirHandler js.Value, dirPath string) (js.Value, error) {
	if dirPath == "/" {
		return dirHandler, nil
	}
	currHandler, ok := mfs.dirs[dirPath]
	if !ok {
		currHandler = dirHandler
		paths := strings.Split(dirPath, "/")
		paths = paths[1:]
		currPath := "/"
		for _, path := range paths {
			currPath = filepath.Join(currPath, path)
			handler, ok := mfs.dirs[currPath]
			if ok {
				currHandler = handler
				continue
			}
			options := js.Global().Get("Object").New()
			options.Set("create", true)
			currHandlers, err := jsbridge.Await(currHandler.Call("getDirectoryHandle", path, options))
			if len(err) > 0 && !err[0].IsNull() {
				return js.Value{}, errors.New("dir_picker: " + err[0].String())
			}
			currHandler = currHandlers[0]
			mfs.dirs[currPath] = currHandler
		}
		if !currHandler.Truthy() {
			return js.Value{}, errors.New("dir_picker: failed to create directory")
		}
	}
	return currHandler, nil
}
