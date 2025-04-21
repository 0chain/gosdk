//go:build !js && !wasm
// +build !js,!wasm

package zboxutil

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

var DefaultTransport = &http.Transport{
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify:     false,
		MinVersion:             tls.VersionTLS12,
		SessionTicketsDisabled: false, // Enable TLS session reuse
	},
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 45 * time.Second,
	}).DialContext,
	MaxIdleConns:        500,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     45 * time.Second,
	DisableKeepAlives:   false,
	ForceAttemptHTTP2:   true,
}
