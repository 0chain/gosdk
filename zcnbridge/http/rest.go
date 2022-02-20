package http

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

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

	result := make(chan *http.Response)
	defer close(result)

	var client = NewRetryableClient()

	for _, sharder := range sharders {
		go func(sharderUrl string) {
			var u = makeURL(params, sharderUrl, relativePath)
			Logger.Info(fmt.Sprintf("Query %s", u.String()))
			resp, err := client.Get(u.String())
			if err != nil {
				msg := fmt.Sprintf("%s: error while requesting sharders: %v", sharderUrl, err)
				Logger.Error(msg)
				return
			}
			result <- resp
		}(sharder)
	}

	for range sharders {
		resp := <-result

		hash, resBody, err := hashAndBytesOfReader(resp.Body)
		_ = resp.Body.Close()

		if err != nil {
			msg = fmt.Sprintf("error while reading response body: %v", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			msg = fmt.Sprintf("response status is not OK; response body: %s", string(resBody))
			continue
		}

		hashCounters[hash]++
		if hashCounters[hash] > hashMaxCounter {
			hashMaxCounter = hashCounters[hash]
			resMaxCounterBody = resBody
		}
	}

	if hashMaxCounter == 0 {
		err := errors.New("request_sharders", "no valid responses, last err: "+msg)
		cb.OnInfoAvailable(opCode, zcncore.StatusError, "", err.Error())
		Logger.Error(err)
		return
	}

	cb.OnInfoAvailable(opCode, zcncore.StatusSuccess, string(resMaxCounterBody), "")

	return
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
