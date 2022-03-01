// Package resty HTTP and REST client library with parallel feature
package resty

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
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
type Handle func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error

// Option set restry option
type Option func(*Resty)

// Resty HTTP and REST client library with parallel feature
type Resty struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	qty        int
	done       chan Result

	transport          *http.Transport
	client             Client
	handle             Handle
	requestInterceptor func(req *http.Request)

	timeout time.Duration
	retry   int
	header  map[string]string
}

// DoGet execute http requests with GET method in parallel
func (r *Resty) DoGet(ctx context.Context, urls ...string) {
	r.Do(ctx, http.MethodGet, nil, urls...)
}

// DoPost execute http requests with POST method in parallel
func (r *Resty) DoPost(ctx context.Context, body io.Reader, urls ...string) {
	r.Do(ctx, http.MethodPost, body, urls...)
}

// DoPut execute http requests with PUT method in parallel
func (r *Resty) DoPut(ctx context.Context, body io.Reader, urls ...string) {
	r.Do(ctx, http.MethodPut, body, urls...)
}

// DoDelete execute http requests with DELETE method in parallel
func (r *Resty) DoDelete(ctx context.Context, urls ...string) {
	r.Do(ctx, http.MethodDelete, nil, urls...)
}

func (r *Resty) Do(ctx context.Context, method string, body io.Reader, urls ...string) {
	r.ctx, r.cancelFunc = context.WithCancel(ctx)

	r.qty = len(urls)
	r.done = make(chan Result, r.qty)

	var bodyReader io.Reader = body

	for _, url := range urls {

		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			r.done <- Result{Request: req, Response: nil, Err: err}
			continue
		}

		for key, value := range r.header {
			req.Header.Set(key, value)
		}

		req.Close = true
		//reuse http connection if it is possible
		req.Header.Set("Connection", "keep-alive")

		if r.requestInterceptor != nil {
			r.requestInterceptor(req)
		}

		go r.httpDo(req.WithContext(r.ctx))
	}
}

func (r *Resty) httpDo(req *http.Request) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func(request *http.Request) {
		var resp *http.Response
		var err error

		if r.retry > 0 {
			var bodyCopy io.ReadCloser
			if (request.Method == http.MethodPost || request.Method == http.MethodPut) && request.Body != nil {
				// clone io.ReadCloser to fix retry issue https://github.com/golang/go/issues/36095
				bodyCopy, _ = request.GetBody() //nolint: errcheck
			}

			for i := 1; ; i++ {
				resp, err = r.client.Do(request)
				//success: 200,201,202,204
				if resp != nil && (resp.StatusCode == http.StatusOK ||
					resp.StatusCode == http.StatusCreated ||
					resp.StatusCode == http.StatusAccepted ||
					resp.StatusCode == http.StatusNoContent) {
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

				if (request.Method == http.MethodPost || request.Method == http.MethodPut) && request.Body != nil {
					request.Body = bodyCopy
				}
			}
		} else {
			resp, err = r.client.Do(request.WithContext(r.ctx))
		}

		result := Result{Request: request, Response: resp, Err: err}
		if resp != nil {
			// read and close body to reuse http connection
			buf, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				result.Err = err
			} else {
				resp.Body.Close() //nolint: errcheck
				result.ResponseBody = buf
			}
		}
		r.done <- result

		wg.Done()

	}(req)

	wg.Wait()

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
			err := r.handle(result.Request, result.Response, result.ResponseBody, r.cancelFunc, result.Err)

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
