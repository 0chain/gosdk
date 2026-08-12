//go:build js && wasm
// +build js,wasm

package main

import (
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zcncore"
)

// faucet calls the faucet smart contract to pour tokens into the wallet.
//   - methodName is the method to call (e.g. "pour")
//   - input is the input data (e.g. JSON string)
//   - token is the number of ZCN tokens to request
func faucet(methodName, input string, token float64) (*transaction.Transaction, error) {
	_, _, _, txn, err := zcncore.Faucet(zcncore.ConvertToValue(token), input)
	return txn, err
}
