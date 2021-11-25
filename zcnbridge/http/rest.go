package http

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zcnbridge/errors"
	"github.com/0chain/gosdk/zcncore"
)

const (
	// SCRestAPIPrefix represents base URL path to execute smart contract rest points.
	SCRestAPIPrefix      = "v1/screst/"
	SmartContractAddress = `6dba10422e368813802877a85039d3985d96760ed844092319743fb3a76712d7`
	RestPrefix           = SCRestAPIPrefix + SmartContractAddress
	GetAuthorizersPath   = "/getAuthorizerNodes"
)

// MakeSCRestAPICall calls smart contract with provided address
// and makes retryable request to smart contract resource with provided relative path using params.
func MakeSCRestAPICall(relativePath string, params map[string]string) ([]byte, error) {
	var (
		resMaxCounterBody []byte
		hashMaxCounter    int
		hashCounters      = make(map[string]int)
		sharders          = extractSharders()
		lastErrMsg        string
	)

	for _, sharder := range sharders {
		var (
			client = NewRetryableClient()
			u      = makeURL(params, sharder, relativePath)
		)

		resp, err := client.Get(u.String())
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
	network := zcncore.GetNetwork()
	return util.GetRandom(network.Sharders, len(network.Sharders))
}

// makeURL creates url.URL to make smart contract request to sharder.
func makeURL(params map[string]string, sharder, relativePath string) *url.URL {
	uString := fmt.Sprintf("%v/%v%v", sharder, RestPrefix, relativePath)
	u, _ := url.Parse(uString)
	q := u.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()

	return u
}
