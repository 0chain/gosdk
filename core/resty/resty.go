package resty

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/0chain/gosdk/core/sys"
)

func clone(m map[string]string) map[string]string {
	cl := make(map[string]string, len(m))
	for k, v := range m {
		cl[k] = v
	}

	return cl
}

// New create a Resty instance.
func New(opts ...Option) *Resty {
	r := &Resty{
		// Default timeout to use for HTTP requests when either the parent context doesn't have a timeout set
		// or the context's timeout is longer than DefaultRequestTimeout.
		timeout: DefaultRequestTimeout,
		retry:   DefaultRetry,
		header:  clone(DefaultHeader),
	}

	for _, option := range opts {
		option(r)
	}

	if r.transport == nil {
		if DefaultTransport == nil {
			DefaultTransport = &http.Transport{
				Dial: (&net.Dialer{
					Timeout: DefaultDialTimeout,
				}).Dial,
				TLSHandshakeTimeout: DefaultDialTimeout,
			}
		}
		r.transport = DefaultTransport
	}

	if r.client == nil {
		r.client = CreateClient(r.transport, r.timeout)
	}

	return r
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
	requestInterceptor func(req *http.Request) error

	timeout time.Duration
	retry   int
	header  map[string]string
}

// Then is used to call the handle function when the request has completed processing
func (r *Resty) Then(fn Handle) *Resty {
	if r == nil {
		return r
	}
	r.handle = fn
	return r
}

// DoGet executes http requests with GET method in parallel
func (r *Resty) DoGet(ctx context.Context, urls ...string) *Resty {
	return r.Do(ctx, http.MethodGet, nil, urls...)
}

// DoPost executes http requests with POST method in parallel
func (r *Resty) DoPost(ctx context.Context, body io.Reader, urls ...string) *Resty {
	return r.Do(ctx, http.MethodPost, body, urls...)
}

// DoPut executes http requests with PUT method in parallel
func (r *Resty) DoPut(ctx context.Context, body io.Reader, urls ...string) *Resty {
	return r.Do(ctx, http.MethodPut, body, urls...)
}

// DoDelete executes http requests with DELETE method in parallel
func (r *Resty) DoDelete(ctx context.Context, urls ...string) *Resty {
	return r.Do(ctx, http.MethodDelete, nil, urls...)
}

func (r *Resty) Do(ctx context.Context, method string, body io.Reader, urls ...string) *Resty {
	r.ctx, r.cancelFunc = context.WithCancel(ctx)

	r.qty = len(urls)
	r.done = make(chan Result, r.qty)

	bodyReader := body

	for _, url := range urls {
		go func(url string) {
			req, err := http.NewRequest(method, url, bodyReader)
			if err != nil {
				r.done <- Result{Request: req, Response: nil, Err: err}
				return
			}
			for key, value := range r.header {
				req.Header.Set(key, value)
			}
			// re-use http connection if it is possible
			req.Header.Set("Connection", "keep-alive")

			if r.requestInterceptor != nil {
				if err := r.requestInterceptor(req); err != nil {
					r.done <- Result{Request: req, Response: nil, Err: err}
					return
				}
			}
			r.httpDo(req.WithContext(r.ctx))
		}(url)
	}
	return r
}

func (r *Resty) httpDo(req *http.Request) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func(request *http.Request) {
		defer wg.Done()
		var resp *http.Response
		var err error
		if r.retry > 0 {
			for i := 1; ; i++ {
				var bodyCopy io.ReadCloser
				if (request.Method == http.MethodPost || request.Method == http.MethodPut) && request.Body != nil {
					// clone io.ReadCloser to fix retry issue https://github.com/golang/go/issues/36095
					bodyCopy, _ = request.GetBody() //nolint: errcheck
				}

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
					sys.Sleep(1 * time.Second)
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
	}(req)
	wg.Wait()
}

// Wait waits for all the requests to be completed
func (r *Resty) Wait() []error {
	defer func() {
		// call cancelFunc, to avoid to memory leaks
		if r.cancelFunc != nil {
			r.cancelFunc()
		}
	}()
	errs := make([]error, 0, r.qty)
	done := 0

	// no urls
	if r.qty == 0 {
		return errs
	}
	for {
		select {
		case <-r.ctx.Done():
			if r.ctx.Err() == context.DeadlineExceeded {
				return []error{r.ctx.Err()}
			}
			return errs

		case result := <-r.done:
			if r.handle != nil {
				err := r.handle(result.Request, result.Response, result.ResponseBody, r.cancelFunc, result.Err)
				if err != nil {
					errs = append(errs, err)
				} else if result.Err != nil {
					// if r.handle doesn't return any error, then append result.Err if it is not nil
					errs = append(errs, result.Err)
				}
			} else {
				if result.Err != nil {
					errs = append(errs, result.Err)
				}
			}
		}
		done++
		if done >= r.qty {
			return errs
		}
	}
}

// First successful result or errors
func (r *Resty) First() []error {
	defer func() {
		// call cancelFunc, avoid to memory leak issue
		if r.cancelFunc != nil {
			r.cancelFunc()
		}
	}()

	errs := make([]error, 0, r.qty)
	done := 0

	// no urls
	if r.qty == 0 {
		return errs
	}

	for {
		select {
		case <-r.ctx.Done():
			return errs

		case result := <-r.done:

			if r.handle != nil {
				err := r.handle(result.Request, result.Response, result.ResponseBody, r.cancelFunc, result.Err)

				if err != nil {
					errs = append(errs, err)
				} else {
					return nil
				}
			} else {
				if result.Err != nil {
					errs = append(errs, result.Err)
				} else {
					return nil
				}

			}
		}

		done++

		if done >= r.qty {
			return errs
		}

	}
}
