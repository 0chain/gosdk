//go:build !mobile
// +build !mobile

package zcncore

import (
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/zcncrypto"
)

func RegisterToMiners(wallet *zcncrypto.Wallet, statusCb WalletCallback) error {
	return registerToMiners(wallet, statusCb)
}

//GetWallet get a wallet object from a wallet string
func GetWallet(walletStr string) (*zcncrypto.Wallet, error) {
	return getWallet(walletStr)
}

func SignWith0Wallet(hash string, w *zcncrypto.Wallet) (string, error) {
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	err := sigScheme.SetPrivateKey(w.Keys[0].PrivateKey)
	if err != nil {
		return "", err
	}
	return sigScheme.Sign(hash)
}

var Logger = &logging

func GetLogger() *logger.Logger {
	return &logging
}

// CloseLog closes log file
func CloseLog() {
	logging.Close()
}
