//go:build js && wasm
// +build js,wasm

package resty

import (
	"net/http"
	"time"
)

var DefaultHeader map[string]string

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

func init() {
	DefaultHeader = make(map[string]string)
	DefaultHeader["js.fetch:mode"] = "cors"
}

// SetSessionID stores the browser session ID so it is attached as
// X-App-Session-ID on every outgoing blobber HTTP request. Blobbers
// forward this value to 0box, which uses it to suppress echoing the
// event back to the session that caused it.
func SetSessionID(id string) {
	if DefaultHeader == nil {
		DefaultHeader = make(map[string]string)
	}
	if id != "" {
		DefaultHeader["X-App-Session-ID"] = id
	}
}
