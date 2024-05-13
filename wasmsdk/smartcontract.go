package main

import (
	"encoding/json"
	"strconv"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"
)

func faucet(methodName, input string, token float64) (*transaction.Transaction, error) {
	tv, err := strconv.ParseUint(zcncore.ConvertToValue(token), 10, 64)
	if err != nil {
		return nil, err
	}
	return executeSmartContract(zcncore.FaucetSmartContractAddress, methodName, input, tv)
}

func executeSmartContract(address, methodName, input string, value uint64) (*transaction.Transaction, error) {
	return sdk.ExecuteSmartContract(address,
		transaction.SmartContractTxnData{
			Name:      methodName,
			InputArgs: json.RawMessage([]byte(input)),
		}, value, 0)
}
