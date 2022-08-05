//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"

	"github.com/0chain/gosdk/zcnswap"
	"github.com/0chain/gosdk/zcnswap/config"
)

func setSwapWallets(usdcTokenAddress, bancorAddress, zcnTokenAddress, ethWalletMnemonic string) {
	config.Configuration = config.SwapConfig{
		UsdcTokenAddress: usdcTokenAddress,
		BancorAddress:    bancorAddress,
		ZcnTokenAddress:  zcnTokenAddress,
		WalletMnemonic:   ethWalletMnemonic,
	}

	fmt.Println("[swap]wallets are initialized")
}

func swapToken(swapAmount int64, tokenSource string) (string, error) {
	return zcnswap.Swap(swapAmount, tokenSource)
}
