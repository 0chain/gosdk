//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zcncore"
)

type AuthCallbackFunc func(msg string) string

var authMsgCallback AuthCallbackFunc
var authCallback AuthCallbackFunc
var authResponseC chan string
var authMsgResponseC chan string
var authMsgLock = make(chan struct{}, 1)

// Register the callback function
func registerAuthorizer(this js.Value, args []js.Value) interface{} {
	// Store the callback function
	authCallback = parseAuthorizerCallback(args[0])
	authResponseC = make(chan string, 1)

	sys.Authorize = func(msg string) (string, error) {
		authCallback(msg)
		return <-authResponseC, nil
	}
	return nil
}

func registerZauthServer(serverAddr string) {
	fmt.Println("registerZauthServer...")
	jsbridge.SetZauthServer(serverAddr)
	sys.SetAuthorize(zcncore.ZauthSignTxn(serverAddr))
	sys.SetAuthCommon(zcncore.ZauthAuthCommon(serverAddr))
}

// zvaultNewWallet generates new split wallet
func zvaultNewWallet(serverAddr, token string) (string, error) {
	return zcncore.CallZvaultNewWalletString(serverAddr, token, "")
}

// zvaultNewSplit generates new split wallet from existing clientID
func zvaultNewSplit(clientID, serverAddr, token string) (string, error) {
	return zcncore.CallZvaultNewWalletString(serverAddr, token, clientID)
}

func zvaultStoreKey(serverAddr, token, privateKey string) (string, error) {
	return zcncore.CallZvaultStoreKeyString(serverAddr, token, privateKey)
}

func zvaultRetrieveKeys(serverAddr, token, clientID string) (string, error) {
	return zcncore.CallZvaultRetrieveKeys(serverAddr, token, clientID)
}

func zvaultRevokeKey(serverAddr, token, clientID, publicKey string) error {
	return zcncore.CallZvaultRevokeKey(serverAddr, token, clientID, publicKey)
}

func zvaultDeletePrimaryKey(serverAddr, token, clientID string) error {
	return zcncore.CallZvaultDeletePrimaryKey(serverAddr, token, clientID)
}

func zvaultRetrieveWallets(serverAddr, token string) (string, error) {
	return zcncore.CallZvaultRetrieveWallets(serverAddr, token)
}

func zvaultRetrieveSharedWallets(serverAddr, token string) (string, error) {
	return zcncore.CallZvaultRetrieveSharedWallets(serverAddr, token)
}

func registerAuthCommon(this js.Value, args []js.Value) interface{} {
	authMsgCallback = parseAuthorizerCallback(args[0])
	authMsgResponseC = make(chan string, 1)

	sys.AuthCommon = func(msg string) (string, error) {
		authMsgLock <- struct{}{}
		defer func() {
			<-authMsgLock
		}()
		authMsgCallback(msg)
		return <-authMsgResponseC, nil
	}
	return nil
}

func authResponse(response string) {
	authResponseC <- response
}

// Use the stored callback function
func callAuth(this js.Value, args []js.Value) interface{} {
	fmt.Println("callAuth is called")
	if len(args) == 0 {
		return nil
	}

	if authCallback != nil {
		msg := args[0].String()
		result, _ := sys.Authorize(msg)
		fmt.Println("auth is called, result:", result)
		return js.ValueOf(result)
	}

	return nil
}

// Parse the JavaScript callback function into Go AuthorizerCallback type
func parseAuthorizerCallback(jsCallback js.Value) AuthCallbackFunc {
	return func(msg string) string {
		// Call the JavaScript callback function from Go
		result := jsCallback.Invoke(msg)
		return result.String()
	}
}
