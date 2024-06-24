package main

import (
	"github.com/0chain/gosdk/zcncore"
)

// Split keys from the primary master key

// nolint: unused
func splitKeys(privateKey string, numSplits int) (string, error) {
	wStr, err := zcncore.SplitKeys(privateKey, numSplits)
	return wStr, err
}

// SetWalletInfo should be set before any transaction or client specific APIs
// splitKeyWallet parameter is valid only if SignatureScheme is "BLS0Chain"
// # Inputs
// - jsonWallet: json format of wallet
// - splitKeyWallet: if wallet keys is split

// nolint: unused
func setWalletInfo(jsonWallet string, splitKeyWallet bool) bool {
	err := zcncore.SetWalletInfoJSON(jsonWallet, splitKeyWallet)
	if err == nil {
		return true
	} else {
		return false
	}
}

// SetAuthUrl will be called by app to set zauth URL to SDK.
// # Inputs
// - url: the url of zAuth server

// nolint: unused
func setAuthUrl(url string) bool {
	err := zcncore.SetAuthUrl(url)
	if err == nil {
		return true
	} else {
		return false
	}
}
