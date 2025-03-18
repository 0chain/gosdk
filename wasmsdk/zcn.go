//go:build js && wasm
// +build js,wasm

package main

import (
	"github.com/0chain/gosdk_common/core/screstapi"
	"github.com/0chain/gosdk_common/zcncore"
)

type Balance struct {
	ZCN   float64 `json:"zcn"`
	USD   float64 `json:"usd"`
	Nonce int64   `json:"nonce"`
}

// getWalletBalance retrieves the wallet balance of the client from the network.
//   - clientId is the client id
func getWalletBalance(clientId string) (*Balance, error) {
	bal, err := screstapi.GetBalance(clientId)
	if err != nil {
		return nil, err
	}
	balance, err := bal.ToToken()
	if err != nil {
		return nil, err
	}

	toUsd, err := zcncore.ConvertTokenToUSD(balance)
	if err != nil {
		return nil, err
	}

	return &Balance{
		ZCN:   balance,
		USD:   toUsd,
		Nonce: bal.Nonce,
	}, nil
}
