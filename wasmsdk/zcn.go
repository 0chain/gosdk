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
	TxnHash string  `json:"txn"`
	Round   int64   `json:"round"`
	Nonce   int64   `json:"nonce"`
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

	clientStateStr, err := zcncore.GetClientState(clientId)
	if err != nil {
		return nil, err
	}
	var clientState struct {
		TxnHash string  `json:"txn"`
		Round   int64   `json:"round"`
		Nonce   int64   `json:"nonce"`
	}
	err = json.Unmarshal([]byte(clientStateStr), &clientState)
	if err != nil {
		return nil, errors.New(`error unmarshaling client state: ` + err.Error())
	}
	return &Balance{
		ZCN: zcnToken,
		USD: usd,
		TxnHash: clientState.TxnHash,
		Round: clientState.Round,
		Nonce: clientState.Nonce,
	}, nil
}

func createReadPool() (string, error) {
	hash, _, err := sdk.CreateReadPool()
	return hash, err
}
