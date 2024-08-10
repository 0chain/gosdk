//go:build mobile
// +build mobile

package zcncore

import (
	"github.com/0chain/gosdk/core/zcncrypto"
)

// Wallet interface to gather all wallet related functions
type Wallet interface {
	// Sign sign the hash
	Sign(hash string) (string, error)
}

type wallet struct {
	zcncrypto.Wallet
}

// Sign sign the given string using the wallet's private key
func (w *wallet) Sign(hash string) (string, error) {
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	err := sigScheme.SetPrivateKey(w.Keys[0].PrivateKey)
	if err != nil {
		return "", err
	}
	return sigScheme.Sign(hash)
}

// GetWalletBalance retrieve wallet balance from sharders
//   - id: client id
func GetWalletBalance(id string) (int64, error) {
	balance, _, err := getWalletBalance(id)
	if err != nil {
		return 0, err
	}
	return int64(balance), nil
}
