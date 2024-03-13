//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"errors"

	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"
)

type Balance struct {
	ZCN float64 `json:"zcn"`
	USD float64 `json:"usd"`
}

func getWalletBalance(clientId string) (*Balance, error) {

	zcn, err := zcncore.GetWalletBalance(clientId)
	if err != nil {
		return nil, err
	}

	zcnToken, err := zcn.ToToken()
	if err != nil {
		return nil, err
	}

	usd, err := zcncore.ConvertTokenToUSD(zcnToken)
	if err != nil {
		return nil, err
	}

	return &Balance{
		ZCN: zcnToken,
		USD: usd,
	}, nil
}

type ClientState struct {
	TxnHash string      `json:"txn"`
	Round   int64       `json:"round"`
	Balance uint64 		`json:"balance"`
	Nonce   int64       `json:"nonce"`
}

func getClientState(clientId string) (*ClientState, error) {
	clientStateStr, err := zcncore.GetClientState(clientId)
	sdkLogger.Debug("clientStateStr: ", clientStateStr, "err: ", err)
	if err != nil {
		return nil, err
	}
	var clientState ClientState
	err = json.Unmarshal([]byte(clientStateStr), &clientState)
	if err != nil {
		return nil, errors.New(`error unmarshaling client state: ` + err.Error())
	}
	sdkLogger.Debug(clientState)
	return &clientState, nil
}

func createReadPool() (string, error) {
	hash, _, err := sdk.CreateReadPool()
	return hash, err
}
