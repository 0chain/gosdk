package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zcncore"
)

func faucet(methodName, input string, token float64) (*transaction.Transaction, error) {
	return executeSmartContract(zcncore.FaucetSmartContractAddress, methodName, input, zcncore.ConvertToValue(token))
}

// executeSmartContract issue a smart contract transaction
//   - address is the smart contract address
//   - methodName is the method name to be called
//   - input is the input data for the method
//   - value is the value to be sent with the transaction
func executeSmartContract(address, methodName, input string, value uint64) (*transaction.Transaction, error) {
	wg := &sync.WaitGroup{}
	cb := &transactionCallback{wg: wg}
	txn, err := zcncore.NewTransaction(cb, 0, 0)
	if err != nil {
		return nil, err
	}

	wg.Add(1)
	t, err := txn.ExecuteSmartContract(address, methodName, json.RawMessage([]byte(input)), value)
	if err != nil {
		return nil, err

	}

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
		return t, nil
	}

	return nil, fmt.Errorf("smartcontract: %v", txn.GetVerifyConfirmationStatus())
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
