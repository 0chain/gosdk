package client

import (
	"encoding/json"
	"fmt"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/util"
	"github.com/shopspring/decimal"
	"net/http"
	"net/url"
	"sync"
)

// SCRestAPIHandler is a function type to handle the response from the SC Rest API
//
//	`response` - the response from the SC Rest API
//	`numSharders` - the number of sharders that responded
//	`err` - the error if any
type SCRestAPIHandler func(response map[string][]byte, numSharders int, err error)

// MakeSCRestAPICall makes a rest api call to the sharders.
//   - scAddress is the address of the smart contract
//   - relativePath is the relative path of the api
//   - params is the query parameters
//   - handler is the handler function to handle the response
func MakeSCRestAPICall(scAddress string, relativePath string, params map[string]string, restApiUrls ...string) ([]byte, error) {
	const (
		consensusThresh = float32(25.0)
		ScRestApiUrl    = "v1/screst/"
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
				fmt.Println(err.Error())
				return
			}
			q := urlObj.Query()
			for k, v := range params {
				q.Add(k, v)
			}
			urlObj.RawQuery = q.Encode()

			req, err := util.NewHTTPGetRequest(urlObj.String())
			if err != nil {
				fmt.Println("1Error creating request", err.Error())
				return
			}
			response, err := req.Get()
			if err != nil {
				fmt.Println("2Error getting response", err.Error())
				return
			}

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
