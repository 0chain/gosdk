package main

import (
	"sync"
)

type StatusCallback2 struct {
	wg       *sync.WaitGroup
	isRepair bool
	success  bool
	err      error
}

func (cb *StatusCallback2) Started(allocationId, filePath string, op, totalBytes int) {

}

func (cb *StatusCallback2) InProgress(allocationId, filePath string, op, completedBytes int, data []byte) {
}

func (cb *StatusCallback2) RepairCompleted(filesRepaired int) {
	if cb.err == nil {
		cb.success = true
	}
	cb.wg.Done()
}

func (cb *StatusCallback2) Completed(allocationId, filePath, filename, mimetype string, size, op int) {
	if !cb.isRepair {
		cb.success = true
		cb.wg.Done()
	}
}

func (cb *StatusCallback2) Error(allocationID, filePath string, op int, err error) {
	cb.success = false
	cb.err = err
	if !cb.isRepair {
		cb.wg.Done()
	}
}

func (cb *StatusCallback2) StatusError() error {
	return cb.err
}

func NewStatusCallback2(wg *sync.WaitGroup) *StatusCallback2 {
	return &StatusCallback2{
		wg: wg,
	}
}
