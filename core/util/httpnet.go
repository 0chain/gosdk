package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type GetRequest struct {
	*PostRequest
}

type GetResponse struct {
	*PostResponse
}

type PostRequest struct {
	req  *http.Request
	Ctx  context.Context
	cncl context.CancelFunc
	url  string
}

type PostResponse struct {
	Url        string
	StatusCode int
	Status     string
	Body       string
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var Client HttpClient

func getEnvAny(names ...string) string {
	for _, n := range names {
		if val := os.Getenv(n); val != "" {
			return val
		}
	}
	return ""
}

func (pfe *proxyFromEnv) initialize() {
	pfe.HTTPProxy = getEnvAny("HTTP_PROXY", "http_proxy")
	pfe.HTTPSProxy = getEnvAny("HTTPS_PROXY", "https_proxy")
	pfe.NoProxy = getEnvAny("NO_PROXY", "no_proxy")

	if pfe.NoProxy != "" {
		return
	}

	if pfe.HTTPProxy != "" {
		pfe.http, _ = url.Parse(pfe.HTTPProxy)
	}
	if pfe.HTTPSProxy != "" {
		pfe.https, _ = url.Parse(pfe.HTTPSProxy)
	}
}

type proxyFromEnv struct {
	HTTPProxy  string
	HTTPSProxy string
	NoProxy    string

	http, https *url.URL
}

var envProxy proxyFromEnv

func init() {
	Client = &http.Client{
		Transport: transport,
	}
	envProxy.initialize()
}

func httpDo(req *http.Request, ctx context.Context, cncl context.CancelFunc, f func(*http.Response, error) error) error {
	c := make(chan error, 1)

	go func() { c <- f(Client.Do(req.WithContext(ctx))) }()

	select {
	case <-ctx.Done():
		// Use the cancel function only after trying to get the result.
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		// Ensure that we call cncl after we are done with the response
		defer cncl() // Move this here to ensure we cancel after processing
		return err
	}
}

// NewHTTPGetRequest create a GetRequest instance with 60s timeout
func NewHTTPGetRequest(url string) (*GetRequest, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
	go func() {
		//call cancel to avoid memory leak here
		<-ctx.Done()

		cancel()

	}()

	return NewHTTPGetRequestContext(ctx, url)
}

// NewHTTPGetRequestContext create a GetRequest with context and url
func NewHTTPGetRequestContext(ctx context.Context, url string) (*GetRequest, error) {

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Access-Control-Allow-Origin", "*")

	gr := new(GetRequest)
	gr.PostRequest = &PostRequest{}
	gr.url = url
	gr.req = req
	gr.Ctx, gr.cncl = context.WithCancel(ctx)
	return gr, nil
}

func NewHTTPPostRequest(url string, data interface{}) (*PostRequest, error) {
	pr := &PostRequest{}
	jsonByte, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonByte))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Access-Control-Allow-Origin", "*")
	pr.url = url
	pr.req = req
	pr.Ctx, pr.cncl = context.WithTimeout(context.Background(), time.Second*60)
	return pr, nil
}

func (r *GetRequest) Get() (*GetResponse, error) {
	response := &GetResponse{}
	presp, err := r.Post()
	if err != nil {
		return nil, err // Return early if there's an error
	}
	response.PostResponse = presp
	return response, nil
}

func (r *PostRequest) Post() (*PostResponse, error) {
	result := &PostResponse{}
	err := httpDo(r.req, r.Ctx, r.cncl, func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		if resp.Body != nil {
			defer resp.Body.Close()
		} else {
			return fmt.Errorf("response body is nil")
		}

		rspBy, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		result.Url = r.url
		result.StatusCode = resp.StatusCode
		result.Status = resp.Status
		result.Body = string(rspBy)
		return nil
	})
	if err != nil {
		return nil, err // Ensure you propagate the error
	}
	return result, nil
}
