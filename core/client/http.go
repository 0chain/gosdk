package client

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/util"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/shopspring/decimal"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

var DefaultTransport = &http.Transport{
	Proxy: EnvProxy.Proxy,
	DialContext: (&net.Dialer{
		Timeout:   3 * time.Minute,
		KeepAlive: 45 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   45 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	MaxIdleConnsPerHost:   25,
}

// SCRestAPIHandler is a function type to handle the response from the SC Rest API
//
//	`response` - the response from the SC Rest API
//	`numSharders` - the number of sharders that responded
//	`err` - the error if any
type SCRestAPIHandler func(response map[string][]byte, numSharders int, err error)

const (
	// clientTimeout represents default http.Client timeout.
	clientTimeout = 10 * time.Second

	// tlsHandshakeTimeout represents default http.Transport TLS handshake timeout.
	tlsHandshakeTimeout = 5 * time.Second

	// dialTimeout represents default net.Dialer timeout.
	dialTimeout = 5 * time.Second
)

// NewClient creates default http.Client with timeouts.
func NewClient() *http.Client {
	return &http.Client{
		Timeout: clientTimeout,
		Transport: &http.Transport{
			TLSHandshakeTimeout: tlsHandshakeTimeout,
			DialContext: (&net.Dialer{
				Timeout: dialTimeout,
			}).DialContext,
		},
	}
}

// NewRetryableClient creates default retryablehttp.Client with timeouts and embedded NewClient result.
func NewRetryableClient(retryMax int) *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.HTTPClient = NewClient()
	client.RetryWaitMax = clientTimeout
	client.RetryMax = retryMax
	client.Logger = nil

	return client
}

// MakeSCRestAPICall calls smart contract with provided address
// and makes retryable request to smart contract resource with provided relative path using params.
func MakeSCRestAPICall(scAddress string, relativePath string, params map[string]string) ([]byte, error) {
	var (
		resMaxCounterBody []byte

		hashMaxCounter int
		hashCounters   = make(map[string]int)

		sharders = extractSharders()

		lastErrMsg string
	)

	for _, sharder := range sharders {
		var (
			retryableClient = NewRetryableClient(5)
			u               = makeScURL(params, sharder, scAddress, relativePath)
		)

		resp, err := retryableClient.Get(u.String())
		if err != nil {
			lastErrMsg = fmt.Sprintf("error while requesting sharders: %v", err)
			continue
		}
		hash, resBody, err := hashAndBytesOfReader(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErrMsg = fmt.Sprintf("error while reading response body: %v", err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			lastErrMsg = fmt.Sprintf("response status is not OK; response body: %s", string(resBody))
			continue
		}

		hashCounters[hash]++
		if hashCounters[hash] > hashMaxCounter {
			hashMaxCounter = hashCounters[hash]
			resMaxCounterBody = resBody
		}
	}

	if hashMaxCounter == 0 {
		return nil, errors.New("request_sharders", "no valid responses, last err: "+lastErrMsg)
	}

	return resMaxCounterBody, nil
}

// hashAndBytesOfReader computes hash of readers data and returns hash encoded to hex and bytes of reader data.
// If error occurs while reading data from reader, it returns non nil error.
func hashAndBytesOfReader(r io.Reader) (hash string, reader []byte, err error) {
	h := sha1.New()
	teeReader := io.TeeReader(r, h)
	readerBytes, err := ioutil.ReadAll(teeReader)
	if err != nil {
		return "", nil, err
	}

	return hex.EncodeToString(h.Sum(nil)), readerBytes, nil
}

// extractSharders returns string slice of randomly ordered sharders existing in the current network.
func extractSharders() []string {
	sharders := nodeClient.Network().Sharders
	return util.GetRandom(sharders, len(sharders))
}

const (
	// ScRestApiUrl represents base URL path to execute smart contract rest points.
	ScRestApiUrl = "v1/screst/"
)

// makeScURL creates url.URL to make smart contract request to sharder.
func makeScURL(params map[string]string, sharder, scAddress, relativePath string) *url.URL {
	uString := fmt.Sprintf("%v/%v%v%v", sharder, ScRestApiUrl, scAddress, relativePath)
	u, _ := url.Parse(uString)
	q := u.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()

	return u
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

var EnvProxy proxyFromEnv

type proxyFromEnv struct {
	HTTPProxy  string
	HTTPSProxy string
	NoProxy    string

	http, https *url.URL
}

func (pfe *proxyFromEnv) Initialize() {
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

func getEnvAny(names ...string) string {
	for _, n := range names {
		if val := os.Getenv(n); val != "" {
			return val
		}
	}
	return ""
}

func GetBalance(clientIDs ...string) (*GetBalanceResponse, error) {
	const GET_BALANCE = "/client/get/balance"
	var (
		balance GetBalanceResponse
		err     error
		res     []byte
	)

	var clientID string
	if len(clientIDs) > 0 {
		clientID = clientIDs[0]
	} else {
		clientID = ClientID()
	}

	if res, err = MakeSCRestAPICall("", GET_BALANCE, map[string]string{
		"client_id": clientID,
	}); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(res, &balance); err != nil {
		return nil, err
	}

	return &balance, nil
}

type GetBalanceResponse struct {
	Txn     string `json:"txn"`
	Round   int64  `json:"round"`
	Balance int64  `json:"balance"`
	Nonce   int64  `json:"nonce"`
}

// ToToken converts Balance to ZCN tokens.
func (b GetBalanceResponse) ToToken() (float64, error) {
	if b.Balance > math.MaxInt64 {
		return 0.0, errors.New("to_token failed", "value is too large")
	}

	f, _ := decimal.New(b.Balance, -10).Float64()
	return f, nil
}
