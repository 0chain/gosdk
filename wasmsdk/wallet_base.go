package main

import (
	"github.com/0chain/gosdk/zcncore"
)

// Split keys from the primary master key
//lint:ignore
func splitKeys(privateKey string, numSplits int) (string, error) {
	wStr, err := zcncore.SplitKeys(privateKey, numSplits)
	return wStr, err
}

// SetWalletInfo should be set before any transaction or client specific APIs
// splitKeyWallet parameter is valid only if SignatureScheme is "BLS0Chain"
// # Inputs
// - jsonWallet: json format of wallet
// - splitKeyWallet: if wallet keys is split
//lint:ignore
func setWalletInfo(jsonWallet string, splitKeyWallet bool) error {
	err := zcncore.SetWalletInfo(jsonWallet, splitKeyWallet)
	return err
}

// SetAuthUrl will be called by app to set zauth URL to SDK.
// # Inputs
// - url: the url of zAuth server
//lint:ignore
func setAuthUrl(url string) error {
	err := zcncore.SetAuthUrl(url)
	return err
}
