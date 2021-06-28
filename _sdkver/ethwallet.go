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

// Exports public functions in github.com/0chain/gosdk/zcncore/ethwallet.go
func CheckEthHashStatus(this js.Value, p []js.Value) interface{} {
	hash := p[0].String()
	status := zcncore.CheckEthHashStatus(hash)
	return status
}
