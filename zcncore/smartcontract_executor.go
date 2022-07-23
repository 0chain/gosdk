//go:build !mobile
// +build !mobile

package zcncore

import "github.com/0chain/gosdk/core/transaction"

type SmartContractExecutor interface {
	// ExecuteSmartContract implements wrapper for smart contract function
	ExecuteSmartContract(address, methodName string, input interface{}, val uint64) (*transaction.Transaction, error)
}

func (t *Transaction) ExecuteSmartContract(address, methodName string, input interface{}, val uint64) (*transaction.Transaction, error) {
	err := t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return nil, err
	}

	go func() {
		t.setNonceAndSubmit()
	}()
	return t.txn, nil
}

func (ta *TransactionWithAuth) ExecuteSmartContract(address, methodName string, input interface{}, val uint64) (*transaction.Transaction, error) {
	err := ta.t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return nil, err
	}
	go func() {
		ta.submitTxn()
	}()
	return ta.t.txn, nil
}
