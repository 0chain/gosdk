//go:build mobile
// +build mobile

package zcncore

import (
	"encoding/json"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/transaction"
)

func (t *Transaction) ExecuteSmartContract(address, methodName string, input string, val int64) error {
	err := t.createSmartContractTxnWithJSON(address, methodName, input, val)
	if err != nil {
		return err
	}
	go func() {
		t.setNonceAndSubmit()
	}()
	return nil
}

func (t *Transaction) createSmartContractTxnWithJSON(address, methodName string, jsonInput string, value int64) error {

	var sn transaction.SmartContractTxnData

	snBytes, err := json.Marshal(sn)
	if err != nil {
		return errors.Wrap(err, "create smart contract failed due to invalid data.")
	}
	t.txn.TransactionType = transaction.TxnTypeSmartContract
	t.txn.ToClientID = address
	t.txn.TransactionData = string(snBytes)
	t.txn.Value = value
	return nil
}
