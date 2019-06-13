package cmd

import (
	"fmt"
	"sync"

	"gopkg.in/cheggaaa/pb.v1"
)

const (
	ZCNStatusSuccess int = 0
	ZCNStatusError   int = 1
)

func (s *StatusBar) Started(allocationId, filePath string, op int, totalBytes int) {
	s.b = pb.StartNew(totalBytes)
	s.b.Set(0)
}
func (s *StatusBar) InProgress(allocationId, filePath string, op int, completedBytes int) {
	s.b.Set(completedBytes)
}

func (s *StatusBar) Completed(allocationId, filePath string, filename string, mimetype string, size int, op int) {
	// Not required
	// s.b.PrependElapsed()
	if s.b != nil {
		s.b.Finish()
	}
	defer s.wg.Done()
	fmt.Println("Status completed callback. Type = " + mimetype + ". Name = " + filename)
}

func (s *StatusBar) Error(allocationID string, filePath string, op int, err error) {
	if s.b != nil {
		s.b.Finish()
	}
	defer s.wg.Done()
	fmt.Println("Error in file upload." + err.Error())
}

type StatusBar struct {
	b  *pb.ProgressBar
	wg *sync.WaitGroup
}

type ZCNStatus struct {
	walletString string
	wg           *sync.WaitGroup
	success      bool
	errMsg       string
}

func (zcn *ZCNStatus) OnWalletCreateComplete(status int, wallet string, err string) {
	defer zcn.wg.Done()
	if status == ZCNStatusError {
		zcn.success = false
		zcn.errMsg = err
		zcn.walletString = ""
		return
	}
	zcn.success = true
	zcn.errMsg = ""
	zcn.walletString = wallet
	return
}
