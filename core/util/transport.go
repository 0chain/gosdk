//go:build !js && !wasm
// +build !js,!wasm

package util

import (
	"net"
	"net/http"
	"time"
)

// Run the HTTP request in a goroutine and pass the response to f.
var transport = &http.Transport{
	Proxy:                 http.ProxyFromEnvironment,
	MaxIdleConns:          1000,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	// Per-blobber idle-connection pool. 5 was far too low for a gateway serving
	// many concurrent reads: the eblobbers are HTTP/1.1, so beyond 5 in-flight
	// requests to the SAME blobber the client could not reuse a pooled connection
	// and paid a fresh TCP dial per read — the read-latency tail (profiled: cold
	// reads block in NewHTTPGetRequest, not CPU). Pool generously so concurrent
	// block downloads reuse warm connections to each blobber.
	MaxIdleConnsPerHost:   256,
	ForceAttemptHTTP2:     true,

	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
}
