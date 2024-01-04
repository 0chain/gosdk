//go:build !js && !wasm
// +build !js,!wasm

package zboxutil

import (
	"net"
	"net/http"
	"time"
)

var DefaultTransport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   45 * time.Second,
		KeepAlive: 45 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	MaxIdleConnsPerHost:   100,
	ReadBufferSize:        8 * 1024 * 1024,
	WriteBufferSize:       8 * 1024 * 1024,
}
