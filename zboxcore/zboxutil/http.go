package zboxutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/encryption"
	coreHttp "github.com/0chain/gosdk/core/http"
	"github.com/0chain/gosdk/core/logger"
	"github.com/hitenjain14/fasthttp"
)

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
		Transport: coreHttp.DefaultTransport,
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
	req.Header.Set("X-App-Client-ID", client.ClientID())
	req.Header.Set("X-App-Client-Key", client.PublicKey())
}

func setClientInfoWithSign(req *http.Request, sig, allocation, baseURL string) error {
	setClientInfo(req)
	req.Header.Set(CLIENT_SIGNATURE_HEADER, sig)

	hashData := allocation + baseURL
	sig2, err := client.Sign(encryption.Hash(hashData))
	if err != nil {
		return err
	}
	req.Header.Set(CLIENT_SIGNATURE_HEADER_V2, sig2)
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

func NewReferencePathRequest(baseUrl, allocationID string, allocationTx string, sig string, paths []string) (*http.Request, error) {
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

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
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

func NewObjectTreeRequest(baseUrl, allocationID string, allocationTx string, sig string, path string) (*http.Request, error) {
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

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewRefsRequest(baseUrl, allocationID, sig, allocationTx, path, pathHash, authToken, offsetPath, updatedDate, offsetDate, fileType, refType string, level, pageLimit int) (*http.Request, error) {
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

	if err := setClientInfoWithSign(req, sig, allocationID, baseUrl); err != nil {
		return nil, err
	}

	return req, nil
}

func NewRecentlyAddedRefsRequest(bUrl, allocID, allocTx, sig string, fromDate, offset int64, pageLimit int) (*http.Request, error) {
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

	if err = setClientInfoWithSign(req, sig, allocTx, bUrl); err != nil {
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

func NewCollaboratorRequest(baseUrl, allocationID, allocationTx, sig string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, COLLABORATOR_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func GetCollaboratorsRequest(baseUrl, allocationID, allocationTx, sig string, query *url.Values) (*http.Request, error) {
	u, err := joinUrl(baseUrl, COLLABORATOR_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	u.RawQuery = query.Encode()
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func DeleteCollaboratorRequest(baseUrl, allocationID, allocationTx, sig string, query *url.Values) (*http.Request, error) {
	u, err := joinUrl(baseUrl, COLLABORATOR_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewFileMetaRequest(baseUrl, allocationID, allocationTx, sig string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, FILE_META_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewFileStatsRequest(baseUrl, allocationID, allocationTx, sig string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, FILE_STATS_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
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
func NewUploadRequestWithMethod(baseURL, allocationID, allocationTx, sig string, body io.Reader, method string) (*http.Request, error) {
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
	if err := setClientInfoWithSign(req, sig, allocationTx, baseURL); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewWriteMarkerLockRequest(
	baseURL, allocationID, allocationTx, sig, connID string) (*http.Request, error) {

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

	if err := setClientInfoWithSign(req, sig, allocationTx, baseURL); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewWriteMarkerUnLockRequest(
	baseURL, allocationID, allocationTx, sig, connID, requestTime string) (*http.Request, error) {

	u, err := joinUrl(baseURL, WM_LOCK_ENDPOINT, allocationTx, connID)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseURL); err != nil {
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

func setFastClientInfoWithSign(req *fasthttp.Request, allocation string) error {
	req.Header.Set("X-App-Client-ID", client.ClientID())
	req.Header.Set("X-App-Client-Key", client.PublicKey())

	sign, err := client.Sign(encryption.Hash(allocation))
	if err != nil {
		return err
	}
	req.Header.Set(CLIENT_SIGNATURE_HEADER, sign)

	return nil
}

func NewUploadRequest(baseUrl, allocationID, allocationTx, sig string, body io.Reader, update bool) (*http.Request, error) {
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

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewConnectionRequest(baseUrl, allocationID, allocationTx, sig string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, CREATE_CONNECTION_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewRenameRequest(baseUrl, allocationID, allocationTx, sig string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, RENAME_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	// url := fmt.Sprintf("%s%s%s", baseUrl, RENAME_ENDPOINT, allocation)
	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewCopyRequest(baseUrl, allocationID, allocationTx, sig string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, COPY_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewMoveRequest(baseUrl, allocationID, allocationTx, sig string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, MOVE_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
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

	sig, err := client.Sign(encryption.Hash(allocationTx))
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
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
	req.Header.Set("X-App-Client-ID", client.ClientID())
	req.Header.Set("X-App-Client-Key", client.PublicKey())

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

func NewDeleteRequest(baseUrl, allocationID, allocationTx, sig string, query *url.Values) (*http.Request, error) {
	u, err := joinUrl(baseUrl, UPLOAD_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewCreateDirRequest(baseUrl, allocationID, allocationTx, sig string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, DIR_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewShareRequest(baseUrl, allocationID, allocationTx, sig string, body io.Reader) (*http.Request, error) {
	u, err := joinUrl(baseUrl, SHARE_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewRevokeShareRequest(baseUrl, allocationID, allocationTx, sig string, query *url.Values) (*http.Request, error) {
	u, err := joinUrl(baseUrl, SHARE_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}
	u.RawQuery = query.Encode()
	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
		return nil, err
	}

	req.Header.Set(ALLOCATION_ID_HEADER, allocationID)

	return req, nil
}

func NewWritemarkerRequest(baseUrl, allocationID, allocationTx, sig string) (*http.Request, error) {

	nurl, err := joinUrl(baseUrl, LATEST_WRITE_MARKER_ENDPOINT, allocationTx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, nurl.String(), nil)
	if err != nil {
		return nil, err
	}

	if err := setClientInfoWithSign(req, sig, allocationTx, baseUrl); err != nil {
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

func HttpDo(ctx context.Context, cncl context.CancelFunc, req *http.Request, f func(*http.Response, error) error) error {
	// Run the HTTP request in a goroutine and pass the response to f.
	c := make(chan error, 1)
	go func() {
		var err error
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

	defer cncl() // Ensure the cancellation function is deferred to release resources.

	select {
	case <-ctx.Done():
		// If the context is canceled or times out, return the context's error.
		<-c // Wait for the goroutine to complete before returning.
		return ctx.Err()
	case err := <-c:
		return err
	}
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
