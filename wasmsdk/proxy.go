// +build js,wasm

package main

import (
	"fmt"

	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zcncore"

	"syscall/js"
)

//-----------------------------------------------------------------------------

func main() {
	fmt.Printf("0CHAIN - GOSDK (version=%v)\n", version.VERSIONSTR)

	jsbridge.BindFuncs(map[string]interface{}{
		"GetVersion": zcncore.GetVersion,
		"InitZCNSDK": zcncore.InitZCNSDK,
		"CloseLog":   zcncore.CloseLog,
	})
	js.Global().Set("initializeConfig", js.FuncOf(InitializeConfig))

	// Just functions for 0proxy.
	js.Global().Set("Upload", js.FuncOf(Upload))
	js.Global().Set("Download", js.FuncOf(Download))
	js.Global().Set("Share", js.FuncOf(Share))
	js.Global().Set("Rename", js.FuncOf(Rename))
	js.Global().Set("Copy", js.FuncOf(Copy))
	js.Global().Set("Delete", js.FuncOf(Delete))
	js.Global().Set("Move", js.FuncOf(Move))

	// ethwallet.go
	js.Global().Set("TokensToEth", js.FuncOf(TokensToEth))
	js.Global().Set("EthToTokens", js.FuncOf(EthToTokens))
	js.Global().Set("GTokensToEth", js.FuncOf(GTokensToEth))
	js.Global().Set("GEthToTokens", js.FuncOf(GEthToTokens))
	js.Global().Set("GetEthBalance", js.FuncOf(GetEthBalance))
	js.Global().Set("ConvertZcnTokenToETH", js.FuncOf(ConvertZcnTokenToETH))
	js.Global().Set("SuggestEthGasPrice", js.FuncOf(SuggestEthGasPrice))
	js.Global().Set("TransferEthTokens", js.FuncOf(TransferEthTokens))
	js.Global().Set("GetWalletAddrFromEthMnemonic", js.FuncOf(GetWalletAddrFromEthMnemonic))
	js.Global().Set("IsValidEthAddress", js.FuncOf(IsValidEthAddress))
	js.Global().Set("CreateWalletFromEthMnemonic", js.FuncOf(CreateWalletFromEthMnemonic))

	// wallet.go
	js.Global().Set("GetMinShardersVerify", js.FuncOf(GetMinShardersVerify))
	js.Global().Set("SetLogFile", js.FuncOf(SetLogFile))

	js.Global().Set("SetNetwork", js.FuncOf(SetNetwork))
	js.Global().Set("SplitKeys", js.FuncOf(SplitKeys))
	js.Global().Set("GetNetworkJSON", js.FuncOf(GetNetworkJSON))
	js.Global().Set("CreateWallet", js.FuncOf(CreateWallet))
	js.Global().Set("RecoverWallet", js.FuncOf(RecoverWallet))
	js.Global().Set("SplitKeys", js.FuncOf(SplitKeys))
	js.Global().Set("RegisterToMiners", js.FuncOf(RegisterToMiners))
	js.Global().Set("GetClientDetails", js.FuncOf(GetClientDetails))
	js.Global().Set("IsMnemonicValid", js.FuncOf(IsMnemonicValid))
	js.Global().Set("SetWalletInfo", js.FuncOf(SetWalletInfo))
	js.Global().Set("SetAuthUrl", js.FuncOf(SetAuthUrl))
	js.Global().Set("GetBalance", js.FuncOf(GetBalance))
	js.Global().Set("GetBalanceWallet", js.FuncOf(GetBalanceWallet))
	js.Global().Set("ConvertToToken", js.FuncOf(ConvertToToken))
	js.Global().Set("ConvertToValue", js.FuncOf(ConvertToValue))
	js.Global().Set("ConvertTokenToUSD", js.FuncOf(ConvertTokenToUSD))
	js.Global().Set("ConvertUSDToToken", js.FuncOf(ConvertUSDToToken))
	js.Global().Set("GetLockConfig", js.FuncOf(GetLockConfig))
	js.Global().Set("GetLockedTokens", js.FuncOf(GetLockedTokens))
	js.Global().Set("GetWallet", js.FuncOf(GetWallet))
	js.Global().Set("GetWalletClientID", js.FuncOf(GetWalletClientID))
	js.Global().Set("GetZcnUSDInfo", js.FuncOf(GetZcnUSDInfo))
	js.Global().Set("SetupAuth", js.FuncOf(SetupAuth))
	js.Global().Set("GetIdForUrl", js.FuncOf(GetIdForUrl))
	js.Global().Set("GetVestingPoolInfo", js.FuncOf(GetVestingPoolInfo))
	js.Global().Set("GetVestingClientList", js.FuncOf(GetVestingClientList))
	js.Global().Set("GetVestingSCConfig", js.FuncOf(GetVestingSCConfig))
	js.Global().Set("GetMiners", js.FuncOf(GetMiners))
	js.Global().Set("GetSharders", js.FuncOf(GetSharders))
	js.Global().Set("GetMinerSCNodeInfo", js.FuncOf(GetMinerSCNodeInfo))
	js.Global().Set("GetMinerSCNodePool", js.FuncOf(GetMinerSCNodePool))
	js.Global().Set("GetMinerSCUserInfo", js.FuncOf(GetMinerSCUserInfo))
	js.Global().Set("GetMinerSCConfig", js.FuncOf(GetMinerSCConfig))
	js.Global().Set("GetWalletStorageSCConfig", js.FuncOf(GetWalletStorageSCConfig))
	js.Global().Set("GetWalletChallengePoolInfo", js.FuncOf(GetWalletChallengePoolInfo))
	js.Global().Set("GetWalletAllocation", js.FuncOf(GetWalletAllocation))
	js.Global().Set("GetWalletAllocations", js.FuncOf(GetWalletAllocations))
	js.Global().Set("GetWalletReadPoolInfo", js.FuncOf(GetWalletReadPoolInfo))
	js.Global().Set("GetWalletStakePoolInfo", js.FuncOf(GetWalletStakePoolInfo))
	js.Global().Set("GetWalletStakePoolUserInfo", js.FuncOf(GetWalletStakePoolUserInfo))
	js.Global().Set("GetWalletBlobbers", js.FuncOf(GetWalletBlobbers))
	js.Global().Set("GetWalletBlobber", js.FuncOf(GetWalletBlobber))
	js.Global().Set("GetWalletWritePoolInfo", js.FuncOf(GetWalletWritePoolInfo))
	js.Global().Set("Encrypt", js.FuncOf(Encrypt))
	js.Global().Set("Decrypt", js.FuncOf(Decrypt))

	// sdk.go

	js.Global().Set("initStorageSDK", js.FuncOf(InitStorageSDK))
	js.Global().Set("InitAuthTicket", js.FuncOf(InitAuthTicket))
	js.Global().Set("GetNetwork", js.FuncOf(GetNetwork))
	js.Global().Set("SetMaxTxnQuery", js.FuncOf(SetMaxTxnQuery))
	js.Global().Set("SetQuerySleepTime", js.FuncOf(SetQuerySleepTime))
	js.Global().Set("SetMinSubmit", js.FuncOf(SetMinSubmit))
	js.Global().Set("SetMinConfirmation", js.FuncOf(SetMinConfirmation))
	js.Global().Set("SetNetwork", js.FuncOf(SetNetwork))
	js.Global().Set("CreateReadPool", js.FuncOf(CreateReadPool))
	js.Global().Set("AllocFilter", js.FuncOf(AllocFilter))
	js.Global().Set("GetReadPoolInfo", js.FuncOf(GetReadPoolInfo))
	js.Global().Set("ReadPoolLock", js.FuncOf(ReadPoolLock))
	js.Global().Set("ReadPoolUnlock", js.FuncOf(ReadPoolUnlock))
	js.Global().Set("GetStakePoolInfo", js.FuncOf(GetStakePoolInfo))
	js.Global().Set("GetStakePoolUserInfo", js.FuncOf(GetStakePoolUserInfo))
	js.Global().Set("StakePoolLock", js.FuncOf(StakePoolLock))
	js.Global().Set("StakePoolUnlock", js.FuncOf(StakePoolUnlock))
	js.Global().Set("StakePoolPayInterests", js.FuncOf(StakePoolPayInterests))
	js.Global().Set("GetWritePoolInfo", js.FuncOf(GetWritePoolInfo))
	js.Global().Set("WritePoolLock", js.FuncOf(WritePoolLock))
	js.Global().Set("WritePoolUnlock", js.FuncOf(WritePoolUnlock))
	js.Global().Set("GetChallengePoolInfo", js.FuncOf(GetChallengePoolInfo))
	js.Global().Set("GetStorageSCConfig", js.FuncOf(GetStorageSCConfig))
	js.Global().Set("GetBlobbers", js.FuncOf(GetBlobbers))
	js.Global().Set("GetBlobber", js.FuncOf(GetBlobber))
	js.Global().Set("GetClientEncryptedPublicKey", js.FuncOf(GetClientEncryptedPublicKey))
	js.Global().Set("GetAllocationFromAuthTicket", js.FuncOf(GetAllocationFromAuthTicket))
	js.Global().Set("GetAllocation", js.FuncOf(GetAllocation))
	js.Global().Set("GetAllocations", js.FuncOf(GetAllocations))
	js.Global().Set("GetAllocationsForClient", js.FuncOf(GetAllocationsForClient))
	js.Global().Set("CreateAllocation", js.FuncOf(CreateAllocation))
	js.Global().Set("CreateAllocationForOwner", js.FuncOf(CreateAllocationForOwner))
	js.Global().Set("UpdateAllocation", js.FuncOf(UpdateAllocation))
	js.Global().Set("FinalizeAllocation", js.FuncOf(FinalizeAllocation))
	js.Global().Set("CancelAllocation", js.FuncOf(CancelAllocation))
	js.Global().Set("UpdateBlobberSettings", js.FuncOf(UpdateBlobberSettings))
	js.Global().Set("CommitToFabric", js.FuncOf(CommitToFabric))
	js.Global().Set("GetAllocationMinLock", js.FuncOf(GetAllocationMinLock))
	js.Global().Set("SetNumBlockDownloads", js.FuncOf(SetNumBlockDownloads))

	<-make(chan bool)

	jsbridge.Close()
}
