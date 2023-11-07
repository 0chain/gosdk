package main

import (
	"errors"
	"syscall/js"

	"github.com/0chain/gosdk/core/sys"
)

var auth sys.AuthorizeFunc

func bridgeAuth(this js.Value, p []js.Value) interface{} {
	if len(p) != 1 || !p[0].Truthy() {
		return nil
	}

	// passed js function from webapp
	jsFunc := p[0]

	goFunc := func(msg string) (string, error) {
		result := jsFunc.Invoke(msg)
		// js function returns a string and an error message
		errorStr := result.Index(1).String()
		if errorStr != "" {
			return result.Index(0).String(), errors.New(errorStr)
		}
		return result.Index(0).String(), nil
	}

	auth = goFunc
	RegisterAuthorizer()
	return nil
}

func createJsToGoBridge() {
	js.Global().Set("registerAuthorizeFunc", js.FuncOf(bridgeAuth))
}

func RegisterAuthorizer() {
	sys.Authorize = auth
}
