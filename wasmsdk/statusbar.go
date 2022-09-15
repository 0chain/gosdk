//go:build js && wasm
// +build js,wasm

package main

import (
	"sync"

	"github.com/0chain/gosdk/core/transaction"
	"gopkg.in/cheggaaa/pb.v1"
)

// StatusBar is to check status of any operation
type StatusBar struct {
	b       *pb.ProgressBar
	wg      *sync.WaitGroup
	success bool
	err     error
}

// Started for statusBar
func (s *StatusBar) Started(allocationID, filePath string, op int, totalBytes int) {
	if logEnabled {
		s.b = pb.StartNew(totalBytes)
		s.b.Set(0)
	}
}

// InProgress for statusBar
func (s *StatusBar) InProgress(allocationID, filePath string, op int, completedBytes int, todo_name_var []byte) {
	if logEnabled && s.b != nil {
		s.b.Set(completedBytes)
	}
}

// Completed for statusBar
func (s *StatusBar) Completed(allocationID, filePath string, filename string, mimetype string, size int, op int) {
	if logEnabled && s.b != nil {
		s.b.Finish()
	}
	s.success = true

	defer s.wg.Done()
	sdkLogger.Info("Status completed callback. Type = " + mimetype + ". Name = " + filename)
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
	PrintError("Error in file operation." + err.Error())
	s.wg.Done()
}

// CommitMetaCompleted when commit meta completes
func (s *StatusBar) CommitMetaCompleted(request, response string, txn *transaction.Transaction, err error) {
	setLastMetadataCommitTxn(txn, err)
	s.wg.Done()
}

// RepairCompleted when repair is completed
func (s *StatusBar) RepairCompleted(filesRepaired int) {
	s.wg.Done()
}
