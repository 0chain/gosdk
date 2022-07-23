//go:build js && wasm
// +build js,wasm

package transaction

import (
	"net/http"
	"time"
)

func createTransport(dialTimeout time.Duration) *http.Transport {
	return &http.Transport{
		TLSHandshakeTimeout: dialTimeout,
	}
}
