package wallet

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/0chain/gosdk/zcncore"
)

// ZCNStatus represents the status of a ZCN operation.
type ZCNStatus struct {
	walletString string
	balance      int64
	value        interface{}
	Wg           *sync.WaitGroup
	Success      bool
	Err          error
}

// NewZCNStatus creates a new ZCNStatus instance.
//   - value: value to be stored in the ZCNStatus instance
func NewZCNStatus(value interface{}) (zcns *ZCNStatus) {
	return &ZCNStatus{
		Wg:    new(sync.WaitGroup),
		value: value,
	}
}

// Begin starts the wait group
func (zcn *ZCNStatus) Begin() {
	zcn.Wg.Add(1)
}

// Wait waits for the wait group to finish
func (zcn *ZCNStatus) Wait() error {
	zcn.Wg.Wait()
	return zcn.Err
}

// OnBalanceAvailable callback when balance is available
//   - status: status of the operation
//   - value: balance value
//   - third parameter is not used, it is kept for compatibility
func (zcn *ZCNStatus) OnBalanceAvailable(status int, value int64, _ string) {
	defer zcn.Wg.Done()
	if status == zcncore.StatusSuccess {
		zcn.Success = true
	} else {
		zcn.Success = false
	}
	zcn.balance = value
}

// OnTransactionComplete callback when a transaction is completed
//   - t: transaction object
//   - status: status of the transaction
func (zcn *ZCNStatus) OnTransactionComplete(t *zcncore.Transaction, status int) {
	defer zcn.Wg.Done()
	if status == zcncore.StatusSuccess {
		zcn.Success = true
	} else {
		zcn.Err = errors.New(t.GetTransactionError())
	}
}

// OnVerifyComplete callback when a transaction is verified
//   - t: transaction object
//   - status: status of the transaction
func (zcn *ZCNStatus) OnVerifyComplete(t *zcncore.Transaction, status int) {
	defer zcn.Wg.Done()
	if status == zcncore.StatusSuccess {
		zcn.Success = true
	} else {
		zcn.Err = errors.New(t.GetVerifyError())
	}
}

// OnTransferComplete callback when a transfer is completed. Not used in this implementation
func (zcn *ZCNStatus) OnAuthComplete(_ *zcncore.Transaction, status int) {
	Logger.Info("Authorization complete with status: ", status)
}

// OnWalletCreateComplete callback when a wallet is created
//   - status: status of the operation
//   - wallet: wallet json string
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

// OnInfoAvailable callback when information is available
//   - op`: operation type (check `zcncore.Op* constants)
//   - status: status of the operation
//   - info: information represneted as a string
//   - err: error message
func (zcn *ZCNStatus) OnInfoAvailable(op int, status int, info string, err string) {
	defer zcn.Wg.Done()

	// If status is 400 for OpGetMintNonce, mintNonce is considered as 0
	if op == zcncore.OpGetMintNonce && status == http.StatusBadRequest {
		zcn.Err = nil
		zcn.Success = true
		return
	}

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

	var errm error
	if errm = json.Unmarshal([]byte(info), zcn.value); errm != nil {
		zcn.Err = fmt.Errorf("decoding response: %v", errm)
		zcn.Success = false
		return
	}

	zcn.Err = nil
	zcn.Success = true
}

// OnSetupComplete callback when setup is completed.
// Paramters are not used in this implementation,
// just kept for compatibility.
func (zcn *ZCNStatus) OnSetupComplete(_ int, _ string) {
	defer zcn.Wg.Done()
}

// OnAuthorizeSendComplete callback when authorization is completed
//   - status: status of the operation
//   - 2nd parameter is not used, it is kept for compatibility
//   - 3rd parameter is not used, it is kept for compatibility
//   - 4th parameter is not used, it is kept for compatibility
//   - creationDate: timestamp of the creation date
//   - signature: signature of the operation
func (zcn *ZCNStatus) OnAuthorizeSendComplete(status int, _ string, _ int64, _ string, creationDate int64, signature string) {
	defer zcn.Wg.Done()

	Logger.Info("Status: ", status)
	Logger.Info("Timestamp:", creationDate)
	Logger.Info("Signature:", signature)
}

// OnVoteComplete callback when a multisig vote is completed
//   - status: status of the operation
//   - proposal: proposal json string
//   - err: error message
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
