//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/0chain/gosdk/core/sys"
)

// Declare a type for the callback function
type AuthCallbackFunc func(msg string) string

// Variable to store the callback function
var authCallback AuthCallbackFunc

// Register the callback function
func registerAuthorizer(this js.Value, args []js.Value) interface{} {
	// Store the callback function
	authCallback = parseAuthorizerCallback(args[0])

	sys.Authorize = func(msg string) (string, error) {
		return authCallback(msg), nil
	}

	return nil
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
