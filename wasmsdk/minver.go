//go:build minver
// +build minver

package main

import (
	"fmt"
	"syscall/js"

	"github.com/0chain/gosdk/core/version"
)

func main() {
	fmt.Printf("0CHAIN - GOSDK COMPACT (version=%v)\n", version.VERSIONSTR)

	c := make(chan struct{}, 0)

	// Just functions for 0proxy.
	js.Global().Set("initializeConfig", js.FuncOf(InitializeConfig))
	js.Global().Set("initStorageSDK", js.FuncOf(InitStorageSDK))
	js.Global().Set("Upload", js.FuncOf(Upload))
	js.Global().Set("Download", js.FuncOf(Download))
	js.Global().Set("Share", js.FuncOf(Share))
	js.Global().Set("Rename", js.FuncOf(Rename))
	js.Global().Set("Copy", js.FuncOf(Copy))
	js.Global().Set("Delete", js.FuncOf(Delete))
	js.Global().Set("Move", js.FuncOf(Move))
	js.Global().Set("GetClientEncryptedPublicKey", js.FuncOf(GetClientEncryptedPublicKey))

	// ethwallet.go
	js.Global().Set("TokensToEth", js.FuncOf(TokensToEth))
	js.Global().Set("EthToTokens", js.FuncOf(EthToTokens))
	js.Global().Set("GTokensToEth", js.FuncOf(GTokensToEth))
	js.Global().Set("GEthToTokens", js.FuncOf(GEthToTokens))
	js.Global().Set("ConvertZcnTokenToETH", js.FuncOf(ConvertZcnTokenToETH))
	js.Global().Set("SuggestEthGasPrice", js.FuncOf(SuggestEthGasPrice))
	js.Global().Set("TransferEthTokens", js.FuncOf(TransferEthTokens))
	js.Global().Set("GetWalletAddrFromEthMnemonic", js.FuncOf(GetWalletAddrFromEthMnemonic))
	js.Global().Set("IsValidEthAddress", js.FuncOf(IsValidEthAddress))
	js.Global().Set("CreateWalletFromEthMnemonic", js.FuncOf(CreateWalletFromEthMnemonic))

	// zboxsdk_min.go
	js.Global().Set("InitAuthTicket", js.FuncOf(InitAuthTicket))
	js.Global().Set("GetClientEncryptedPublicKey", js.FuncOf(GetClientEncryptedPublicKey))
	js.Global().Set("GetAllocation", js.FuncOf(GetAllocation))
	js.Global().Set("SetNumBlockDownloads", js.FuncOf(SetNumBlockDownloads))
	js.Global().Set("GetAllocations", js.FuncOf(GetAllocations))
	js.Global().Set("GetAllocationFromAuthTicket", js.FuncOf(GetAllocationFromAuthTicket))
	<-c
}
