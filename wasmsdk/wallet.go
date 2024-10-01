//go:build js && wasm
// +build js,wasm

package main

import (
	"errors"
	"os"
	"strconv"

	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/wasmsdk/jsbridge"
)

func setWallet(clientID, clientKey, peerPublicKey, publicKey, privateKey, mnemonic string, isSplit bool) error {
	fmt.Println("Set Wallet called")
	fmt.Println("ClientID : ", clientID)
	fmt.Println("ClientKey : ", clientKey)
	fmt.Println("PeerPublicKey : ", peerPublicKey)
	fmt.Println("PublicKey : ", publicKey)
	fmt.Println("PrivateKey : ", privateKey)
	fmt.Println("Mnemonic : ", mnemonic)
	fmt.Println("IsSplit : ", isSplit)

	if mnemonic == "" && !isSplit {
		return errors.New("mnemonic is required")
	}

	fmt.Println("Here 1")

	mode := os.Getenv("MODE")
	fmt.Println("gosdk setWallet, mode:", mode, "is split:", isSplit)
	keys := []zcncrypto.KeyPair{
		{
			PrivateKey: privateKey,
			PublicKey:  publicKey,
		},
	}

	w := &zcncrypto.Wallet{
		ClientID:      clientID,
		ClientKey:     clientKey,
		PeerPublicKey: peerPublicKey,
		Mnemonic:      mnemonic,
		Keys:          keys,
		IsSplit:       isSplit,
	}
	fmt.Println("set Wallet, is split:", isSplit)
	client.SetWallet(*w)
	fmt.Println("Here 2")
	fmt.Println("Wallet ID", client.ClientID())

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
