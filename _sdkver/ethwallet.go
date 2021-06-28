package main

import (
	"fmt"
	"syscall/js"
	"github.com/0chain/gosdk/zcncore"
)

func GetWalletAddrFromEthMnemonic(this js.Value, p []js.Value) interface{} {
	mnemonic := p[0].String()
	result, err := zcncore.GetWalletAddrFromEthMnemonic(mnemonic)
	if err != nil {
		fmt.Println("error:", err)
	}
	return result
}

// Exports public functions in github.com/0chain/gosdk/zcncore/ethwallet.go
func IsValidEthAddress(this js.Value, p []js.Value) interface{} {
	ethAddr := p[0].String()
	success, err := zcncore.IsValidEthAddress(ethAddr)
	if err != nil {
		fmt.Println("error:", err)
	}
	return success
}

func CheckEthHashStatus(this js.Value, p []js.Value) interface{} {
	hash := p[0].String()
	status := zcncore.CheckEthHashStatus(hash)
	return status
}

// func TokensToEth(this js.Value, p []js.Value) interface{} {
// 	;
// }

func CreateWalletFromEthMnemonic(this js.Value, p []js.Value) interface{} {
	mnemonic := p[0].String()
	password := p[1].String()

	// @Artem you probably want to replace 'nil' with an actual status callback
	// function.
	err := zcncore.CreateWalletFromEthMnemonic(mnemonic, password, nil)
	if err != nil {
		fmt.Println("error:", err)
	}

	return nil
}
