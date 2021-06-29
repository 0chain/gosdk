package main

import (
	"fmt"
	"strconv"
	"syscall/js"
	"github.com/0chain/gosdk/zcncore"
)

// JS does not have int64 so we must take a string instead of int64.
func strToInt64(s string) int64 {
	tokens, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return tokens
}

func int64ToStr(x int64) string {
	return strconv.FormatInt(x, 10)
}

func TokensToEth(this js.Value, p []js.Value) interface{} {
	tokens := strToInt64(p[0].String())
	result := zcncore.TokensToEth(tokens)
	return result
}

func EthToTokens(this js.Value, p []js.Value) interface{} {
	tokens := p[0].Float()
	result := zcncore.EthToTokens(tokens)
	return int64ToStr(result)
}

func GTokensToEth(this js.Value, p []js.Value) interface{} {
	tokens := strToInt64(p[0].String())
	result := zcncore.GTokensToEth(tokens)
	return result
}

func GEthToTokens(this js.Value, p []js.Value) interface{} {
	tokens := p[0].Float()
	result := zcncore.GEthToTokens(tokens)
	return int64ToStr(result)
}

func GetWalletAddrFromEthMnemonic(this js.Value, p []js.Value) interface{} {
	mnemonic := p[0].String()
	result, err := zcncore.GetWalletAddrFromEthMnemonic(mnemonic)
	if err != nil {
		fmt.Println("error:", err)
	}
	return result
}

// func GetEthBalance(this js.Value, p []js.Value) interface{} {
// 	;
// }

func ConvertZcnTokenToETH(this js.Value, p []js.Value) interface{} {
	token := p[0].Float()
	result, err := zcncore.ConvertZcnTokenToETH(token)
	if err != nil {
		fmt.Println("error:", err)
	}
	return result
}

func SuggestEthGasPrice(this js.Value, p []js.Value) interface{} {
	result, err := zcncore.SuggestEthGasPrice()
	if err != nil {
		fmt.Println("error:", err)
	}
	return int64ToStr(result)
}

func TransferEthTokens(this js.Value, p []js.Value) interface{} {
	fromPrivKey := p[0].String()
	amountTokens := strToInt64(p[1].String())
	gasPrice := strToInt64(p[2].String())
	result, err := zcncore.TransferEthTokens(fromPrivKey, amountTokens, gasPrice)
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
