package sdk

import (
	"fmt"
	"strconv"
	"sync"

	"errors"

	"github.com/0chain/gosdk/core/transaction"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zcncore"
)

type transactionCallback struct {
	wg      *sync.WaitGroup
	success bool
	errMsg  string

	txn *zcncore.Transaction
}

func (cb *transactionCallback) OnTransactionComplete(t *zcncore.Transaction, status int) {
	defer cb.wg.Done()
	cb.txn = t
	if status == zcncore.StatusSuccess {
		cb.success = true
	} else {
		cb.errMsg = t.GetTransactionError()
	}
}

func (cb *transactionCallback) OnVerifyComplete(t *zcncore.Transaction, status int) {
	defer cb.wg.Done()
	cb.txn = t
	if status == zcncore.StatusSuccess {
		cb.success = true
	} else {
		cb.errMsg = t.GetVerifyError()
	}
}

func (cb *transactionCallback) OnAuthComplete(t *zcncore.Transaction, status int) {
	cb.txn = t
	fmt.Println("Authorization complete on zauth.", status)
}

// ExecuteSmartContract executes the smart contract
func ExecuteSmartContract(address string, sn transaction.SmartContractTxnData, value, fee uint64) (*transaction.Transaction, error) {
	wg := &sync.WaitGroup{}
	cb := &transactionCallback{wg: wg}
	txn, err := zcncore.NewTransaction(cb, strconv.FormatUint(fee, 10), 0)
	if err != nil {
		return nil, err
	}

	wg.Add(1)
	err = txn.ExecuteSmartContract(address, sn.Name, sn.InputArgs, strconv.FormatUint(value, 10))
	if err != nil {
		return nil, err
	}

	msg := fmt.Sprintf("Executing transaction '%s' with hash %s ", sn.Name, txn.Hash())
	l.Logger.Info(msg)
	l.Logger.Info("estimated txn fee: ", txn.Fee())

	wg.Wait()

	if !cb.success {
		return nil, fmt.Errorf("smartcontract: %s", cb.errMsg)
	}

	cb.success = false
	wg.Add(1)
	err = txn.Verify()
	if err != nil {
		return nil, err
	}

	wg.Wait()

	if !cb.success {
		return nil, fmt.Errorf("smartcontract: %s", cb.errMsg)
	}

	switch txn.GetVerifyConfirmationStatus() {
	case zcncore.ChargeableError:
		return nil, fmt.Errorf("smartcontract: %s", txn.GetVerifyOutput())
	case zcncore.Success:
		return txn.Txn(), nil
	}

	return nil, fmt.Errorf("smartcontract: %v", txn.GetVerifyConfirmationStatus())
}

// ExecuteSmartContractSend create send transaction to transfer tokens from the caller to target address
func ExecuteSmartContractSend(to string, tokens, fee uint64, desc string) (string, error) {
	wg := &sync.WaitGroup{}
	cb := &transactionCallback{wg: wg}
	txn, err := zcncore.NewTransaction(cb, strconv.FormatUint(fee, 10), 0)
	if err != nil {
		return "", err
	}

	wg.Add(1)
	err = txn.Send(to, strconv.FormatUint(tokens, 10), desc)
	if err == nil {
		wg.Wait()
	} else {
		return "", err
	}

	if cb.success {
		cb.success = false
		wg.Add(1)
		err := txn.Verify()
		if err == nil {
			wg.Wait()
		} else {
			return "", err
		}
		if cb.success {
			return txn.GetVerifyOutput(), nil
		}
	}

	return "", errors.New(cb.errMsg)
}
