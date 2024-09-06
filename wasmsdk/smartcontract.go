package main

import (
	"encoding/json"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"
)

func faucet(methodName, input string, token float64) (*transaction.Transaction, error) {
	return executeSmartContract(zcncore.FaucetSmartContractAddress,
		methodName, input, zcncore.ConvertToValue(token))
}

// executeSmartContract issue a smart contract transaction
//   - address is the smart contract address
//   - methodName is the method name to be called
//   - input is the input data for the method
//   - value is the value to be sent with the transaction
func executeSmartContract(address, methodName, input string, value uint64) (*transaction.Transaction, error) {
	return sdk.ExecuteSmartContract(address,
		transaction.SmartContractTxnData{
			Name:      methodName,
			InputArgs: json.RawMessage([]byte(input)),
		}, value, 0)
}
