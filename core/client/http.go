package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/util"
	"github.com/shopspring/decimal"
)

// SCRestAPIHandler is a function type to handle the response from the SC Rest API
//
//	`response` - the response from the SC Rest API
//	`numSharders` - the number of sharders that responded
//	`err` - the error if any
type SCRestAPIHandler func(response map[string][]byte, numSharders int, err error)

func MakeSCRestAPICallToSharder(scAddress string, relativePath string, params map[string]string, restApiUrls ...string) ([]byte, error) {
	const (
		consensusThresh = float32(25.0)
		ScRestApiUrl    = "v1/screst/"
	)

	restApiUrl := ScRestApiUrl
	if len(restApiUrls) > 0 {
		restApiUrl = restApiUrls[0]
	}

	sharders := nodeClient.sharders.HealthyByLFB()
	responses := make(map[int]int)
	entityResult := make(map[string][]byte)

	var (
		retObj   []byte
		maxCount int
		dominant = 200
		wg       sync.WaitGroup
		mu       sync.Mutex // Mutex to protect shared resources
	)

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
				logger.GetLogger().Error("Error parsing URL: ", err.Error())
				return
			}
			q := urlObj.Query()
			for k, v := range params {
				q.Add(k, v)
			}
			urlObj.RawQuery = q.Encode()

			req, err := util.NewHTTPGetRequest(urlObj.String())
			if err != nil {
				logger.GetLogger().Error("Error creating request: ", err.Error())
				return
			}

			response, err := req.Get()
			if err != nil {
				nodeClient.sharders.Fail(sharder)
				logger.GetLogger().Error("Error getting response: ", err.Error())
				return
			}

			mu.Lock() // Lock before updating shared maps
			defer mu.Unlock()

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
				retObj = []byte(response.Body)
			}

			entityResult[sharder] = []byte(response.Body)
			nodeClient.sharders.Success(sharder)
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

// GetNonce returns the highest nonce for clientID across all LFB-healthy sharders.
// Using the maximum protects against stale nonces from lagging or stuck sharders.
func GetNonce(clientID string) (int64, error) {
	if clientID == "" {
		clientID = Id()
	}

	sharders := nodeClient.sharders.HealthyByLFB()
	if len(sharders) == 0 {
		return 0, errors.New("no_sharders", "no healthy sharders available")
	}

	type result struct {
		nonce int64
		err   error
	}

	results := make(chan result, len(sharders))

	for _, sharder := range sharders {
		go func(s string) {
			urlString := fmt.Sprintf("%s/v1/client/get/balance?client_id=%s", s, clientID)
			req, err := util.NewHTTPGetRequest(urlString)
			if err != nil {
				results <- result{err: err}
				return
			}

			resp, err := req.Get()
			if err != nil {
				results <- result{err: err}
				return
			}

			if resp.StatusCode != http.StatusOK {
				results <- result{err: fmt.Errorf("status %d: %s", resp.StatusCode, resp.Body)}
				return
			}

			var bal GetBalanceResponse
			if err := json.Unmarshal([]byte(resp.Body), &bal); err != nil {
				results <- result{err: err}
				return
			}
			results <- result{nonce: bal.Nonce}
		}(sharder)
	}

	maxNonce := int64(-1)
	var lastErr error
	for i := 0; i < len(sharders); i++ {
		r := <-results
		if r.err != nil {
			lastErr = r.err
			continue
		}
		if r.nonce > maxNonce {
			maxNonce = r.nonce
		}
	}

	if maxNonce < 0 {
		if lastErr != nil {
			return 0, lastErr
		}
		return 0, errors.New("no_nonce", "could not get nonce from any sharder")
	}
	return maxNonce, nil
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
		clientID = Id()
	}

	if res, err = MakeSCRestAPICallToSharder("", GetBalance, map[string]string{
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

// MakeSCRestAPICall is a backward-compatible alias for MakeSCRestAPICallToSharder.
// Older callers (e.g. blobber core/transaction/http.go) reference this name directly.
var MakeSCRestAPICall = MakeSCRestAPICallToSharder
