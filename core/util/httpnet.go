package util

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
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
	ctx  context.Context
	cncl context.CancelFunc
	url  string
}

type PostResponse struct {
	Url        string
	StatusCode int
	Status     string
	Body       string
}

// Run the HTTP request in a goroutine and pass the response to f.
var transport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          1000,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	MaxIdleConnsPerHost:   5,
}

func httpDo(req *http.Request, ctx context.Context, cncl context.CancelFunc, f func(*http.Response, error) error) error {
	client := &http.Client{Transport: transport}
	c := make(chan error, 1)
	go func() { c <- f(client.Do(req.WithContext(ctx))) }()
	defer cncl()
	select {
	case <-ctx.Done():
		transport.CancelRequest(req)
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func NewHTTPGetRequest(url string) (*GetRequest, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Access-Control-Allow-Origin", "*")
	pr := &GetRequest{}
	pr.PostRequest = &PostRequest{}
	pr.url = url
	pr.req = req
	pr.ctx, pr.cncl = context.WithTimeout(context.Background(), time.Second*60)
	return pr, nil
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
	pr.ctx, pr.cncl = context.WithTimeout(context.Background(), time.Second*60)
	return pr, nil
}

func (r *GetRequest) Get() (*GetResponse, error) {
	response := &GetResponse{}
	presp, err := r.Post()
	response.PostResponse = presp
	return response, err
}

func (r *PostRequest) Post() (*PostResponse, error) {
	result := &PostResponse{}
	err := httpDo(r.req, r.ctx, r.cncl, func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		if resp.Body != nil {
			defer resp.Body.Close()
		}

		rspBy, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		result.Url = r.url
		result.StatusCode = resp.StatusCode
		result.Status = resp.Status
		result.Body = string(rspBy)
		return nil
	})
	return result, err
}
