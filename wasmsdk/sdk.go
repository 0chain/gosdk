//go:build js && wasm
// +build js,wasm

package main

import (
	"io"
	"os"

	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/wasmsdk/zbox"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/encryption"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"
)

var CreateObjectURL func(buf []byte, mimeType string) string

// initSDKs init sharder/miners ,
func initSDKs(chainID, blockWorker, signatureScheme string,
	minConfirmation, minSubmit, confirmationChainLength int) error {

	err := sdk.InitStorageSDK("{}", blockWorker, chainID, signatureScheme, nil, 0)
	if err != nil {
		return err
	}

	err = zcncore.InitZCNSDK(blockWorker, signatureScheme,
		zcncore.WithChainID(chainID),
		zcncore.WithMinConfirmation(minConfirmation),
		zcncore.WithMinSubmit(minSubmit),
		zcncore.WithConfirmationChainLength(confirmationChainLength))

	if err != nil {
		return err
	}

	return nil
}

func SetWallet(clientID, publicKey string) {
	c := client.GetClient()
	c.ClientID = clientID
	c.ClientKey = publicKey

	if len(zboxHost) > 0 {
		zboxClient = zbox.NewClient(zboxHost, c.ClientID, c.ClientKey)
	}
}

func GetEncryptedPublicKey(mnemonic string) (string, error) {
	encScheme := encryption.NewEncryptionScheme()
	_, err := encScheme.Initialize(mnemonic)
	if err != nil {
		return "", err
	}
	return encScheme.GetPublicKey()
}

var sdkLogger *logger.Logger
var zcnLogger *logger.Logger
var logEnabled = false

func showLogs() {
	zcnLogger.SetLevel(logger.DEBUG)
	sdkLogger.SetLevel(logger.DEBUG)

	zcncore.GetLogger().SetLogFile(os.Stdout, true)
	sdkLogger.SetLogFile(os.Stdout, true)

	logEnabled = true
}

func hideLogs() {
	zcnLogger.SetLevel(logger.ERROR)
	sdkLogger.SetLevel(logger.ERROR)

	zcnLogger.SetLogFile(io.Discard, false)
	sdkLogger.SetLogFile(io.Discard, false)

	logEnabled = false
}
