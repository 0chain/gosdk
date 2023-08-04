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

}
func (s *StatusBar) InProgress(allocationId, filePath string, op int, completedBytes int, data []byte) {

}

func (s *StatusBar) Completed(allocationId, filePath string, filename string, mimetype string, size int, op int) {
	s.success = true
	l.Logger.Info("Repair for file completed. File = ", filePath)
}

func (s *StatusBar) Error(allocationID string, filePath string, op int, err error) {
	s.success = false
	s.err = err
	l.Logger.Error("Error in status callback. Error = ", err.Error())
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
	wg      *sync.WaitGroup
	allocID string
	success bool
	err     error
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
	return &StatusBar{
		wg:      &sync.WaitGroup{},
		allocID: allocID,
	}
}

func mutUnlock(allocID string) {
	mapLock.Lock()
	mutMap[allocID].Unlock()
	mapLock.Unlock()
}
