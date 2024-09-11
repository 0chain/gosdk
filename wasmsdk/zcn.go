//go:build js && wasm
// +build js,wasm

package main

import (
	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"
)

type Balance struct {
	ZCN   float64 `json:"zcn"`
	USD   float64 `json:"usd"`
	Nonce int64   `json:"nonce"`
}

// getWalletBalance retrieves the wallet balance of the client from the network.
//   - clientId is the client id
func getWalletBalance(clientId string) (*Balance, error) {
	bal, err := client.GetBalance(clientId)
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

// createReadPool creates a read pool for the client where they should lock tokens to be able to read data.
func createReadPool() (string, error) {
	hash, _, err := sdk.CreateReadPool()
	return hash, err
}
