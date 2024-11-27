package zbox

import (
	"path"
	"sync"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

// StatusBar is to check status of any operation
type StatusBar struct {
	wg      *sync.WaitGroup
	success bool
	err     error

	totalBytes     int
	completedBytes int
	objURL         string
	localPath      string
	callback       func(totalBytes int, completedBytes int, fileName, objURL, err string)
	isRepair       bool
	totalBytesMap  map[string]int
}

var statusCallbackMutex sync.Mutex

// Started for statusBar
func (s *StatusBar) Started(allocationID, filePath string, op int, totalBytes int) {
	fileName := path.Base(filePath)
	s.totalBytes = totalBytes
	if s.callback != nil {
		if !s.isRepair || op == sdk.OpUpload || op == sdk.OpUpdate {
			if s.isRepair {
				fileName = filePath
			}
			statusCallbackMutex.Lock()
			defer statusCallbackMutex.Unlock()
			s.totalBytesMap[filePath] = totalBytes
			s.callback(totalBytes, s.completedBytes, fileName, "", "")
		}
	}
}

// InProgress for statusBar
func (s *StatusBar) InProgress(allocationID, filePath string, op int, completedBytes int, todo_name_var []byte) {
	fileName := path.Base(filePath)
	if s.callback != nil {
		if !s.isRepair || op == sdk.OpUpload || op == sdk.OpUpdate {
			if s.isRepair {
				fileName = filePath
			}
			statusCallbackMutex.Lock()
			defer statusCallbackMutex.Unlock()
			s.callback(s.totalBytesMap[filePath], completedBytes, fileName, "", "")
		}
	}
}

// Completed for statusBar
func (s *StatusBar) Completed(allocationID, filePath string, filename string, mimetype string, size int, op int) {
	s.success = true

	if s.localPath != "" {
		fs, _ := sys.Files.Open(s.localPath)
		mf, _ := fs.(*sys.MemFile)
		if mf != nil {
			s.objURL = string(mf.Buffer)
		}
	}
	if s.callback != nil {
		if !s.isRepair || op == sdk.OpUpload || op == sdk.OpUpdate {
			if s.isRepair {
				filename = filePath
			}
			statusCallbackMutex.Lock()
			defer statusCallbackMutex.Unlock()
			totalBytes := s.totalBytesMap[filePath]
			delete(s.totalBytesMap, filePath)
			s.callback(totalBytes, totalBytes, filename, s.objURL, "")
		}
	}
	if !s.isRepair {
		defer s.wg.Done()
	}
}

// Error for statusBar
func (s *StatusBar) Error(allocationID string, filePath string, op int, err error) {
	s.success = false
	s.err = err
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Error("Recovered in statusBar Error", r)
		}
	}()
	fileName := path.Base(filePath)
	logger.Logger.Error("Error in file operation: ", err)
	if s.callback != nil {
		if !s.isRepair || op == sdk.OpUpload || op == sdk.OpUpdate {
			if s.isRepair {
				fileName = filePath
			}
			statusCallbackMutex.Lock()
			defer statusCallbackMutex.Unlock()
			s.callback(s.totalBytesMap[filePath], s.completedBytes, fileName, "", err.Error())
		}
	}
	if !s.isRepair {
		s.wg.Done()
	}
}

// RepairCompleted when repair is completed
func (s *StatusBar) RepairCompleted(filesRepaired int) {
	s.wg.Done()
}

// NewStatusBar creates a new StatusBar instance
func NewStatusBar(wg *sync.WaitGroup) *StatusBar {
	return &StatusBar{
		wg:            wg,
		totalBytesMap: make(map[string]int),
	}
}
