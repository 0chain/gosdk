//go:build mobile
// +build mobile

package zcncore

import (
	"strconv"

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

func GetWalletBalanceMobile(id string) (string, error) {
	balance, err := GetWalletBalance(id)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(int64(balance), 10), nil
}

func RegisterToMiners(clientId, pubKey string, callback WalletCallback) error {
	wallet := zcncrypto.Wallet{ClientID: clientId, ClientKey: pubKey}
	return registerToMiners(&wallet, callback)
}
