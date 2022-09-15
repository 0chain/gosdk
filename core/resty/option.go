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

// WithRequestInterceptor intercept request
func WithRequestInterceptor(interceptor func(req *http.Request) error) Option {
	return func(r *Resty) {
		r.requestInterceptor = interceptor
	}
}

// WithTransport set transport
func WithTransport(transport *http.Transport) Option {
	return func(r *Resty) {
		r.transport = transport
	}
}
