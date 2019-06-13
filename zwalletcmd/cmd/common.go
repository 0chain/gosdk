package cmd

import (
	"sync"

	"0chain.net/clientsdk/zcncore"
	"gopkg.in/cheggaaa/pb.v1"
)

type StatusBar struct {
	b  *pb.ProgressBar
	wg *sync.WaitGroup
}

type ZCNStatus struct {
	walletString string
	wg           *sync.WaitGroup
	success      bool
	errMsg       string
	balance      int64
}

func (zcn *ZCNStatus) OnBalanceAvailable(status int, value int64) {
	defer zcn.wg.Done()
	if status == zcncore.StatusSuccess {
		zcn.success = true
	} else {
		zcn.success = false
	}
	zcn.balance = value
}

func (zcn *ZCNStatus) OnTransactionComplete(t *zcncore.Transaction, status int) {
	defer zcn.wg.Done()
	if status == zcncore.StatusSuccess {
		zcn.success = true
	} else {
		zcn.errMsg = t.GetTransactionError()
	}
	// fmt.Println("Txn Hash:", t.GetTransactionHash())
}

func (zcn *ZCNStatus) OnVerifyComplete(t *zcncore.Transaction, status int) {
	defer zcn.wg.Done()
	if status == zcncore.StatusSuccess {
		zcn.success = true
	} else {
		zcn.errMsg = t.GetVerifyError()
	}
	// fmt.Println(t.GetVerifyOutput())
}

func (zcn *ZCNStatus) OnWalletCreateComplete(status int, wallet string, err string) {
	defer zcn.wg.Done()
	if status != zcncore.StatusSuccess {
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

func (zcn *ZCNStatus) OnInfoAvailable(Op int, status int, config string, err string) {
	defer zcn.wg.Done()
	if status != zcncore.StatusSuccess {
		zcn.success = false
		zcn.errMsg = err
		return
	}
	zcn.success = true
	zcn.errMsg = config
}
