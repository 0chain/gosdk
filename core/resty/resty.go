// Package resty HTTP and REST client library with parallel feature
package resty

import (
	"context"
	"net/http"
	"time"
)

// New create a Resty instance.
func New(transport *http.Transport, handle Handle, opts ...Option) *Resty {
	r := &Resty{
		transport: transport,
		handle:    handle,
	}

	for _, option := range opts {
		option(r)
	}

	if r.transport == nil {
		r.transport = &http.Transport{}
	}

	r.client = CreateClient(r.transport, r.timeout)

	return r
}

// CreateClient a function that create a client instance
var CreateClient = func(t *http.Transport, timeout time.Duration) Client {
	client := &http.Client{
		Transport: t,
	}
	if timeout > 0 {
		client.Timeout = timeout
	}

	return client
}

// Client http client
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// Handle handler of http response
type Handle func(req *http.Request, resp *http.Response, cf context.CancelFunc, err error) error

// Option set restry option
type Option func(*Resty)

// Resty HTTP and REST client library with parallel feature
type Resty struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	qty        int
	done       chan Result

	transport *http.Transport
	client    Client
	handle    Handle

	timeout time.Duration
	retry   int
	header  map[string]string
}

// DoGet execute http requests with GET method in parallel
func (r *Resty) DoGet(ctx context.Context, urls ...string) {
	r.ctx, r.cancelFunc = context.WithCancel(ctx)

	r.qty = len(urls)
	r.done = make(chan Result, r.qty)

	for _, url := range urls {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		for key, value := range r.header {
			req.Header.Set(key, value)
		}

		req.Close = true
		req.Header.Set("Connection", "close")

		if err != nil {

			r.done <- Result{Request: req, Response: nil, Err: err}

			continue
		}

		go r.httpDo(req.WithContext(r.ctx))
	}

}

func (r *Resty) httpDo(req *http.Request) {

	var resp *http.Response
	var err error

	if r.retry > 0 {
		for i := 1; ; i++ {
			resp, err = r.client.Do(req)
			if resp != nil && resp.StatusCode == 200 {
				break
			}
			// close body ReadClose to release resource before retrying it
			if resp != nil && resp.Body != nil {
				// don't close it if it is latest retry
				if i < r.retry {
					resp.Body.Close()
				}

			}

			if i == r.retry {
				break
			}

			if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
				time.Sleep(1 * time.Second)
			}

			if (req.Method == http.MethodPost || req.Method == http.MethodPut) && req.Body != nil {
				// rebuild io.ReadCloser to fix https://github.com/golang/go/issues/36095
				req.Body, _ = req.GetBody()
			}
		}
	} else {
		resp, err = r.client.Do(req)
	}

	r.done <- Result{Request: req, Response: resp, Err: err}
}

// Wait wait all of requests to done
func (r *Resty) Wait() []error {
	defer func() {
		// call cancelFunc, avoid to memory leak issue
		if r.cancelFunc != nil {
			r.cancelFunc()
		}
	}()

	errs := make([]error, 0, r.qty)
	done := 0

	for {

		result := <-r.done

		if r.handle != nil {
			err := r.handle(result.Request, result.Response, r.cancelFunc, result.Err)

			if err != nil {
				errs = append(errs, err)
			}
		} else {
			if result.Err != nil {
				errs = append(errs, result.Err)
			}
		}

		done++
		if done >= r.qty {
			return errs
		}
	}

}
