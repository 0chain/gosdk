package resty

import (
	"time"
)

// WithRetry set retry times if request is failure with 5xx status code. retry is ingore if it is less than 1.
func WithRetry(retry int) Option {
	return func(r *Resty) {
		if retry > 0 {
			r.retry = retry
		}
	}
}

// WithHeader set header for http request
func WithHeader(header map[string]string) Option {
	return func(r *Resty) {
		if len(header) > 0 {
			r.header = header
		}
	}
}

// WithTimeout set timeout of http request.
func WithTimeout(timeout time.Duration) Option {
	return func(r *Resty) {
		if timeout > 0 {
			r.timeout = timeout
		}
	}
}
