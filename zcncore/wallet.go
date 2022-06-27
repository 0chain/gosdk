//go:build !mobile
// +build !mobile

package zcncore

import "github.com/0chain/gosdk/core/zcncrypto"

func RegisterToMiners(wallet *zcncrypto.Wallet, statusCb WalletCallback) error {
	return registerToMiners(wallet, statusCb)
}
