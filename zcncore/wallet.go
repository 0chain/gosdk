//go:build !mobile
// +build !mobile

package zcncore

import (
	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/zcncrypto"
)

// GetWallet get a wallet object from a wallet string
func GetWallet(walletStr string) (*zcncrypto.Wallet, error) {
	return getWallet(walletStr)
}

// GetWalletBalance retrieve wallet balance from sharders
//   - id: client id
func GetWalletBalance(clientId string) (common.Balance, int64, error) {
	return getWalletBalance(clientId)
}

func SignWith0Wallet(hash string, w *zcncrypto.Wallet) (string, error) {
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	err := sigScheme.SetPrivateKey(w.Keys[0].PrivateKey)
	if err != nil {
		return "", err
	}
	return sigScheme.Sign(hash)
}
