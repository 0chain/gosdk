//go:build js && wasm
// +build js,wasm

package main

import (
	"time"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zcncore"
)

func setWallet(clientID, publicKey, privateKey string) error {
	c := client.GetClient()
	c.ClientID = clientID
	c.ClientKey = publicKey

	w := &zcncrypto.Wallet{
		ClientID:  clientID,
		ClientKey: publicKey,
		Keys: []zcncrypto.KeyPair{
			{
				PrivateKey: privateKey,
				PublicKey:  publicKey,
			},
		},
	}
	err := zcncore.SetWallet(*w, false)
	if err != nil {
		return err
	}

	zboxApiClient.SetWallet(clientID, privateKey, publicKey)

	forceRefreshWalletNonce <- true

	return nil
}

var forceRefreshWalletNonce = make(chan bool, 1)

func startRefreshWalletNonce() {
	for {
		select {
		case <-forceRefreshWalletNonce:
		case <-time.After(1 * time.Minute):
		}
		c := client.GetClient()
		clientID := c.ClientID
		if clientID != "" {
			nonce, err := zcncore.GetWalletNonce(c.ClientID)
			if err != nil {
				zcnLogger.Error("wallet: get wallet nonce ", err)
			} else {
				c.Nonce = nonce
				zcnLogger.Info("wallet: latest nonce ", nonce)
			}
		}
	}
}
