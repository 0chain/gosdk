//go:build js && wasm
// +build js,wasm

package main

import (
	"errors"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zcncore"
)

// setWallet sets the wallet used by the client for the network transactions and the backend API requests
//   - clientID is the client id
//   - publicKey is the public key of the client
//   - privateKey is the private key of the client
//   - mnemonic is the mnemonic of the client
func setWallet(clientID, publicKey, privateKey, mnemonic string) error {
	if mnemonic == "" {
		return errors.New("mnemonic is required")
	}
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
