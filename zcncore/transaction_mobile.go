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
