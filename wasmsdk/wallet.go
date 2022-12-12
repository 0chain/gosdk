package main

import (
	"github.com/0chain/gosdk/zcncore"
)

func createWallet() (string, error) {
	return zcncore.CreateWalletOffline()

}

func recoverWallet(mnemonics string) (string, error) {
	return zcncore.RecoverOfflineWallet(mnemonics)
}
