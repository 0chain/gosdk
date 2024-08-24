//go:build js && wasm
// +build js,wasm

package zboxutil

import (
	coreHttp "github.com/0chain/gosdk/core/http"
	"net/http"
	"time"
)

var DefaultTransport = &http.Transport{
	Proxy: coreHttp.EnvProxy.Proxy,

	MaxIdleConns:          100,
	IdleConnTimeout:       60 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	MaxIdleConnsPerHost:   100,
}
