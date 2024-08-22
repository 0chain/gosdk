//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/0chain/gosdk/core/sys"
)

type AuthCallbackFunc func(msg string) string

var authCallback AuthCallbackFunc
var authResponseC chan string

// registerAuthorizer Register the callback function to authorize the transaction.
// This function is called from JavaScript.
// It stores the callback function in the global variable authCallback.
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

// authResponse Publishes the response to the authorization request.
// 		`response` is the response to the authorization request.
func authResponse(response string) {
	authResponseC <- response
}

// callAuth Call the authorization callback function and provide the message to pass to it.
// The message is passed as the first argument to the js calling.
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
