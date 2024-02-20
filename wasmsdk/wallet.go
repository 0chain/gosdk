//go:build js && wasm
// +build js,wasm

package main

import (
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/core/client"
)

func setWallet(clientID, publicKey, privateKey, mnemonic string) error {
	keys := []zcncrypto.KeyPair{
		{
			PrivateKey: privateKey,
			PublicKey:  publicKey,
		},
	}

	w := &zcncrypto.Wallet{
		ClientID:  clientID,
		ClientKey: publicKey,
		Mnemonic:  mnemonic,
		Keys:      keys,
	}
	err := client.SetWallet(*w)
	if err != nil {
		return err
	}

	zboxApiClient.SetWallet(clientID, privateKey, publicKey)

	return nil
}
