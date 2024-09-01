//go:build js && wasm
// +build js,wasm

package main

import (
	"errors"

	"fmt"
	"os"
	"strconv"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zcncore"
)

func setWallet(clientID, clientKey, peerPublicKey, publicKey, privateKey, mnemonic string, isSplit bool) error {
	if mnemonic == "" && !isSplit {
		return errors.New("mnemonic is required")
	}
	mode := os.Getenv("MODE")
	fmt.Println("gosdk setWallet, mode:", mode, "is split:", isSplit)
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
	if mode == "" { // main thread, need to notify the web worker to update wallet
		// notify the web worker to update wallet
		if err := jsbridge.PostMessageToAllWorkers(jsbridge.MsgTypeUpdateWallet, map[string]string{
			"client_id":       clientID,
			"client_key":      clientKey,
			"peer_public_key": peerPublicKey,
			"public_key":      publicKey,
			"private_key":     privateKey,
			"mnemonic":        mnemonic,
			"is_split":        strconv.FormatBool(isSplit),
		}); err != nil {
			return err
		}
	}

	return nil
}
