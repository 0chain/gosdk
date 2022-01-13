//go:build js && wasm
// +build js,wasm

package main

import (
	"os"

	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"
)

var CreateObjectURL func(buf []byte, mimeType string) string
var AppendVideo func(buf []byte)

// Init init sharder/miners ,
func Init(chainID, blockWorker, signatureScheme string,
	minConfirmation, minSubmit, confirmationChainLength int) error {

	err := sdk.InitStorageSDK("{}", blockWorker, chainID, signatureScheme, nil)
	if err != nil {
		return err
	}
	zcncore.InitZCNSDK(blockWorker, signatureScheme,
		zcncore.WithChainID(chainID),
		zcncore.WithMinConfirmation(minConfirmation),
		zcncore.WithMinSubmit(minSubmit),
		zcncore.WithConfirmationChainLength(confirmationChainLength))

	return nil
}

func SetWallet(clientID, publicKey string) {
	c := client.GetClient()
	c.ClientID = clientID
	c.ClientKey = publicKey
}

func GetEncryptedPublicKey(mnemonic string) (string, error) {
	encScheme := encryption.NewEncryptionScheme()
	_, err := encScheme.Initialize(mnemonic)
	if err != nil {
		return "", err
	}
	return encScheme.GetPublicKey()
}

func showLogs() {
	zcncore.GetLogger().SetLevel(logger.DEBUG)
	sdk.GetLogger().SetLevel(logger.DEBUG)

	zcncore.GetLogger().SetLogFile(os.Stdout, true)
	sdk.GetLogger().SetLogFile(os.Stdout, true)
}

func hideLogs() {
	zcncore.GetLogger().SetLevel(logger.ERROR)
	sdk.GetLogger().SetLevel(logger.ERROR)

	zcncore.GetLogger().SetLogFile(os.Stdout, false)
	sdk.GetLogger().SetLogFile(os.Stdout, false)
}
