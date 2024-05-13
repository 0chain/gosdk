//go:build !mobile
// +build !mobile

package main

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"
)

func faucet(methodName, input string, token float64) (*transaction.Transaction, error) {
	return executeSmartContract(zcncore.FaucetSmartContractAddress, methodName, input, zcncore.ConvertToValue(token))
}

func executeSmartContract(address, methodName, input string, value uint64) (*transaction.Transaction, error) {
	return sdk.ExecuteSmartContract(address,
		transaction.SmartContractTxnData{
			Name:      methodName,
			InputArgs: json.RawMessage([]byte(input)),
		}, value, 0)
}
