//go:build !js && !wasm
// +build !js,!wasm

package transaction

import (
	"net"
	"net/http"
	"time"
)

func createTransport(dialTimeout time.Duration) *http.Transport {
	return &http.Transport{
		Dial: (&net.Dialer{
			Timeout: dialTimeout,
		}).Dial,
		TLSHandshakeTimeout: dialTimeout,
	}

}
