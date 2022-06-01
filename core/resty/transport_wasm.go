//go:build js && wasm
// +build js,wasm

package resty

import (
	"net/http"
	"time"
)

// Run the HTTP request in a goroutine and pass the response to f.
var DefaultTransport = &http.Transport{
	Proxy:                 http.ProxyFromEnvironment,
	MaxIdleConns:          1000,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	MaxIdleConnsPerHost:   5,
	ForceAttemptHTTP2:     true,
}
