package mocks

import (
	"context"
	http "net/http"
	"time"
)

// Timeout mock any request with timeout
type Timeout struct {
	Timeout time.Duration
}

// Do provides a mock function with given fields: req
func (t *Timeout) Do(req *http.Request) (*http.Response, error) {
	time.Sleep(t.Timeout)

	time.Sleep(1 * time.Second)

	return nil, context.DeadlineExceeded
}
