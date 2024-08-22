//go:build js && wasm
// +build js,wasm

package main

import (
	"path"
	"sync"

	"github.com/0chain/gosdk/core/sys"
	"gopkg.in/cheggaaa/pb.v1"
)

// StatusBar is to check status of any operation
type StatusBar struct {
	b       *pb.ProgressBar
	wg      *sync.WaitGroup
	success bool
	err     error

	totalBytes     int
	completedBytes int
	objURL         string
	localPath      string
	callback       func(totalBytes int, completedBytes int, fileName, objURL, err string)
	isRepair       bool
}

var jsCallbackMutex sync.Mutex

// Started for statusBar
func (s *StatusBar) Started(allocationID, filePath string, op int, totalBytes int) {
	if logEnabled {
		s.b = pb.StartNew(totalBytes)
		s.b.Set(0)
	}
	fileName := path.Base(filePath)
	s.totalBytes = totalBytes
	if s.callback != nil {
		jsCallbackMutex.Lock()
		defer jsCallbackMutex.Unlock()
		s.callback(s.totalBytes, s.completedBytes, fileName, "", "")
	}
}

// InProgress for statusBar
func (s *StatusBar) InProgress(allocationID, filePath string, op int, completedBytes int, todo_name_var []byte) {
	if logEnabled && s.b != nil {
		s.b.Set(completedBytes)
	}
	fileName := path.Base(filePath)
	s.completedBytes = completedBytes
	if s.callback != nil {
		jsCallbackMutex.Lock()
		defer jsCallbackMutex.Unlock()
		s.callback(s.totalBytes, s.completedBytes, fileName, "", "")
	}
}

// Completed for statusBar
func (s *StatusBar) Completed(allocationID, filePath string, filename string, mimetype string, size int, op int) {
	if logEnabled && s.b != nil {
		s.b.Finish()
	}
	s.success = true

	s.completedBytes = s.totalBytes
	if s.localPath != "" {
		fs, _ := sys.Files.Open(s.localPath)
		mf, _ := fs.(*sys.MemFile)
		s.objURL = CreateObjectURL(mf.Buffer, mimetype)
	}
	if s.callback != nil {
		jsCallbackMutex.Lock()
		defer jsCallbackMutex.Unlock()
		s.callback(s.totalBytes, s.completedBytes, filename, s.objURL, "")
	}
	if !s.isRepair {
		defer s.wg.Done()
	}
	sdkLogger.Info("Status completed callback. Type = " + mimetype + ". Name = " + filename + ". URL = " + s.objURL)
}

// Error for statusBar
func (s *StatusBar) Error(allocationID string, filePath string, op int, err error) {
	if s.b != nil {
		s.b.Finish()
	}
	s.success = false
	s.err = err
	defer func() {
		if r := recover(); r != nil {
			PrintError("Recovered in statusBar Error", r)
		}
	}()
	fileName := path.Base(filePath)
	PrintError("Error in file operation." + err.Error())
	if s.callback != nil {
		jsCallbackMutex.Lock()
		defer jsCallbackMutex.Unlock()
		s.callback(s.totalBytes, s.completedBytes, fileName, "", err.Error())
	}
	if !s.isRepair {
		s.wg.Done()
	}
}

// RepairCompleted when repair is completed
func (s *StatusBar) RepairCompleted(filesRepaired int) {
	s.wg.Done()
}
