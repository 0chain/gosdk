//go:build js && wasm
// +build js,wasm

package resty

import (
	"net/http"
	"time"
)

// CreateClient a function that create a client instance
var CreateClient = func(t *http.Transport, timeout time.Duration) Client {
	c := &WasmClient{
		Client: &http.Client{
			Transport: t,
		},
	}

	if timeout > 0 {
		c.Client.Timeout = timeout
	}

	return c
}

type WasmClient struct {
	*http.Client
}

func (c *WasmClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("js.fetch:mode", "cors")

	return c.Client.Do(req)
}
