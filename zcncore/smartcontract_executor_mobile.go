//go:build mobile
// +build mobile

package zcncore

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/0chain/errors"
)

type SmartContractExecutor interface {
	// ExecuteSmartContract implements wrapper for smart contract function
	ExecuteSmartContract(address, methodName string, input string, val uint64) error
}

func (t *Transaction) ExecuteSmartContract(address, methodName string, input string, val uint64) error {
	err := t.createSmartContractTxn(address, methodName, json.RawMessage(input), val)
	if err != nil {
		return err
	}
	go func() {
		t.setNonceAndSubmit()
	}()
	return nil
}

func (ta *TransactionWithAuth) ExecuteSmartContract(address, methodName string, input string, val uint64) error {
	err := ta.t.createSmartContractTxn(address, methodName, json.RawMessage(input), val)
	if err != nil {
		return err
	}
	go func() {
		ta.submitTxn()
	}()
	return nil
}

// NewTransaction allocation new generic transaction object for any operation
func NewTransaction(cb TransactionCallback, txnFee string, nonce int64) (TransactionScheme, error) {
	fee, err := strconv.ParseUint(txnFee, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction fee value: %v", txnFee)
	}

	if fee/uint64(TOKEN_UNIT) == 0 {
		return nil, fmt.Errorf("transaction fee must be multiple value of 1e10")
	}

	err = CheckConfig()
	if err != nil {
		return nil, err
	}
	if _config.isSplitWallet {
		if _config.authUrl == "" {
			return nil, errors.New("", "auth url not set")
		}
		Logger.Info("New transaction interface with auth")
		return newTransactionWithAuth(cb, fee, nonce)
	}
	Logger.Info("New transaction interface")
	t, err := newTransaction(cb, fee, nonce)
	return t, err
}

// ConvertToValue converts ZCN tokens to value
func ConvertToValue(token float64) string {
	return strconv.FormatUint(uint64(token*float64(TOKEN_UNIT)), 10)
}
