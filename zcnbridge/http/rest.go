package http

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"go.uber.org/zap"

	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zcnbridge/errors"
	"github.com/0chain/gosdk/zcncore"
)

const (
	// SCRestAPIPrefix represents base URL path to execute smart contract rest points.
	SCRestAPIPrefix        = "v1/screst/"
	SmartContractAddress   = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7`
	RestPrefix             = SCRestAPIPrefix + SmartContractAddress
	PathGetAuthorizerNodes = "/getAuthorizerNodes"
	PathGetGlobalConfig    = "/getGlobalConfig"
	PathGetAuthorizer      = "/getAuthorizer"
)

type Params map[string]string

var Logger logger.Logger
var defaultLogLevel = logger.DEBUG

func init() {
	Logger.Init(defaultLogLevel, "0chain-zcnbridge-sdk")
}

// MakeSCRestAPICall calls smart contract with provided address
// and makes retryable request to smart contract resource with provided relative path using params.
func MakeSCRestAPICall(opCode int, relativePath string, params Params, cb zcncore.GetInfoCallback) {
	var (
		resMaxCounterBody []byte
		hashMaxCounter    int
		msg               string
		hashCounters      = make(map[string]int)
		sharders          = extractSharders()
	)

	type queryResult struct {
		hash string
		body []byte
	}

	results := make(chan *queryResult, len(sharders))
	defer close(results)

	var client = NewRetryableClient()

	wg := &sync.WaitGroup{}
	for _, sharder := range sharders {
		wg.Add(1)
		go func(sharderUrl string) {
			defer wg.Done()

			var u = makeURL(params, sharderUrl, relativePath)
			Logger.Info(fmt.Sprintf("Query %s", u.String()))

			resp, err := client.Get(u.String())
			//goland:noinspection ALL
			defer resp.Body.Close()

			if err != nil {
				Logger.Error("MakeSCRestAPICall - failed to get response from", zap.String("URL", sharderUrl), zap.Any("error", err))
				return
			}

			if resp.StatusCode != http.StatusOK {
				Logger.Error("MakeSCRestAPICall - error getting response from", zap.String("URL", sharderUrl), zap.Any("error", err))
				return
			}

			hash, body, err := hashAndBytesOfReader(resp.Body)
			if err != nil {
				Logger.Error("MakeSCRestAPICall - error while reading response body", zap.String("URL", sharderUrl), zap.Any("error", err))
				return
			}

			results <- &queryResult{hash: hash, body: body}
		}(sharder)
	}

	wg.Wait()
	close(results)

	select {
	case result := <-results:
		hashCounters[result.hash]++
		if hashCounters[result.hash] > hashMaxCounter {
			hashMaxCounter = hashCounters[result.hash]
			resMaxCounterBody = result.body
		}
	default:
	}

	if hashMaxCounter == 0 {
		err := errors.New("request_sharders", "no valid responses, last err: "+msg)
		cb.OnInfoAvailable(opCode, zcncore.StatusError, "", err.Error())
		Logger.Error(err)
		return
	}

	cb.OnInfoAvailable(opCode, zcncore.StatusSuccess, string(resMaxCounterBody), "")
}

// hashAndBytesOfReader computes hash of readers data and returns hash encoded to hex and bytes of reader data.
// If error occurs while reading data from reader, it returns non nil error.
func hashAndBytesOfReader(r io.Reader) (string, []byte, error) {
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
	network := zcncore.GetNetwork()
	return util.GetRandom(network.Sharders, len(network.Sharders))
}

// makeURL creates url.URL to make smart contract request to sharder.
func makeURL(params Params, baseURL, relativePath string) *url.URL {
	uString := fmt.Sprintf("%v/%v%v", baseURL, RestPrefix, relativePath)
	u, _ := url.Parse(uString)
	q := u.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()

	return u
}
