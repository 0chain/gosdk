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
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/hitenjain14/fasthttp"
)

const SC_REST_API_URL = "v1/screst/"

const MAX_RETRIES = 5
const SLEEP_BETWEEN_RETRIES = 5

// In percentage
const consensusThresh = float32(25.0)

// SCRestAPIHandler is a function type to handle the response from the SC Rest API
//
//	`response` - the response from the SC Rest API
//	`numSharders` - the number of sharders that responded
//	`err` - the error if any
type SCRestAPIHandler func(response map[string][]byte, numSharders int, err error)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type FastClient interface {
	DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error
}

var (
	Client         HttpClient
	FastHttpClient FastClient
	log            logger.Logger
)

const (
	respBodyPoolLimit = 1024 * 1024 * 16 //16MB
)

func GetLogger() *logger.Logger {
	return &log
}

const (
	ALLOCATION_ENDPOINT          = "/allocation"
	UPLOAD_ENDPOINT              = "/v1/file/upload/"
	RENAME_ENDPOINT              = "/v1/file/rename/"
	COPY_ENDPOINT                = "/v1/file/copy/"
	MOVE_ENDPOINT                = "/v1/file/move/"
	LIST_ENDPOINT                = "/v1/file/list/"
	REFERENCE_ENDPOINT           = "/v1/file/referencepath/"
	CONNECTION_ENDPOINT          = "/v1/connection/details/"
	COMMIT_ENDPOINT              = "/v1/connection/commit/"
	DOWNLOAD_ENDPOINT            = "/v1/file/download/"
	LATEST_READ_MARKER           = "/v1/readmarker/latest"
	FILE_META_ENDPOINT           = "/v1/file/meta/"
	FILE_STATS_ENDPOINT          = "/v1/file/stats/"
	OBJECT_TREE_ENDPOINT         = "/v1/file/objecttree/"
	REFS_ENDPOINT                = "/v1/file/refs/"
	RECENT_REFS_ENDPOINT         = "/v1/file/refs/recent/"
	COLLABORATOR_ENDPOINT        = "/v1/file/collaborator/"
	CALCULATE_HASH_ENDPOINT      = "/v1/file/calculatehash/"
	SHARE_ENDPOINT               = "/v1/marketplace/shareinfo/"
	DIR_ENDPOINT                 = "/v1/dir/"
	PLAYLIST_LATEST_ENDPOINT     = "/v1/playlist/latest/"
	PLAYLIST_FILE_ENDPOINT       = "/v1/playlist/file/"
	WM_LOCK_ENDPOINT             = "/v1/writemarker/lock/"
	CREATE_CONNECTION_ENDPOINT   = "/v1/connection/create/"
	LATEST_WRITE_MARKER_ENDPOINT = "/v1/file/latestwritemarker/"
	ROLLBACK_ENDPOINT            = "/v1/connection/rollback/"
	REDEEM_ENDPOINT              = "/v1/connection/redeem/"

	// CLIENT_SIGNATURE_HEADER represents http request header contains signature.
	CLIENT_SIGNATURE_HEADER    = "X-App-Client-Signature"
	CLIENT_SIGNATURE_HEADER_V2 = "X-App-Client-Signature-V2"
	ALLOCATION_ID_HEADER       = "ALLOCATION-ID"
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

func GetFastHTTPClient() *fasthttp.Client {
	fc, ok := FastHttpClient.(*fasthttp.Client)
	if ok {
		return fc
	}
	return nil
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

	FastHttpClient = &fasthttp.Client{
		MaxIdleConnDuration:           45 * time.Second,
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		// increase DNS cache time to an hour instead of default minute
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
		ReadTimeout:         180 * time.Second,
		WriteTimeout:        180 * time.Second,
		MaxConnDuration:     45 * time.Second,
		MaxResponseBodySize: 1024 * 1024 * 64, //64MB
		MaxConnsPerHost:     1024,
	}
	fasthttp.SetBodySizePoolLimit(respBodyPoolLimit, respBodyPoolLimit)
	envProxy.initialize()
	log.Init(logger.DEBUG, "0box-sdk")
}

func NewHTTPRequest(method string, url string, data []byte) (*http.Request, context.Context, context.CancelFunc, error) {
	var (
		req *http.Request
		err error
	)
	if len(data) > 0 {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(data))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Access-Control-Allow-Origin", "*")
	ctx, cncl := context.WithTimeout(context.Background(), time.Second*10)
	return req, ctx, cncl, err
}

func setClientInfo(req *http.Request) {
	req.Header.Set("X-App-Client-ID", client.GetClientID())
	req.Header.Set("X-App-Client-Key", client.GetClientPublicKey())
}

func setClientInfoWithSign(req *http.Request, allocation, baseURL string) error {
	setClientInfo(req)

	hashData := allocation
	sign, err := client.Sign(encryption.Hash(hashData))
	if err != nil {
		return err
	}
	req.Header.Set(CLIENT_SIGNATURE_HEADER, sign)

	hashData = allocation + baseURL
	sign, err = client.Sign(encryption.Hash(hashData))
	if err != nil {
		return err
	}
	req.Header.Set(CLIENT_SIGNATURE_HEADER_V2, sign)
	return nil
}

func setFastClientInfoWithSign(req *fasthttp.Request, allocation string) error {
	req.Header.Set("X-App-Client-ID", client.GetClientID())
	req.Header.Set("X-App-Client-Key", client.GetClientPublicKey())

	sign, err := client.Sign(encryption.Hash(allocation))
	if err != nil {
		return err
	}
	req.Header.Set(CLIENT_SIGNATURE_HEADER, sign)

	return nil
}

func NewCommitRequest(baseUrl, allocationID string, allocationTx string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, COMMIT_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}
	setClientInfo(req)

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewReferencePathRequest(baseUrl, allocationID string, allocationTx string, paths []string) (*http.Request, error) {
	nurl, err := joinUrl(baseUrl, REFERENCE_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

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

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewCalculateHashRequest(baseUrl, allocationID string, allocationTx string, paths []string) (*http.Request, error) {
	nurl, err := joinUrl(baseUrl, CALCULATE_HASH_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
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

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewObjectTreeRequest(baseUrl, allocationID string, allocationTx string, path string) (*http.Request, error) {
	nurl, err := joinUrl(baseUrl, OBJECT_TREE_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("path", path)
	//url := fmt.Sprintf("%s%s%s?path=%s", baseUrl, LIST_ENDPOINT, allocation, path)
	nurl.RawQuery = params.Encode() // Escape Query Parameters
	req, err := http.NewRequest(http.MethodGet, nurl.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewRefsRequest(baseUrl, allocationID, allocationTx, path, pathHash, authToken, offsetPath, updatedDate, offsetDate, fileType, refType string, level, pageLimit int) (*http.Request, error) {
	nUrl, err := joinUrl(baseUrl, REFS_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("path", path)
	params.Add("path_hash", pathHash)
	params.Add("auth_token", authToken)
	params.Add("offsetPath", offsetPath)
	params.Add("pageLimit", strconv.Itoa(pageLimit))
	params.Add("updatedDate", updatedDate)
	params.Add("offsetDate", offsetDate)
	params.Add("fileType", fileType)
	params.Add("refType", refType)
	params.Add("level", strconv.Itoa(level))
	nUrl.RawQuery = params.Encode()
	req, err := http.NewRequest(http.MethodGet, nUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	return req, nil
}

func NewRecentlyAddedRefsRequest(bUrl, allocID, allocTx string, fromDate, offset int64, pageLimit int) (*http.Request, error) {
	nUrl, err := joinUrl(bUrl, RECENT_REFS_ENDPOINT, allocID)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("limit", strconv.Itoa(pageLimit))
	params.Add("offset", strconv.FormatInt(offset, 10))
	params.Add("from-date", strconv.FormatInt(fromDate, 10))

	nUrl.RawQuery = params.Encode()
	req, err := http.NewRequest(http.MethodGet, nUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocID)

	if err := setClientInfoWithSign(req, allocTx, bUrl); err != nil {
		return nil, err
	}

	return req, nil
}

func NewAllocationRequest(baseUrl, allocationID, allocationTx string) (*http.Request, error) {
	nurl, err := joinUrl(baseUrl, ALLOCATION_ENDPOINT)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("id", allocationTx)
	nurl.RawQuery = params.Encode() // Escape Query Parameters
	req, err := http.NewRequest(http.MethodGet, nurl.String(), nil)
	if err != nil {
		return nil, err
	}
	setClientInfo(req)

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)
	return req, nil
}

func NewCollaboratorRequest(baseUrl string, allocationID string, allocationTx string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, COLLABORATOR_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func GetCollaboratorsRequest(baseUrl string, allocationID string, allocationTx string, query *url.Values) (*http.Request, error) {
	u, err := joinUrl(baseUrl, COLLABORATOR_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	u.RawQuery = query.Encode()
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func DeleteCollaboratorRequest(baseUrl string, allocationID string, allocationTx string, query *url.Values) (*http.Request, error) {
	u, err := joinUrl(baseUrl, COLLABORATOR_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewFileMetaRequest(baseUrl string, allocationID string, allocationTx string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, FILE_META_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewFileStatsRequest(baseUrl string, allocationID string, allocationTx string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, FILE_STATS_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewListRequest(baseUrl, allocationID, allocationTx, path, pathHash, auth_token string, list bool, offset, pageLimit int) (*http.Request, error) {
	nurl, err := joinUrl(baseUrl, LIST_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("path", path)
	params.Add("path_hash", pathHash)
	params.Add("auth_token", auth_token)
	if list {
		params.Add("list", "true")
	}
	params.Add("offset", strconv.Itoa(offset))
	params.Add("limit", strconv.Itoa(pageLimit))
	nurl.RawQuery = params.Encode() // Escape Query Parameters
	req, err := http.NewRequest(http.MethodGet, nurl.String(), nil)
	if err != nil {
		return nil, err
	}
	setClientInfo(req)

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

// NewUploadRequestWithMethod create a http request of upload
func NewUploadRequestWithMethod(baseURL, allocationID string, allocationTx string, body io.Reader, method string) (*http.Request, error) {
	u, err := joinUrl(baseURL, UPLOAD_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	var req *http.Request

	req, err = http.NewRequest(method, u.String(), body)

	if err != nil {
		return nil, err
	}

	// set header: X-App-Client-Signature
	if err := setClientInfoWithSign(req, allocationTx, baseURL); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewWriteMarkerLockRequest(
	baseURL, allocationID, allocationTx, connID string) (*http.Request, error) {

	u, err := joinUrl(baseURL, WM_LOCK_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("connection_id", connID)
	u.RawQuery = params.Encode() // Escape Query Parameters

	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseURL); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewWriteMarkerUnLockRequest(
	baseURL, allocationID, allocationTx, connID, requestTime string) (*http.Request, error) {

	u, err := joinUrl(baseURL, WM_LOCK_ENDPOINT, allocationTx, connID)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseURL); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewFastUploadRequest(baseURL, allocationID string, allocationTx string, body []byte, method string) (*fasthttp.Request, error) {
	u, err := joinUrl(baseURL, UPLOAD_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	req := fasthttp.AcquireRequest()

	req.Header.SetMethod(method)
	req.SetRequestURI(u.String())
	req.SetBodyRaw(body)

	// set header: X-App-Client-Signature
	if err := setFastClientInfoWithSign(req, allocationTx); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)
	return req, nil
}

func NewUploadRequest(baseUrl, allocationID string, allocationTx string, body io.Reader, update bool) (*http.Request, error) {
	u, err := joinUrl(baseUrl, UPLOAD_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	var req *http.Request
	if update {
		req, err = http.NewRequest(http.MethodPut, u.String(), body)
	} else {
		req, err = http.NewRequest(http.MethodPost, u.String(), body)
	}
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewConnectionRequest(baseUrl, allocationID string, allocationTx string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, CREATE_CONNECTION_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewRenameRequest(baseUrl, allocationID string, allocationTx string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, RENAME_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	// url := fmt.Sprintf("%s%s%s", baseUrl, RENAME_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewCopyRequest(baseUrl, allocationID string, allocationTx string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, COPY_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewMoveRequest(baseUrl, allocationID string, allocationTx string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, MOVE_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewDownloadRequest(baseUrl, allocationID, allocationTx string) (*http.Request, error) {
	u, err := joinUrl(baseUrl, DOWNLOAD_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	// url := fmt.Sprintf("%s%s%s", baseUrl, DOWNLOAD_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewFastDownloadRequest(baseUrl, allocationID, allocationTx string) (*fasthttp.Request, error) {
	u, err := joinUrl(baseUrl, DOWNLOAD_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	// url := fmt.Sprintf("%s%s%s", baseUrl, DOWNLOAD_ENDPOINT, allocation)
	// req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	// if err != nil {
	// 	return nil, err
	// }
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(u.String())
	req.Header.Set("X-App-Client-ID", client.GetClientID())
	req.Header.Set("X-App-Client-Key", client.GetClientPublicKey())

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewRedeemRequest(baseUrl, allocationID, allocationTx string) (*http.Request, error) {
	u, err := joinUrl(baseUrl, REDEEM_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return nil, err
	}
	setClientInfo(req)
	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)
	return req, nil
}

func NewDeleteRequest(baseUrl, allocationID string, allocationTx string, query *url.Values) (*http.Request, error) {
	u, err := joinUrl(baseUrl, UPLOAD_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewCreateDirRequest(baseUrl, allocationID string, allocationTx string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, DIR_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewShareRequest(baseUrl, allocationID string, allocationTx string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, SHARE_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewRevokeShareRequest(baseUrl, allocationID string, allocationTx string, query *url.Values) (*http.Request, error) {
	u, err := joinUrl(baseUrl, SHARE_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	u.RawQuery = query.Encode()
	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewWritemarkerRequest(baseUrl, allocationID, allocationTx string) (*http.Request, error) {

	nurl, err := joinUrl(baseUrl, LATEST_WRITE_MARKER_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, nurl.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewRollbackRequest(baseUrl, allocationID string, allocationTx string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, ROLLBACK_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}
	setClientInfo(req)

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

// MakeSCRestAPICall makes a rest api call to the sharders.
//   - scAddress is the address of the smart contract
//   - relativePath is the relative path of the api
//   - params is the query parameters
//   - handler is the handler function to handle the response
func MakeSCRestAPICall(scAddress string, relativePath string, params map[string]string, handler SCRestAPIHandler) ([]byte, error) {
	numSharders := len(blockchain.GetSharders())
	sharders := blockchain.GetSharders()
	responses := make(map[int]int)
	mu := &sync.Mutex{}
	entityResult := make(map[string][]byte)
	var retObj []byte
	maxCount := 0
	dominant := 200
	wg := sync.WaitGroup{}

	cfg, err := conf.GetClientConfig()
	if err != nil {
		return nil, err
	}

	for _, sharder := range sharders {
		wg.Add(1)
		go func(sharder string) {
			defer wg.Done()
			urlString := fmt.Sprintf("%v/%v%v%v", sharder, SC_REST_API_URL, scAddress, relativePath)
			urlObj, err := url.Parse(urlString)
			if err != nil {
				log.Error(err)
				return
			}
			q := urlObj.Query()
			for k, v := range params {
				q.Add(k, v)
			}
			urlObj.RawQuery = q.Encode()
			client := &http.Client{Transport: DefaultTransport}
			response, err := client.Get(urlObj.String())
			if err != nil {
				blockchain.Sharders.Fail(sharder)
				return
			}

			defer response.Body.Close()
			entityBytes, _ := ioutil.ReadAll(response.Body)
			mu.Lock()
			if response.StatusCode > http.StatusBadRequest {
				blockchain.Sharders.Fail(sharder)
			} else {
				blockchain.Sharders.Success(sharder)
			}
			responses[response.StatusCode]++
			if responses[response.StatusCode] > maxCount {
				maxCount = responses[response.StatusCode]
			}

			if isCurrentDominantStatus(response.StatusCode, responses, maxCount) {
				dominant = response.StatusCode
				retObj = entityBytes
			}

			entityResult[sharder] = entityBytes
			blockchain.Sharders.Success(sharder)
			mu.Unlock()
		}(sharder)
	}
	wg.Wait()

	rate := float32(maxCount*100) / float32(cfg.SharderConsensous)
	if rate < consensusThresh {
		err = errors.New("consensus_failed", "consensus failed on sharders")
	}

	if dominant != 200 {
		var objmap map[string]json.RawMessage
		err := json.Unmarshal(retObj, &objmap)
		if err != nil {
			return nil, errors.New("", string(retObj))
		}

		var parsed string
		err = json.Unmarshal(objmap["error"], &parsed)
		if err != nil || parsed == "" {
			return nil, errors.New("", string(retObj))
		}

		return nil, errors.New("", parsed)
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
	go func() {
		var err error
		// indefinitely try if io.EOF error occurs. As per some research over google
		// it occurs when client http tries to send byte stream in connection that is
		// closed by the server
		for {
			var resp *http.Response
			resp, err = Client.Do(req.WithContext(ctx))
			if errors.Is(err, io.EOF) {
				continue
			}

			err = f(resp, err)
			break
		}
		c <- err
	}()

	// TODO: Check cncl context required in any case
	// defer cncl()
	select {
	case <-ctx.Done():
		DefaultTransport.CancelRequest(req) //nolint
		<-c                                 // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}

// isCurrentDominantStatus determines whether the current response status is the dominant status among responses.
//
// The dominant status is where the response status is counted the most.
// On tie-breakers, 200 will be selected if included.
//
// Function assumes runningTotalPerStatus can be accessed safely concurrently.
func isCurrentDominantStatus(respStatus int, currentTotalPerStatus map[int]int, currentMax int) bool {
	// mark status as dominant if
	// - running total for status is the max and response is 200 or
	// - running total for status is the max and count for 200 is lower
	return currentTotalPerStatus[respStatus] == currentMax && (respStatus == 200 || currentTotalPerStatus[200] < currentMax)
}

func joinUrl(baseURl string, paths ...string) (*url.URL, error) {
	u, err := url.Parse(baseURl)
	if err != nil {
		return nil, err
	}
	p := path.Join(paths...)
	u.Path = path.Join(u.Path, p)
	return u, nil
}
