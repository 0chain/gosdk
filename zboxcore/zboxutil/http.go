package zboxutil

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
)

const SC_REST_API_URL = "v1/screst/"
const REGISTER_CLIENT = "v1/client/put"

const MAX_RETRIES = 5
const SLEEP_BETWEEN_RETRIES = 5

type SCRestAPIHandler func(response map[string][]byte, numSharders int, err error)

const UPLOAD_ENDPOINT = "/v1/file/upload/"
const RENAME_ENDPOINT = "/v1/file/rename/"
const COPY_ENDPOINT = "/v1/file/copy/"
const LIST_ENDPOINT = "/v1/file/list/"
const REFERENCE_ENDPOINT = "/v1/file/referencepath/"
const CONNECTION_ENDPOINT = "/v1/connection/details/"
const COMMIT_ENDPOINT = "/v1/connection/commit/"
const DOWNLOAD_ENDPOINT = "/v1/file/download/"
const LATEST_READ_MARKER = "/v1/readmarker/latest"
const FILE_META_ENDPOINT = "/v1/file/meta/"
const FILE_STATS_ENDPOINT = "/v1/file/stats/"
const OBJECT_TREE_ENDPOINT = "/v1/file/objecttree/"

var transport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	MaxIdleConnsPerHost:   100,
}

func NewHTTPRequest(method string, url string, data []byte) (*http.Request, context.Context, context.CancelFunc, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Access-Control-Allow-Origin", "*")
	ctx, cncl := context.WithTimeout(context.Background(), time.Second*10)
	return req, ctx, cncl, err
}

func setClientInfo(req *http.Request, err error) (*http.Request, error) {
	if err == nil {
		req.Header.Set("X-App-Client-ID", client.GetClientID())
		req.Header.Set("X-App-Client-Key", client.GetClientPublicKey())
	}
	return req, err
}

func NewCommitRequest(baseUrl, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, COMMIT_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	return setClientInfo(req, err)
}

func NewReferencePathRequest(baseUrl, allocation string, paths []string) (*http.Request, error) {
	nurl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	nurl.Path += REFERENCE_ENDPOINT + allocation
	pathBytes, err := json.Marshal(paths)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("paths", string(pathBytes))
	//url := fmt.Sprintf("%s%s%s?path=%s", baseUrl, LIST_ENDPOINT, allocation, path)
	nurl.RawQuery = params.Encode() // Escape Query Parameters
	req, err := http.NewRequest(http.MethodGet, nurl.String(), nil)
	return setClientInfo(req, err)
}

func NewObjectTreeRequest(baseUrl, allocation string, path string) (*http.Request, error) {
	nurl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	nurl.Path += OBJECT_TREE_ENDPOINT + allocation
	params := url.Values{}
	params.Add("path", path)
	//url := fmt.Sprintf("%s%s%s?path=%s", baseUrl, LIST_ENDPOINT, allocation, path)
	nurl.RawQuery = params.Encode() // Escape Query Parameters
	req, err := http.NewRequest(http.MethodGet, nurl.String(), nil)
	return setClientInfo(req, err)
}

func NewFileMetaRequest(baseUrl string, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, FILE_META_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	return setClientInfo(req, err)
}

func NewFileStatsRequest(baseUrl string, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, FILE_STATS_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	return setClientInfo(req, err)
}

func NewListRequest(baseUrl, allocation string, path string, auth_token string) (*http.Request, error) {
	nurl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	nurl.Path += LIST_ENDPOINT + allocation
	params := url.Values{}
	params.Add("path_hash", path)
	params.Add("auth_token", auth_token)
	//url := fmt.Sprintf("%s%s%s?path=%s", baseUrl, LIST_ENDPOINT, allocation, path)
	nurl.RawQuery = params.Encode() // Escape Query Parameters
	req, err := http.NewRequest(http.MethodGet, nurl.String(), nil)
	return setClientInfo(req, err)
}

func NewUploadRequest(baseUrl, allocation string, body io.Reader, update bool) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, UPLOAD_ENDPOINT, allocation)
	var req *http.Request
	var err error
	if update {
		req, err = http.NewRequest(http.MethodPut, url, body)
	} else {
		req, err = http.NewRequest(http.MethodPost, url, body)
	}
	return setClientInfo(req, err)
}

func NewRenameRequest(baseUrl, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, RENAME_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	return setClientInfo(req, err)
}

func NewCopyRequest(baseUrl, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, COPY_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	return setClientInfo(req, err)
}

func NewDownloadRequest(baseUrl, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, DOWNLOAD_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	return setClientInfo(req, err)
}

func NewDeleteRequest(baseUrl, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, UPLOAD_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodDelete, url, body)
	return setClientInfo(req, err)
}

func MakeSCRestAPICall(scAddress string, relativePath string, params map[string]string, handler SCRestAPIHandler) ([]byte, error) {
	numSharders := len(blockchain.GetSharders())
	sharders := blockchain.GetSharders()
	responses := make(map[string]int)
	entityResult := make(map[string][]byte)
	var retObj []byte
	maxCount := 0
	for _, sharder := range sharders {
		urlString := fmt.Sprintf("%v/%v%v%v", sharder, SC_REST_API_URL, scAddress, relativePath)
		urlObj, _ := url.Parse(urlString)
		q := urlObj.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		urlObj.RawQuery = q.Encode()
		h := sha1.New()
		client := &http.Client{Transport: transport}

		response, err := client.Get(urlObj.String())
		if err != nil {
			numSharders--
		} else {
			if response.StatusCode != 200 {
				continue
			}
			defer response.Body.Close()
			tReader := io.TeeReader(response.Body, h)
			entityBytes, err := ioutil.ReadAll(tReader)
			if err != nil {
				continue
			}
			hashBytes := h.Sum(nil)
			hash := hex.EncodeToString(hashBytes)
			responses[hash]++
			if responses[hash] > maxCount {
				maxCount = responses[hash]
				retObj = entityBytes
			}
			entityResult[sharder] = retObj
		}
	}
	var err error

	if maxCount <= (numSharders / 2) {
		err = common.NewError("invalid_response", "Sharder responses were invalid. Hash mismatch")
	}
	if handler != nil {
		handler(entityResult, numSharders, err)
	}
	if maxCount > (numSharders / 2) {
		return retObj, nil
	}
	return nil, err
}

func HttpDo(ctx context.Context, cncl context.CancelFunc, req *http.Request, f func(*http.Response, error) error) error {
	// Run the HTTP request in a goroutine and pass the response to f.
	client := &http.Client{Transport: transport}
	c := make(chan error, 1)
	go func() { c <- f(client.Do(req.WithContext(ctx))) }()
	// TODO: Check cncl context required in any case
	// defer cncl()
	select {
	case <-ctx.Done():
		transport.CancelRequest(req)
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}
