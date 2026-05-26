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
	MaxIdleConns:        5000,
	MaxIdleConnsPerHost: 2048, // Increased for high-concurrency workloads (e.g., warp tests)
	IdleConnTimeout:     45 * time.Second,
	DisableKeepAlives:   false,
	// HTTP/1.1 connection pool per blobber instead of a single multiplexed
	// HTTP/2 connection. Under load, h2 funneled ALL metadata requests
	// (GetRefs / file-meta) through one conn per blobber (~2 conns total,
	// 80-160 multiplexed streams), serializing effective concurrency to ~3.5
	// servers regardless of client concurrency — the cause of flat-across-
	// concurrency GET throughput. With a non-nil TLSClientConfig, setting this
	// false disables h2 negotiation, so each concurrent request gets its own
	// pooled h1 connection (bounded by MaxIdleConnsPerHost above).
	ForceAttemptHTTP2: false,
}
