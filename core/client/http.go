package client

import (
	"encoding/json"
	"fmt"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/logger"
	"io/ioutil"
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

const SC_REST_API_URL = "v1/screst/"

const MAX_RETRIES = 5
const SLEEP_BETWEEN_RETRIES = 5

// In percentage
const consensusThresh = float32(25.0)

// MakeSCRestAPICall makes a rest api call to the sharders.
//   - scAddress is the address of the smart contract
//   - relativePath is the relative path of the api
//   - params is the query parameters
//   - handler is the handler function to handle the response
func MakeSCRestAPICall(scAddress string, relativePath string, params map[string]string, handler SCRestAPIHandler) ([]byte, error) {
	nodeClient, err := GetNode()
	if err != nil {
		return nil, err
	}
	numSharders := len(nodeClient.Sharders().Healthy())
	sharders := nodeClient.Sharders().Healthy()
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
				logger.Log.Error(err)
				return
			}
			q := urlObj.Query()
			for k, v := range params {
				q.Add(k, v)
			}
			urlObj.RawQuery = q.Encode()
			clientObj := &http.Client{Transport: DefaultTransport}
			response, err := clientObj.Get(urlObj.String())
			if err != nil {
				nodeClient.Sharders().Fail(sharder)
				return
			}

			defer response.Body.Close()
			entityBytes, _ := ioutil.ReadAll(response.Body)
			mu.Lock()
			if response.StatusCode > http.StatusBadRequest {
				nodeClient.Sharders().Fail(sharder)
			} else {
				nodeClient.Sharders().Success(sharder)
			}
			responses[response.StatusCode]++
			if responses[response.StatusCode] > maxCount {
				maxCount = responses[response.StatusCode]
			}

			if IsCurrentDominantStatus(response.StatusCode, responses, maxCount) {
				dominant = response.StatusCode
				retObj = entityBytes
			}

			entityResult[sharder] = entityBytes
			nodeClient.Sharders().Success(sharder)
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

// IsCurrentDominantStatus determines whether the current response status is the dominant status among responses.
//
// The dominant status is where the response status is counted the most.
// On tie-breakers, 200 will be selected if included.
//
// Function assumes runningTotalPerStatus can be accessed safely concurrently.
func IsCurrentDominantStatus(respStatus int, currentTotalPerStatus map[int]int, currentMax int) bool {
	// mark status as dominant if
	// - running total for status is the max and response is 200 or
	// - running total for status is the max and count for 200 is lower
	return currentTotalPerStatus[respStatus] == currentMax && (respStatus == 200 || currentTotalPerStatus[200] < currentMax)
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

func GetBalance(clientId string) (*GetBalanceResponse, error) {
	const GET_BALANCE = "/client/get/balance"
	var (
		balance GetBalanceResponse
		err     error
		res     []byte
	)

	if res, err = MakeSCRestAPICall("", GET_BALANCE, map[string]string{
		"client_id": clientId,
	}, nil); err != nil {
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
