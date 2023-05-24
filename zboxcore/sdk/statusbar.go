package sdk

import (
	"sync"

	"gopkg.in/cheggaaa/pb.v1"
	"github.com/0chain/gosdk/zboxcore/logger"
)
var logEnabled = false

// StatusBar is to check status of any operation
type StatusBar struct {
	b       *pb.ProgressBar
	wg      *sync.WaitGroup
	success bool
	err     error

	totalBytes     int
	completedBytes int
	callback       func(totalBytes int, completedBytes int, err string)
}

var jsCallbackMutex sync.Mutex

// Started for statusBar
func (s *StatusBar) Started(allocationID, filePath string, op int, totalBytes int) {
	if logEnabled {
		s.b = pb.StartNew(totalBytes)
		s.b.Set(0)
	}

	s.totalBytes = totalBytes
	if s.callback != nil {
		jsCallbackMutex.Lock()
		defer jsCallbackMutex.Unlock()
		s.callback(s.totalBytes, s.completedBytes, "")
	}
}

// InProgress for statusBar
func (s *StatusBar) InProgress(allocationID, filePath string, op int, completedBytes int, todo_name_var []byte) {
	if logEnabled && s.b != nil {
		s.b.Set(completedBytes)
	}

	s.completedBytes = completedBytes
	if s.callback != nil {
		jsCallbackMutex.Lock()
		defer jsCallbackMutex.Unlock()
		s.callback(s.totalBytes, s.completedBytes, "")
	}
}

// Completed for statusBar
func (s *StatusBar) Completed(allocationID, filePath string, filename string, mimetype string, size int, op int) {
	if logEnabled && s.b != nil {
		s.b.Finish()
	}
	s.success = true

	s.completedBytes = s.totalBytes
	if s.callback != nil {
		jsCallbackMutex.Lock()
		defer jsCallbackMutex.Unlock()
		s.callback(s.totalBytes, s.completedBytes, "")
	}

	defer s.wg.Done()
	logger.Logger.Info("Status completed callback. Type = " + mimetype + ". Name = " + filename)
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
			logger.Logger.Error("Recovered in statusBar Error", r)
		}
	}()
	logger.Logger.Error("Error in file operation." + err.Error())
	if s.callback != nil {
		jsCallbackMutex.Lock()
		defer jsCallbackMutex.Unlock()
		s.callback(s.totalBytes, s.completedBytes, err.Error())
	}
	s.wg.Done()
}

// RepairCompleted when repair is completed
func (s *StatusBar) RepairCompleted(filesRepaired int) {
	s.wg.Done()
}
