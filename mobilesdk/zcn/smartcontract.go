//go:build mobile
// +build mobile

package zcn

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/0chain/gosdk/zcncore"
)

// Faucet
func Faucet(methodName, jsonInput string, zcnToken float64) (string, error) {
	return ExecuteSmartContract(zcncore.FaucetSmartContractAddress, methodName, jsonInput, zcncore.ConvertToValue(zcnToken))
}

func ExecuteSmartContract(address, methodName, input, sasToken string) (string, error) {
	wg := &sync.WaitGroup{}
	cb := &transactionCallback{wg: wg}
	txn, err := zcncore.NewTransaction(cb, "0", 0)
	if err != nil {
		return "", err
	}

	wg.Add(1)

	err = txn.ExecuteSmartContract(address, methodName, input, sasToken)
	if err != nil {
		return "", err

	}

	wg.Wait()

	if !cb.success {
		return "", fmt.Errorf("smartcontract: %s", cb.errMsg)
	}

	cb.success = false
	wg.Add(1)
	err = txn.Verify()
	if err != nil {
		return "", err
	}

	wg.Wait()

	if !cb.success {
		return "", fmt.Errorf("smartcontract: %s", cb.errMsg)
	}

	switch txn.GetVerifyConfirmationStatus() {
	case zcncore.ChargeableError:
		return "", fmt.Errorf("smartcontract: %s", txn.GetVerifyOutput())
	case zcncore.Success:
		js, _ := json.Marshal(cb.txn)
		return string(js), nil
	}

	return "", fmt.Errorf("smartcontract: %v", txn.GetVerifyConfirmationStatus())
}

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
