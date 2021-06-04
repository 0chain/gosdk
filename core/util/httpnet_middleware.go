package util

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"gopkg.in/yaml.v2"

	"github.com/0chain/gosdk/zboxcore/logger"
)

//StartDevServer mocking API requests
func StartDevServer(config string) MiddlewareFunc {
	dev := &DevServer{}

	file, err := ioutil.ReadFile(config)

	if err != nil {
		logger.Logger.Error("[httpnet]", err)
	}

	err = yaml.Unmarshal(file, &dev)

	if err != nil {
		logger.Logger.Error("[httpnet]", err)
	}

	return func(c *HTTPNetConext) {
		if c.req != nil && len(dev.Items) > 0 {

			resp, ok := dev.filter(c.req)

			if ok {
				c.resp = resp
				c.Abort()
			} else {
				c.Next()
				logger.Logger.Info("[httpnet] PASS ", c.req.Method, " ", c.req.URL.String(), " ", c.resp.StatusCode)
			}

		}

	}
}

//DevServer development server of 0chain's APIs
type DevServer struct {
	Items map[string]APIRequest
}

// APIRequest api request
type APIRequest struct {
	GET    *APIResponse
	POST   *APIResponse
	DELETE *APIResponse
	PUT    *APIResponse
}

//APIResponse http
type APIResponse struct {
	Body       string
	StatusCode int
	Header     http.Header
}

// filter try intercept request based on configuration
func (dev *DevServer) filter(req *http.Request) (*http.Response, bool) {

	if req == nil {
		return nil, false
	}

	target := req.URL.String()

	it, ok := dev.Items[target]

	if ok {
		switch req.Method {
		case http.MethodGet:
			if it.GET != nil {
				return dev.mockResponse(req, target, it.GET.StatusCode, it.GET.Body, it.GET.Header), true
			}
		case http.MethodPost:
			if it.POST != nil {
				return dev.mockResponse(req, target, it.POST.StatusCode, it.POST.Body, it.POST.Header), true
			}
		case http.MethodPut:
			if it.PUT != nil {
				return dev.mockResponse(req, target, it.PUT.StatusCode, it.PUT.Body, it.PUT.Header), true
			}
		case http.MethodDelete:
			if it.DELETE != nil {
				return dev.mockResponse(req, target, it.PUT.StatusCode, it.PUT.Body, it.PUT.Header), true
			}

		}
	}

	return nil, false

}

func (dev *DevServer) mockResponse(req *http.Request, target string, statusCode int, body string, header http.Header) *http.Response {
	logger.Logger.Info("[httpnet] MOCK ", req.Method, " ", target, " -> ", body)

	return &http.Response{
		StatusCode: statusCode,
		// Proto:         "HTTP/1.1",
		// ProtoMajor:    1,
		// ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
		Request:       req,
		Header:        header,
	}

}
