package http

import (
	"net/http"
	"time"

	"github.com/0chain/gosdk/zboxcore/zboxutil"
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
	return &http.Client{
		Transport: zboxutil.DefaultTransport,
	}
}

func CleanClient() *http.Client {
	client := &http.Client{
		Transport: zboxutil.DefaultTransport,
	}
	client.Timeout = 250 * time.Second
	return client
}

// NewRetryableClient creates default retryablehttp.Client with timeouts and embedded NewClient result.
func NewRetryableClient() *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.HTTPClient = &http.Client{
		Transport: zboxutil.DefaultTransport,
	}
	//client.RetryWaitMax = RetryWaitMax
	//client.RetryMax = RetryMax
	//client.Logger = nil

	return client
}
