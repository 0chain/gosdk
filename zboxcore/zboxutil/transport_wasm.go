//go:build js && wasm
// +build js,wasm

package zboxutil

import (
	"net/http"
	"time"
)

var DefaultTransport = &http.Transport{
	Proxy: envProxy.Proxy,

	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	MaxIdleConnsPerHost:   100,
}
