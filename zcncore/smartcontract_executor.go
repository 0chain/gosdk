//go:build !mobile
// +build !mobile

package zcncore

import "github.com/0chain/errors"

type SmartContractExecutor interface {
	// ExecuteSmartContract implements wrapper for smart contract function
	ExecuteSmartContract(address, methodName string, input interface{}, val uint64) error
}

func (t *Transaction) ExecuteSmartContract(address, methodName string, input interface{}, val uint64) error {
	err := t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return err
	}
	go func() {
		t.setNonceAndSubmit()
	}()
	return nil
}

func (ta *TransactionWithAuth) ExecuteSmartContract(address, methodName string, input interface{}, val uint64) error {
	err := ta.t.createSmartContractTxn(address, methodName, input, val)
	if err != nil {
		return err
	}
	go func() {
		ta.submitTxn()
	}()
	return nil
}

// NewTransaction allocation new generic transaction object for any operation
func NewTransaction(cb TransactionCallback, txnFee uint64, nonce int64) (TransactionScheme, error) {
	err := CheckConfig()
	if err != nil {
		return nil, err
	}
	if _config.isSplitWallet {
		if _config.authUrl == "" {
			return nil, errors.New("", "auth url not set")
		}
		Logger.Info("New transaction interface with auth")
		return newTransactionWithAuth(cb, txnFee, nonce)
	}
	Logger.Info("New transaction interface")
	t, err := newTransaction(cb, txnFee, nonce)
	return t, err
}

// ConvertToValue converts ZCN tokens to value
func ConvertToValue(token float64) uint64 {
	return uint64(token * float64(TOKEN_UNIT))
}
