package resty

import (
	"net/http"
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
		if r.header == nil {
			r.header = make(map[string]string)
		}

		for k, v := range header {
			r.header[k] = v
		}
		r.header = header

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

// WithBefore do something before request is sent.
func WithBefore(handle func(req *http.Request)) Option {
	return func(r *Resty) {
		if handle != nil {
			r.beforeSend = handle
		}
	}
}
