package client

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/util"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/shopspring/decimal"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
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

// MakeSCRestAPICall makes a rest api call to the sharders.
//   - scAddress is the address of the smart contract
//   - relativePath is the relative path of the api
//   - params is the query parameters
//   - handler is the handler function to handle the response
func MakeSCRestAPICall(scAddress string, relativePath string, params map[string]string, restApiUrls ...string) ([]byte, error) {
	const (
		consensusThresh = float32(25.0)
	)

	restApiUrl := ScRestApiUrl
	if len(restApiUrls) > 0 {
		restApiUrl = restApiUrls[0]
	}

	sharders := nodeClient.Network().Sharders
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
			urlString := fmt.Sprintf("%v/%v%v%v", sharder, restApiUrl, scAddress, relativePath)
			urlObj, err := url.Parse(urlString)
			if err != nil {
				log.Println(err)
				return
			}
			q := urlObj.Query()
			for k, v := range params {
				q.Add(k, v)
			}
			urlObj.RawQuery = q.Encode()
			client := &http.Client{Transport: DefaultTransport}
			urlStr := urlObj.String()
			response, err := client.Get(urlStr)
			if err != nil {
				fmt.Println("Failing url:", urlStr, "on sharder:", sharder, "Error:", err)
				nodeClient.sharders.Fail(sharder)
				return
			}

			fmt.Println("Success url:", urlStr, "on sharder:", sharder)

			defer response.Body.Close()
			entityBytes, _ := io.ReadAll(response.Body)
			mu.Lock()
			if response.StatusCode > http.StatusBadRequest {
				nodeClient.sharders.Fail(sharder)
			} else {
				nodeClient.sharders.Success(sharder)
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
			nodeClient.sharders.Success(sharder)
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

	if rate > consensusThresh {
		return retObj, nil
	}
	return nil, err
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

// hashAndBytesOfReader computes hash of readers data and returns hash encoded to hex and bytes of reader data.
// If error occurs while reading data from reader, it returns non nil error.
func hashAndBytesOfReader(r io.Reader) (hash string, reader []byte, err error) { //nolint:unused
	h := sha1.New()
	teeReader := io.TeeReader(r, h)
	readerBytes, err := io.ReadAll(teeReader)
	if err != nil {
		return "", nil, err
	}

	return hex.EncodeToString(h.Sum(nil)), readerBytes, nil
}

// extractSharders returns string slice of randomly ordered sharders existing in the current network.
func extractSharders() []string { //nolint:unused
	sharders := nodeClient.Network().Sharders
	return util.GetRandom(sharders, len(sharders))
}

const (
	// ScRestApiUrl represents base URL path to execute smart contract rest points.
	ScRestApiUrl = "v1/screst/"
)

// makeScURL creates url.URL to make smart contract request to sharder.
func makeScURL(params map[string]string, sharder, restApiUrl, scAddress, relativePath string) *url.URL { //nolint:unused
	uString := fmt.Sprintf("%v/%v%v%v", sharder, restApiUrl, scAddress, relativePath)

	u, _ := url.Parse(uString)
	q := u.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()

	//log.Println("SC URL:", u.RawQuery)
	//log.Println("Sharders:", sharder)
	//log.Println("Rest API URL:", restApiUrl)
	//log.Println("SC Address:", scAddress)
	//log.Println("Relative Path:", relativePath)

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
	const GetBalance = "client/get/balance"
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

	if res, err = MakeSCRestAPICall("", GetBalance, map[string]string{
		"client_id": clientID,
	}, "v1/"); err != nil {
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
	f, _ := decimal.New(b.Balance, -10).Float64()
	return f, nil
}
