package sdk

import (
	"sync"

	l "github.com/0chain/gosdk/zboxcore/logger"
	"gopkg.in/cheggaaa/pb.v1"
)

var (
	mutMap  = make(map[string]*sync.Mutex)
	mapLock sync.Mutex
)

func (s *StatusBar) Started(allocationId, filePath string, op int, totalBytes int) {
	s.b = pb.StartNew(totalBytes)
	s.b.Set(0)
}
func (s *StatusBar) InProgress(allocationId, filePath string, op int, completedBytes int, data []byte) {
	s.b.Set(completedBytes)
}

func (s *StatusBar) Completed(allocationId, filePath string, filename string, mimetype string, size int, op int) {
	if s.b != nil {
		s.b.Finish()
	}
	s.success = true
	s.wg.Done()
	mutUnlock(s.allocID)
}

func (s *StatusBar) Error(allocationID string, filePath string, op int, err error) {
	if s.b != nil {
		s.b.Finish()
	}
	s.success = false
	defer s.wg.Done()
	defer mutUnlock(s.allocID)

	var errDetail interface{} = "Unknown Error"
	if err != nil {
		errDetail = err.Error()
	}

	l.Logger.Error("Error in status callback. Error = ", errDetail)
}

func (s *StatusBar) RepairCompleted(filesRepaired int) {
	defer s.wg.Done()
	mutUnlock(s.allocID)
	l.Logger.Info("Repair completed. Files repaired = ", filesRepaired)
}

type StatusBar struct {
	b       *pb.ProgressBar
	wg      *sync.WaitGroup
	allocID string
	success bool
}

func NewRepairBar(allocID string) *StatusBar {
	if _, ok := mutMap[allocID]; !ok {
		mapLock.Lock()
		mutMap[allocID] = &sync.Mutex{}
		mapLock.Unlock()
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
