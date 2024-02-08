//go:build !mobile
// +build !mobile

package zcncore

import (
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/zcncrypto"
)

// GetWallet get a wallet object from a wallet string
func GetWallet(walletStr string) (*zcncrypto.Wallet, error) {
	return getWallet(walletStr)
}

func GetWalletBalance(clientId string) (common.Balance, error) {
	return getWalletBalance(clientId)
}

// Deprecated: use Sign() method in zcncrypto.Wallet
func SignWith0Wallet(hash string, w *zcncrypto.Wallet) (string, error) {
	cfg, err := conf.GetClientConfig()
	if err != nil {
		return "", err
	}
	sigScheme := zcncrypto.NewSignatureScheme(cfg.SignatureScheme)
	err = sigScheme.SetPrivateKey(w.Keys[0].PrivateKey)
	if err != nil {
		return "", err
	}
	return sigScheme.Sign(hash)
}
