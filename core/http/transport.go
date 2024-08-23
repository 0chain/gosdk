//go:build !js && !wasm
// +build !js,!wasm

package http

import (
	"net"
	"net/http"
	"time"
)

var DefaultTransport = &http.Transport{
	Proxy: EnvProxy.Proxy,
	DialContext: (&net.Dialer{
		Timeout:   3 * time.Minute,
		KeepAlive: 45 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   45 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	MaxIdleConnsPerHost:   25,
}
