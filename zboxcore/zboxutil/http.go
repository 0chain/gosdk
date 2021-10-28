package zboxutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
)

const SC_REST_API_URL = "v1/screst/"
const REGISTER_CLIENT = "v1/client/put"

const MAX_RETRIES = 5
const SLEEP_BETWEEN_RETRIES = 5

// In percentage
const consensusThresh = float32(25.0)

type SCRestAPIHandler func(response map[string][]byte, numSharders int, err error)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var Client HttpClient

const (
	ALLOCATION_ENDPOINT      = "/allocation"
	UPLOAD_ENDPOINT          = "/v1/file/upload/"
	ATTRS_ENDPOINT           = "/v1/file/attributes/"
	RENAME_ENDPOINT          = "/v1/file/rename/"
	COPY_ENDPOINT            = "/v1/file/copy/"
	LIST_ENDPOINT            = "/v1/file/list/"
	REFERENCE_ENDPOINT       = "/v1/file/referencepath/"
	CONNECTION_ENDPOINT      = "/v1/connection/details/"
	COMMIT_ENDPOINT          = "/v1/connection/commit/"
	DOWNLOAD_ENDPOINT        = "/v1/file/download/"
	LATEST_READ_MARKER       = "/v1/readmarker/latest"
	FILE_META_ENDPOINT       = "/v1/file/meta/"
	FILE_STATS_ENDPOINT      = "/v1/file/stats/"
	OBJECT_TREE_ENDPOINT     = "/v1/file/objecttree/"
	REFS_ENDPOINT            = "/v1/file/refs/"
	COMMIT_META_TXN_ENDPOINT = "/v1/file/commitmetatxn/"
	COLLABORATOR_ENDPOINT    = "/v1/file/collaborator/"
	CALCULATE_HASH_ENDPOINT  = "/v1/file/calculatehash/"
	SHARE_ENDPOINT           = "/v1/marketplace/shareinfo/"
	DIR_ENDPOINT             = "/v1/dir/"

	// CLIENT_SIGNATURE_HEADER represents http request header contains signature.
	CLIENT_SIGNATURE_HEADER = "X-App-Client-Signature"
)

func getEnvAny(names ...string) string {
	for _, n := range names {
		if val := os.Getenv(n); val != "" {
			return val
		}
	}
	return ""
}

type proxyFromEnv struct {
	HTTPProxy  string
	HTTPSProxy string
	NoProxy    string

	http, https *url.URL
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

func (pfe *proxyFromEnv) isLoopback(host string) (ok bool) {
	host, _, _ = net.SplitHostPort(host)
	if host == "localhost" {
		return true
	}
	return net.ParseIP(host).IsLoopback()
}

func (pfe *proxyFromEnv) Proxy(req *http.Request) (proxy *url.URL, err error) {
	if pfe.isLoopback(req.URL.Host) {
		switch req.URL.Scheme {
		case "http":
			return pfe.http, nil
		case "https":
			return pfe.https, nil
		default:
		}
	}
	return http.ProxyFromEnvironment(req)
}

var envProxy proxyFromEnv

func init() {
	Client = &http.Client{
		Transport: DefaultTransport,
	}
	envProxy.initialize()
}

var DefaultTransport = &http.Transport{
	Proxy: envProxy.Proxy,
	DialContext: (&net.Dialer{
		Timeout:   45 * time.Second,
		KeepAlive: 45 * time.Second,
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

func setClientInfo(req *http.Request) {
	req.Header.Set("X-App-Client-ID", client.GetClientID())
	req.Header.Set("X-App-Client-Key", client.GetClientPublicKey())
}

func setClientInfoWithSign(req *http.Request, allocation string) error {
	setClientInfo(req)

	sign, err := client.Sign(encryption.Hash(allocation))
	if err != nil {
		return err
	}
	req.Header.Set(CLIENT_SIGNATURE_HEADER, sign)

	return nil
}

func NewCommitRequest(baseUrl, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, COMMIT_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	setClientInfo(req)
	return req, nil
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
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
}

func NewCalculateHashRequest(baseUrl, allocation string, paths []string) (*http.Request, error) {
	nurl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	nurl.Path += CALCULATE_HASH_ENDPOINT + allocation
	pathBytes, err := json.Marshal(paths)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("paths", string(pathBytes))
	nurl.RawQuery = params.Encode() // Escape Query Parameters
	req, err := http.NewRequest(http.MethodPost, nurl.String(), nil)
	if err != nil {
		return nil, err
	}
	setClientInfo(req)
	return req, nil
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
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
}

func NewRefsRequest(baseUrl, allocationID, path, offsetPath, updatedDate, offsetDate, fileType, refType string, level, pageLimit int) (*http.Request, error) {
	nUrl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	nUrl.Path += REFS_ENDPOINT + allocationID
	params := url.Values{}
	params.Add("path", path)
	params.Add("offsetPath", offsetPath)
	params.Add("pageLimit", strconv.Itoa(pageLimit))
	params.Add("updatedDate", updatedDate)
	params.Add("offsetDate", offsetDate)
	params.Add("type", fileType)
	params.Add("refType", refType)
	params.Add("level", strconv.Itoa(level))
	nUrl.RawQuery = params.Encode()
	req, err := http.NewRequest(http.MethodGet, nUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationID); err != nil {
		return nil, err
	}

	return req, nil
}

func NewAllocationRequest(baseUrl, allocation string) (*http.Request, error) {
	nurl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	nurl.Path += ALLOCATION_ENDPOINT
	params := url.Values{}
	params.Add("id", allocation)
	nurl.RawQuery = params.Encode() // Escape Query Parameters
	req, err := http.NewRequest(http.MethodGet, nurl.String(), nil)
	if err != nil {
		return nil, err
	}
	setClientInfo(req)
	return req, nil
}

func NewCommitMetaTxnRequest(baseUrl string, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, COMMIT_META_TXN_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	setClientInfo(req)
	return req, nil
}

func NewCollaboratorRequest(baseUrl string, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, COLLABORATOR_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
}

func GetCollaboratorsRequest(baseUrl string, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, COLLABORATOR_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodGet, url, body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
}

func DeleteCollaboratorRequest(baseUrl string, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, COLLABORATOR_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodDelete, url, body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
}

func NewFileMetaRequest(baseUrl string, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, FILE_META_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	err = setClientInfoWithSign(req, allocation)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func NewFileStatsRequest(baseUrl string, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, FILE_STATS_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
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
	if err != nil {
		return nil, err
	}
	setClientInfo(req)
	return req, nil
}

// NewUploadRequestWithMethod create a http reqeust of upload
func NewUploadRequestWithMethod(baseURL, allocation string, body io.Reader, method string) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseURL, UPLOAD_ENDPOINT, allocation)
	var req *http.Request
	var err error

	req, err = http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
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
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
}

func NewAttributesRequest(baseUrl, allocation string, body io.Reader) (
	req *http.Request, err error) {

	var url = fmt.Sprintf("%s%s%s", baseUrl, ATTRS_ENDPOINT, allocation)
	req, err = http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
}

func NewRenameRequest(baseUrl, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, RENAME_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
}

func NewCopyRequest(baseUrl, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, COPY_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
}

func NewDownloadRequest(baseUrl, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, DOWNLOAD_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	setClientInfo(req)
	return req, nil
}

func NewDeleteRequest(baseUrl, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, UPLOAD_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodDelete, url, body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
}

func NewCreateDirRequest(baseUrl, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, DIR_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
}

func NewShareRequest(baseUrl, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, SHARE_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
}

func NewRevokeShareRequest(baseUrl, allocation string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s%s", baseUrl, SHARE_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodDelete, url, body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocation); err != nil {
		return nil, err
	}

	return req, nil
}

func MakeSCRestAPICall(scAddress string, relativePath string, params map[string]string, handler SCRestAPIHandler) ([]byte, error) {
	numSharders := len(blockchain.GetSharders())
	sharders := blockchain.GetSharders()
	responses := make(map[int]float32)
	entityResult := make(map[string][]byte)
	var retObj []byte
	maxCount := float32(0)
	for _, sharder := range util.Shuffle(sharders) {
		urlString := fmt.Sprintf("%v/%v%v%v", sharder, SC_REST_API_URL, scAddress, relativePath)
		urlObj, _ := url.Parse(urlString)
		q := urlObj.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		urlObj.RawQuery = q.Encode()
		client := &http.Client{Transport: DefaultTransport}

		response, err := client.Get(urlObj.String())
		if err != nil {
			continue
		} else {
			if response.StatusCode != 200 {
				continue
			}
			defer response.Body.Close()
			entityBytes, err := ioutil.ReadAll(response.Body)
			if err != nil {
				continue
			}
			responses[response.StatusCode]++
			if responses[response.StatusCode] > maxCount {
				maxCount = responses[response.StatusCode]
				retObj = entityBytes
			}
			entityResult[sharder] = retObj
		}

		var rate = maxCount * 100 / float32(numSharders)
		if rate >= consensusThresh {
			break // got it
		}
	}

	var err error
	rate := maxCount * 100 / float32(numSharders)
	if rate < consensusThresh {
		err = errors.New("consensus_failed", "consensus failed on sharders")
	}

	if handler != nil {
		handler(entityResult, numSharders, err)
	}

	if rate > consensusThresh {
		return retObj, nil
	}
	return nil, err
}

func HttpDo(ctx context.Context, cncl context.CancelFunc, req *http.Request, f func(*http.Response, error) error) error {
	// Run the HTTP request in a goroutine and pass the response to f.
	c := make(chan error, 1)
	go func() { c <- f(Client.Do(req.WithContext(ctx))) }()
	// TODO: Check cncl context required in any case
	// defer cncl()
	select {
	case <-ctx.Done():
		DefaultTransport.CancelRequest(req)
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}
