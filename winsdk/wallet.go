package main

/*
#include <stdlib.h>
*/
import (
	"C"
)

import (
	"github.com/0chain/gosdk/core/client"
	"os"
	"path/filepath"

	"github.com/0chain/gosdk/zcncore"
)

// CreateWallet - create a new wallet, and save it to ~/.zcn/wallet.json
// ## Outputs
//
//	{
//	"error":"",
//	"result":\"{}\"",
//	}
//
//export CreateWallet
func CreateWallet() *C.char {
	w, err := zcncore.CreateWalletOffline()
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	d, err := getZcnWorkDir()
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	if err = os.WriteFile(filepath.Join(d, "wallet.json"), []byte(w), 0644); err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	return WithJSON(w, nil)
}

// RecoverWallet - recover the wallet, and save it to ~/.zcn/wallet.json
// ## Outputs
//
//	{
//	"error":"",
//	"result":\"{}\"",
//	}
//
//export RecoverWallet
func RecoverWallet(mnemonic *C.char) *C.char {
	w, err := zcncore.RecoverOfflineWallet(C.GoString(mnemonic))
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	d, err := getZcnWorkDir()
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	if err = os.WriteFile(filepath.Join(d, "wallet.json"), []byte(w), 0644); err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	return WithJSON(w, nil)
}

// GetWalletBalance - get wallet balance
// ## Inputs
// - clientID
// ## Outputs
//
//	{
//	"error":"",
//	"result":\"0.00\"",
//	}
//
//export GetWalletBalance
func GetWalletBalance(clientID *C.char) *C.char {
	b, err := client.GetBalance(C.GoString(clientID))
	if err != nil {
		log.Error("win: ", err)
		return WithJSON(0, err)
	}

	return WithJSON(b.ToToken())
}
