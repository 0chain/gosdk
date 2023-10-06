package main

/*
#include <stdlib.h>
*/
import (
	"C"
)

import (
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
		return WithJSON("", err)
	}

	d, err := os.UserHomeDir()
	if err != nil {
		return WithJSON("", err)
	}

	z := filepath.Join(d, ".zcn")

	// create ~/.zcn folder if it doesn't exists
	os.MkdirAll(z, 0766) //nolint: errcheck

	if err = os.WriteFile(filepath.Join(z, "wallet.json"), []byte(w), 0644); err != nil {
		return WithJSON("", err)
	}

	return WithJSON(w, nil)
}
