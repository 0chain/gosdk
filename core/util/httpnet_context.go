package util

import (
	"math"
	"net/http"
)

const abortIndex int8 = math.MaxInt8 / 2

var (
	defaultMiddlewareChain = make(MiddlewareChain, 0)
)

// HTTPNetConext provide a context to intercept any request from httpnet with response mockable feature
type HTTPNetConext struct {
	req  *http.Request
	resp *http.Response
	err  error

	index    int8
	handlers []MiddlewareFunc
}

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
func (c *HTTPNetConext) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}

}

// Abort prevents pending handlers from being called.
func (c *HTTPNetConext) Abort() {
	c.index = abortIndex
}

// Get call middleware chain, and sends HTTP GET request and returns an HTTP response
func (c *HTTPNetConext) Get(url string, client *http.Client) (*http.Response, error) {

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req, client.Do)
}

// Do call middleware chain, and sends an HTTP request and returns an HTTP response
func (c *HTTPNetConext) Do(req *http.Request, do func(req *http.Request) (*http.Response, error)) (*http.Response, error) {

	c.req = req

	c.handlers = append(c.handlers, func(hc *HTTPNetConext) {
		hc.resp, hc.err = do(hc.req)
		hc.Next()
	})

	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}

	return c.resp, c.err
}

// NewHTTPNetContext create a new HTTPNetContext with default midlleware chain
func NewHTTPNetContext() *HTTPNetConext {
	c := &HTTPNetConext{}
	c.handlers = make([]MiddlewareFunc, 0, len(defaultMiddlewareChain))
	c.handlers = append(c.handlers, defaultMiddlewareChain...)

	return c
}

// InterceptHTTPReqeust register interceptor in httpnet
func InterceptHTTPReqeust(interceptor MiddlewareFunc) {
	if interceptor != nil {
		defaultMiddlewareChain = append(defaultMiddlewareChain, interceptor)
	}
}

// MiddlewareFunc middleware function for http request in httpnet utils
type MiddlewareFunc func(c *HTTPNetConext)

// MiddlewareChain is a collection of middleware that will be invoked in there index order
type MiddlewareChain []MiddlewareFunc

// Use regiser middlware for httpnet utils
func Use(middlewares ...MiddlewareFunc) {
	defaultMiddlewareChain = append(defaultMiddlewareChain, middlewares...)
}
