package wallet

import (
	"fmt"
	"os"
	"sync"

	"github.com/0chain/gosdk/zcncore"
	"gopkg.in/cheggaaa/pb.v1"
)

type StatusBar struct {
	b  *pb.ProgressBar
	wg *sync.WaitGroup
}

type ZCNStatus struct {
	walletString string
	Wg           *sync.WaitGroup
	Success      bool
	ErrMsg       string
	balance      int64
}

func NewZCNStatus() (zcns *ZCNStatus) {
	return &ZCNStatus{Wg: new(sync.WaitGroup)}
}

func (zcn *ZCNStatus) Begin() { zcn.Wg.Add(1) }
func (zcn *ZCNStatus) Wait()  { zcn.Wg.Wait() }

func (zcn *ZCNStatus) OnBalanceAvailable(status int, value int64, _ string) {
	defer zcn.Wg.Done()
	if status == zcncore.StatusSuccess {
		zcn.Success = true
	} else {
		zcn.Success = false
	}
	zcn.balance = value
}

func (zcn *ZCNStatus) OnTransactionComplete(t *zcncore.Transaction, status int) {
	defer zcn.Wg.Done()
	if status == zcncore.StatusSuccess {
		zcn.Success = true
	} else {
		zcn.ErrMsg = t.GetTransactionError()
	}
}

func (zcn *ZCNStatus) OnVerifyComplete(t *zcncore.Transaction, status int) {
	defer zcn.Wg.Done()
	if status == zcncore.StatusSuccess {
		zcn.Success = true
	} else {
		zcn.ErrMsg = t.GetVerifyError()
	}
}

func (zcn *ZCNStatus) OnAuthComplete(_ *zcncore.Transaction, status int) {
	fmt.Println("Authorization complete on zauth.", status)
}

func (zcn *ZCNStatus) OnWalletCreateComplete(status int, wallet string, err string) {
	defer zcn.Wg.Done()
	if status != zcncore.StatusSuccess {
		zcn.Success = false
		zcn.ErrMsg = err
		zcn.walletString = ""
		return
	}
	zcn.Success = true
	zcn.ErrMsg = ""
	zcn.walletString = wallet
}

func (zcn *ZCNStatus) OnInfoAvailable(_ int, status int, config string, err string) {
	defer zcn.Wg.Done()
	if status != zcncore.StatusSuccess {
		zcn.Success = false
		zcn.ErrMsg = err
		return
	}
	zcn.Success = true
	zcn.ErrMsg = config
}

func (zcn *ZCNStatus) OnSetupComplete(_ int, _ string) {
	defer zcn.Wg.Done()
}

func (zcn *ZCNStatus) OnAuthorizeSendComplete(status int, _ string, _ int64, _ string, creationDate int64, signature string) {
	defer zcn.Wg.Done()
	fmt.Println("Status:", status)
	fmt.Println("Timestamp:", creationDate)
	fmt.Println("Signature:", signature)
}

// OnVoteComplete callback when a multisig vote is completed
func (zcn *ZCNStatus) OnVoteComplete(status int, proposal string, err string) {
	defer zcn.Wg.Done()
	if status != zcncore.StatusSuccess {
		zcn.Success = false
		zcn.ErrMsg = err
		zcn.walletString = ""
		return
	}
	zcn.Success = true
	zcn.ErrMsg = ""
	zcn.walletString = proposal
}

//goland:noinspection GoUnusedExportedFunction
func PrintError(v ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, v...)
}

//goland:noinspection GoUnusedExportedFunction
func ExitWithError(v ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}
