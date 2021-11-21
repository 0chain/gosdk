package wallet

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zcncore"
	"github.com/spf13/pflag"
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
	balance      common.Balance
	wallets      []string
	clientID     string
}

func NewZCNStatus() (zcns *ZCNStatus) {
	return &ZCNStatus{Wg: new(sync.WaitGroup)}
}

func (zcn *ZCNStatus) Begin() { zcn.Wg.Add(1) }
func (zcn *ZCNStatus) Wait()  { zcn.Wg.Wait() }

func (zcn *ZCNStatus) OnBalanceAvailable(status int, value int64, info string) {
	defer zcn.Wg.Done()
	if status == zcncore.StatusSuccess {
		zcn.Success = true
	} else {
		zcn.Success = false
	}
	zcn.balance = common.Balance(value)
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

func (zcn *ZCNStatus) OnAuthComplete(t *zcncore.Transaction, status int) {
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

func (zcn *ZCNStatus) OnSetupComplete(status int, err string) {
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

func PrintError(v ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, v...)
}

func ExitWithError(v ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}

func setupInputMap(flags *pflag.FlagSet, sKeys, sValues string) map[string]string {
	var err error
	var keys []string
	if flags.Changed(sKeys) {
		keys, err = flags.GetStringSlice(sKeys)
		if err != nil {
			log.Fatal(err)
		}
	}

	var values []string
	if flags.Changed(sValues) {
		values, err = flags.GetStringSlice(sValues)
		if err != nil {
			log.Fatal(err)
		}
	}

	input := make(map[string]string)
	if len(keys) != len(values) {
		log.Fatal("number " + sKeys + " must equal the number " + sValues)
	}
	for i := 0; i < len(keys); i++ {
		v := strings.TrimSpace(values[i])
		k := strings.TrimSpace(keys[i])
		input[k] = v
	}
	return input
}

func printMap(outMap map[string]string) {
	keys := make([]string, 0, len(outMap))
	for k := range outMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Println(k, "\t", outMap[k])
	}
}
