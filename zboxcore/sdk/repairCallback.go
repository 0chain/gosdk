package sdk

import (
	"sync"

	l "github.com/0chain/gosdk/zboxcore/logger"
)

var (
	mutMap  = make(map[string]*sync.Mutex)
	mapLock sync.Mutex
)

func (s *StatusBar) Started(allocationId, filePath string, op int, totalBytes int) {
	if s.sb != nil {
		s.sb.Started(allocationId, filePath, op, totalBytes)
	}

}
func (s *StatusBar) InProgress(allocationId, filePath string, op int, completedBytes int, data []byte) {
	if s.sb != nil {
		s.sb.InProgress(allocationId, filePath, op, completedBytes, data)
	}
}

func (s *StatusBar) Completed(allocationId, filePath string, filename string, mimetype string, size int, op int) {
	s.success = true
	if s.isRepair {
		l.Logger.Info("Repair for file completed. File = ", filePath)
	} else {
		l.Logger.Info("Operation completed. File = ", filePath)
		s.wg.Done()
	}
}

func (s *StatusBar) Error(allocationID string, filePath string, op int, err error) {
	s.success = false
	s.err = err
	l.Logger.Error("Error in status callback. Error = ", err.Error())
	if !s.isRepair {
		s.wg.Done()
	}
}

func (s *StatusBar) RepairCompleted(filesRepaired int) {
	if s.err == nil {
		s.success = true
	}
	defer s.wg.Done()
	mutUnlock(s.allocID)
	l.Logger.Info("Repair completed. Files repaired = ", filesRepaired)

}

type StatusBar struct {
	wg       *sync.WaitGroup
	allocID  string
	success  bool
	err      error
	isRepair bool
	sb       StatusCallback
}

func NewRepairBar(allocID string) *StatusBar {
	mapLock.Lock()
	defer mapLock.Unlock()
	if _, ok := mutMap[allocID]; !ok {
		mutMap[allocID] = &sync.Mutex{}
	}
	if !mutMap[allocID].TryLock() {
		return nil
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	return &StatusBar{
		wg:       wg,
		allocID:  allocID,
		isRepair: true,
	}
}

func (s *StatusBar) Wait() {
	s.wg.Wait()
}

func (s *StatusBar) CheckError() error {
	if !s.success {
		return s.err
	}
	return nil
}

func mutUnlock(allocID string) {
	mapLock.Lock()
	mutMap[allocID].Unlock()
	mapLock.Unlock()
}
