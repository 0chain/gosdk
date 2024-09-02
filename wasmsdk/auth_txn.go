//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"os"
	"syscall/js"

	"github.com/0chain/gosdk/core/sys"
)

type AuthCallbackFunc func(msg string) string

var authResponseC chan string
var authMsgResponseC chan string
var authMsgLock = make(chan struct{}, 1)

// Register the callback function
func registerAuthorizer(this js.Value, args []js.Value) interface{} {
	// Store the callback function
	authCallback := parseAuthorizerCallback(args[0])
	authResponseC = make(chan string, 1)

	sys.Authorize = func(msg string) (string, error) {
		authCallback(msg)
		return <-authResponseC, nil
	}
	return nil
}

func registerAuthCommon(this js.Value, args []js.Value) interface{} {
	authMsgCallback := parseAuthorizerCallback(args[0])
	authMsgResponseC = make(chan string, 1)
	mode := os.Getenv("MODE")

	sys.AuthCommon = func(msg string) (string, error) {
		authMsgLock <- struct{}{}
		defer func() {
			<-authMsgLock
		}()
		authMsgCallback(msg)
		return <-authMsgResponseC, nil
	}
	fmt.Println("register auth common, mode: ", mode, "common:", sys.AuthCommon)
	return nil
}

func authResponse(response string) {
	authResponseC <- response
}

func authMsgResponse(response string) {
	authMsgResponseC <- response
}

// Use the stored callback function
// func callAuth(this js.Value, args []js.Value) interface{} {
// 	fmt.Println("callAuth is called")
// 	if len(args) == 0 {
// 		return nil
// 	}

// 	// if authCallback != nil {
// 	msg := args[0].String()
// 	result, _ := sys.Authorize(msg)
// 	return js.ValueOf(result)
// 	// }

// 	// return nil
// }

// Parse the JavaScript callback function into Go AuthorizerCallback type
func parseAuthorizerCallback(jsCallback js.Value) AuthCallbackFunc {
	return func(msg string) string {
		// Call the JavaScript callback function from Go
		result := jsCallback.Invoke(msg)
		return result.String()
	}
}
