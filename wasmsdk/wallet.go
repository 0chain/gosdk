package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zcncore"
)

//{"id":"02b1afd866f1e9c4de64ecd7c0399dd628f9d6fe7c115423eda5a64feb7a170d","version":"1.0","creation_date":1658280029,"public_key":"2c3dd374bc469e2f6d9ba754e7c9ddf53f9d217dea0f75e7308f0e476d3cb50e2fb93da6c9acb76a6241b0f0b38d8405806ed24dc7d4f8981f434cdcdeabb296"}
type createWalletResponse struct {
	ClientID    string `json:"client_id"`
	ClientKey   string `json:"client_key"`
	Version     string `json:"version"`
	DateCreated string `json:"date_created"`
}

func createWallet(clientID, clientPrivateKey, clientPublicKey, mnemonic string) (*createWalletResponse, error) {
	w := &zcncrypto.Wallet{
		ClientID:  clientID,
		ClientKey: clientPublicKey,
		Mnemonic:  mnemonic,
		Keys: []zcncrypto.KeyPair{
			{
				PrivateKey: clientPrivateKey,
				PublicKey:  clientPublicKey,
			},
		},
	}
	wg := &sync.WaitGroup{}
	cb := &createWalletCallback{wg: wg}
	wg.Add(1)
	err := zcncore.RegisterToMiners(w, cb)
	if err != nil {
		return nil, err
	}
	wg.Wait()

	if cb.success {
		cw := &createWalletResponse{}

		if err := json.Unmarshal([]byte(cb.walletString), cw); err != nil {
			return nil, fmt.Errorf("wallet: [json]%s", cb.walletString)
		}

		w.Version = cw.Version
		w.DateCreated = cw.Version

		zcncore.SetWallet(*w, false)

		return cw, nil
	}

	return nil, errors.New(cb.errMsg)

}

type createWalletCallback struct {
	walletString string
	wg           *sync.WaitGroup
	success      bool
	errMsg       string
}

func (wc *createWalletCallback) OnWalletCreateComplete(status int, wallet string, err string) {

	defer wc.wg.Done()

	if status == zcncore.StatusError {
		wc.success = false
		wc.errMsg = err
		wc.walletString = ""
		return
	}
	wc.success = true
	wc.errMsg = ""
	wc.walletString = wallet
}
