//go:build mobile
// +build mobile

package sdk

import (
	"fmt"
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
func ExecuteSmartContract(address string, sn transaction.SmartContractTxnData, value, fee string) (*transaction.Transaction, error) {
	wg := &sync.WaitGroup{}
	cb := &transactionCallback{wg: wg}
	txn, err := zcncore.NewTransaction(cb, fee, 0)
	if err != nil {
		return nil, err
	}

	wg.Add(1)

	inputRaw, ok := sn.InputArgs.(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert input args")
	}

	err = txn.ExecuteSmartContract(address, sn.Name, inputRaw, value)
	if err != nil {
		return nil, err
	}

	t := txn.GetDetails()

	msg := fmt.Sprintf("Executing transaction '%s' with hash %s ", sn.Name, t.Hash)
	l.Logger.Info(msg)
	l.Logger.Info("estimated txn fee: ", t.TransactionFee)

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
		return t, errors.New(txn.GetVerifyOutput())
	case zcncore.Success:
		return t, nil
	}

	return nil, fmt.Errorf("smartcontract: %v", txn.GetVerifyConfirmationStatus())
}

// ExecuteSmartContractSend create send transaction to transfer tokens from the caller to target address
func ExecuteSmartContractSend(to, tokens, fee, desc string) (string, error) {
	wg := &sync.WaitGroup{}
	cb := &transactionCallback{wg: wg}
	txn, err := zcncore.NewTransaction(cb, fee, 0)
	if err != nil {
		return "", err
	}

	wg.Add(1)
	err = txn.Send(to, tokens, desc)
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
