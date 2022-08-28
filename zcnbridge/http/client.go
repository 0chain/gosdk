package http

import (
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	RetryWaitMax = 120 * time.Second
	RetryMax     = 60

	// clientTimeout represents default http.Client timeout.
	clientTimeout = 120 * time.Second

	// tlsHandshakeTimeout represents default http.Transport TLS handshake timeout.
	tlsHandshakeTimeout = 10 * time.Second

	// dialTimeout represents default net.Dialer timeout.
	dialTimeout = 5 * time.Second
)

// NewClient creates default http.Client with timeouts.
func NewClient() *http.Client {
	d := &net.Dialer{
		Timeout: dialTimeout,
	}

	transport := &http.Transport{
		TLSHandshakeTimeout: tlsHandshakeTimeout,
		DialContext:         d.DialContext,
	}

	return &http.Client{
		Timeout:   clientTimeout,
		Transport: transport,
	}
}

func CleanClient() *http.Client {
	client := cleanhttp.DefaultPooledClient()
	client.Timeout = 10 * time.Second
	return client
}

// NewRetryableClient creates default retryablehttp.Client with timeouts and embedded NewClient result.
func NewRetryableClient() *retryablehttp.Client {
	client := retryablehttp.NewClient()
	//client.HTTPClient = NewClient()
	//client.RetryWaitMax = RetryWaitMax
	//client.RetryMax = RetryMax
	//client.Logger = nil

	return client
}
