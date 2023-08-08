package wallet

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/0chain/gosdk/zcncore"
)

type ZCNStatus struct {
	walletString string
	balance      int64
	value        interface{}
	Wg           *sync.WaitGroup
	Success      bool
	Err          error
}

func NewZCNStatus(value interface{}) (zcns *ZCNStatus) {
	return &ZCNStatus{
		Wg:    new(sync.WaitGroup),
		value: value,
	}
}

func (zcn *ZCNStatus) Begin() {
	zcn.Wg.Add(1)
}

func (zcn *ZCNStatus) Wait() error {
	zcn.Wg.Wait()
	return zcn.Err
}

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
		zcn.Err = errors.New(t.GetTransactionError())
	}
}

func (zcn *ZCNStatus) OnVerifyComplete(t *zcncore.Transaction, status int) {
	defer zcn.Wg.Done()
	if status == zcncore.StatusSuccess {
		zcn.Success = true
	} else {
		zcn.Err = errors.New(t.GetVerifyError())
	}
}

func (zcn *ZCNStatus) OnAuthComplete(_ *zcncore.Transaction, status int) {
	Logger.Info("Authorization complete with status: ", status)
}

func (zcn *ZCNStatus) OnWalletCreateComplete(status int, wallet string, err string) {
	defer zcn.Wg.Done()
	if status != zcncore.StatusSuccess {
		zcn.Success = false
		zcn.Err = errors.New(err)
		zcn.walletString = ""
		return
	}
	zcn.Success = true
	zcn.Err = nil
	zcn.walletString = wallet
}

func (zcn *ZCNStatus) OnInfoAvailable(_ int, status int, info string, err string) {
	defer zcn.Wg.Done()
	if status != zcncore.StatusSuccess {
		zcn.Err = errors.New(err)
		zcn.Success = false
		return
	}

	if info == "" || info == "{}" {
		zcn.Err = errors.New("empty response")
		zcn.Success = false
		return
	}

	fmt.Printf("OnInfoAvailable got info: %v\n", info)

	var errm error
	if errm = json.Unmarshal([]byte(info), zcn.value); errm != nil {
		zcn.Err = fmt.Errorf("decoding response: %v", errm)
		zcn.Success = false
		return
	}

	zcn.Err = nil
	zcn.Success = true
}

func (zcn *ZCNStatus) OnSetupComplete(_ int, _ string) {
	defer zcn.Wg.Done()
}

func (zcn *ZCNStatus) OnAuthorizeSendComplete(status int, _ string, _ int64, _ string, creationDate int64, signature string) {
	defer zcn.Wg.Done()

	Logger.Info("Status: ", status)
	Logger.Info("Timestamp:", creationDate)
	Logger.Info("Signature:", signature)
}

// OnVoteComplete callback when a multisig vote is completed
func (zcn *ZCNStatus) OnVoteComplete(status int, proposal string, err string) {
	defer zcn.Wg.Done()
	if status != zcncore.StatusSuccess {
		zcn.Success = false
		zcn.Err = errors.New(err)
		zcn.walletString = ""
		return
	}
	zcn.Success = true
	zcn.Err = nil
	zcn.walletString = proposal
}

//goland:noinspection ALL
func PrintError(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
}

//goland:noinspection ALL
func ExitWithError(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}
