//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"
)

var CreateObjectURL func(buf []byte, mimeType string) string

// initSDKs init sharder/miners ,
func initSDKs(chainID, blockWorker, signatureScheme string,
	minConfirmation, minSubmit, confirmationChainLength int, zboxHost, zboxAppType string) error {

	zboxApiClient.SetRequest(zboxHost, zboxAppType)

	err := sdk.InitStorageSDK("{}", blockWorker, chainID, signatureScheme, nil, 0)
	if err != nil {
		fmt.Println("wasm: InitStorageSDK ", err)
		return err
	}

	err = zcncore.InitZCNSDK(blockWorker, signatureScheme,
		zcncore.WithChainID(chainID),
		zcncore.WithMinConfirmation(minConfirmation),
		zcncore.WithMinSubmit(minSubmit),
		zcncore.WithConfirmationChainLength(confirmationChainLength))

	if err != nil {
		fmt.Println("wasm: InitZCNSDK ", err)
		return err
	}

	return nil
}

func SetWallet(clientID, publicKey, privateKey string) error {
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

	return nil
}

var sdkLogger *logger.Logger
var zcnLogger *logger.Logger
var logEnabled = false

func showLogs() {
	zcnLogger.SetLevel(logger.DEBUG)
	sdkLogger.SetLevel(logger.DEBUG)

	zcnLogger.SetLogFile(os.Stdout, true)
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

func isWalletID(clientID string) bool {
	if clientID == "" {
		return false
	}

	if !isHash(clientID) {
		return false
	}

	return true

}

const HASH_LENGTH = 32

func isHash(str string) bool {
	bytes, err := hex.DecodeString(str)
	return err == nil && len(bytes) == HASH_LENGTH
}

// getLookupHash get lookup hash with allocation id and path
func getLookupHash(allocationID string, path string) string {
	return encryption.Hash(allocationID + ":" + path)
}
