package main

import (
	"fmt"
	"syscall/js"
	"github.com/0chain/gosdk/zcncore"
)

// Exports public functions in github.com/0chain/gosdk/zcncore/ethwallet.go
func IsValidEthAddress(this js.Value, p []js.Value) interface{} {
	ethAddr := p[0].String()
	success, err := zcncore.IsValidEthAddress(ethAddr)
	if err != nil {
		fmt.Println("error:", err)
	}
	return success
}

// func TokensToEth(this js.Value, p []js.Value) interface{} {
// 	;
// }

func CreateWalletFromEthMnemonic(this js.Value, p []js.Value) interface{} {
	mnemonic := p[0].String()
	password := p[1].String()

	err := zcncore.CreateWalletFromEthMnemonic(mnemonic, password, nil)
	if err != nil {
		fmt.Println("error:", err)
	}

	return nil
}

// Exports public functions in github.com/0chain/gosdk/zcncore/ethwallet.go
func CheckEthHashStatus(this js.Value, p []js.Value) interface{} {
	hash := p[0].String()
	status := zcncore.CheckEthHashStatus(hash)
	return status
}
