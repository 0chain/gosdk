//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"syscall/js"
)

func validateClientDetails(allocation, clientJSON string) error {
	if len(allocation) == 0 || len(clientJSON) == 0 {
		return NewError("invalid_param", "Please provide allocation and client_json for the client")
	}
	return nil
}

// This function try to execute wasm functions that are wrapped with "Promise"
// see: https://stackoverflow.com/questions/68426700/how-to-wait-a-js-async-function-from-golang-wasm/68427221#comment120939975_68427221
func await(awaitable js.Value) ([]js.Value, []js.Value) {
	then := make(chan []js.Value)
	defer close(then)
	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		then <- args
		return nil
	})
	defer thenFunc.Release()

	catch := make(chan []js.Value)
	defer close(catch)
	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		catch <- args
		return nil
	})
	defer catchFunc.Release()

	awaitable.Call("then", thenFunc).Call("catch", catchFunc)

	select {
	case result := <-then:
		return result, []js.Value{js.Null()}
	case err := <-catch:
		return []js.Value{js.Null()}, err
	}
}

//-----------------------------------------------------------------------------
// Ported over from `code/go/0proxy.io/core/common/errors.go`
//-----------------------------------------------------------------------------

/*Error type for a new application error */
type Error struct {
	Code string `json:"code,omitempty"`
	Msg  string `json:"msg"`
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s: %s", err.Code, err.Msg)
}

/*InvalidRequest - create error messages that are needed when validating request input */
func InvalidRequest(msg string) error {
	return NewError("invalid_request", fmt.Sprintf("Invalid request (%v)", msg))
}
