//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zcncore"
)

func setWallet(clientID, clientKey, peerPublicKey, publicKey, privateKey, mnemonic string, isSplit bool) error {
	keys := []zcncrypto.KeyPair{
		{
			PrivateKey: privateKey,
			PublicKey:  publicKey,
		},
	}

	c := client.GetClient()
	c.Mnemonic = mnemonic
	c.ClientID = clientID
	c.ClientKey = clientKey
	c.PeerPublicKey = peerPublicKey
	c.Keys = keys
	c.IsSplit = isSplit

	w := &zcncrypto.Wallet{
		ClientID:      clientID,
		ClientKey:     clientKey,
		PeerPublicKey: peerPublicKey,
		Mnemonic:      mnemonic,
		Keys:          keys,
		IsSplit:       isSplit,
	}
	fmt.Println("set Wallet, is split:", isSplit)
	err := zcncore.SetWallet(*w, isSplit)
	if err != nil {
		return err
	}

	zboxApiClient.SetWallet(clientID, privateKey, publicKey)

	return nil
}
