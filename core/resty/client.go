//go:build !js && !wasm
// +build !js,!wasm

package resty

import (
	"net/http"
	"time"
)

// CreateClient a function that create a client instance
var CreateClient = func(t *http.Transport, timeout time.Duration) Client {
	client := &http.Client{
		Transport: t,
	}
	if timeout > 0 {
		client.Timeout = timeout
	}

	return client
}
