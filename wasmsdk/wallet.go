//go:build js && wasm
// +build js,wasm

package main

import (
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zcncore"
)

func setWallet(clientID, publicKey, privateKey, mnemonic string) error {
	keys := []zcncrypto.KeyPair{
		{
			PrivateKey: privateKey,
			PublicKey:  publicKey,
		},
	}

	c := client.GetClient()
	c.Mnemonic = mnemonic
	c.ClientID = clientID
	c.ClientKey = publicKey
	c.Keys = keys

	w := &zcncrypto.Wallet{
		ClientID:  clientID,
		ClientKey: publicKey,
		Mnemonic:  mnemonic,
		Keys:      keys,
	}
	err := zcncore.SetWallet(*w, false)
	if err != nil {
		return err
	}

	zboxApiClient.SetWallet(clientID, privateKey, publicKey)

	return nil
}
