//go:build mobile
// +build mobile

package zcncore

import (
	"github.com/0chain/gosdk/core/zcncrypto"
)

type Wallet interface {
	Sign(hash string) (string, error)
}

type wallet struct {
	zcncrypto.Wallet
}

func (w *wallet) Sign(hash string) (string, error) {
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	err := sigScheme.SetPrivateKey(w.Keys[0].PrivateKey)
	if err != nil {
		return "", err
	}
	return sigScheme.Sign(hash)
}

func GetWalletBalance(id string) (int64, error) {
	balance, err := getWalletBalance(id)
	if err != nil {
		return 0, err
	}
	return int64(balance), nil
}
