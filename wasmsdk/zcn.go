//go:build js && wasm
// +build js,wasm

package main

import "github.com/0chain/gosdk/zcncore"

type Balance struct {
	ZCN float64 `json:"zcn"`
	USD float64 `json:"usd"`
}

func getWalletBalance(clientId string) (*Balance, error) {

	zcn, err := zcncore.GetWalletBalance(clientId)
	if err != nil {
		return nil, err
	}

	usd, err := zcncore.ConvertTokenToUSD(zcn.ToToken())
	if err != nil {
		return nil, err
	}

	return &Balance{
		ZCN: zcn.ToToken(),
		USD: usd,
	}, nil
}
