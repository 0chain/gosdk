package tokenrate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/zboxcore/mocks"
)

func TestGetUSD(t *testing.T) {
	mockProviderUrl := "coinProvider"
	var mockClient = mocks.HttpClient{}
	resty.CreateClient = func(t *http.Transport, timeout time.Duration) resty.Client {
		return &mockClient
	}

	for _, tc := range []struct {
		name          string
		expectedErr   error
		expectedValue float64
		timeout       time.Duration
		symbol        string
		provider      string
		setup         func(testcaseName, symbol, provider string)
		response      func(testCaseName, mockProviderURL, providerName, symbol string, timeout time.Duration) (float64, error)
	}{
		{
			name:        "ContextDeadlineExceeded",
			expectedErr: context.DeadlineExceeded,
			timeout:     1 * time.Microsecond,
			symbol:      "ZCN",
			provider:    "bancor",
			setup: func(testCaseName, symbol, provider string) {
				setupMockHttpResponse(&mockClient, provider, mockProviderUrl, "TestGetUSD", testCaseName, "GET", symbol, http.StatusOK, getProviderJsonResponse(t, provider))
			},
			response: getBancorResponse(),
		},
		{
			name:          "TestBancorCorrectSymbol",
			expectedErr:   nil,
			expectedValue: 0.118837,
			timeout:       10 * time.Second,
			symbol:        "ZCN",
			provider:      "bancor",
			setup: func(testCaseName, symbol, provider string) {
				setupMockHttpResponse(&mockClient, provider, mockProviderUrl, "TestGetUSD", testCaseName, "GET", symbol, http.StatusOK, getProviderJsonResponse(t, provider))
			},
			response: getBancorResponse(),
		},
		{
			name:          "TestCoinmarketcapCorrectSymbol",
			expectedErr:   nil,
			expectedValue: 0.2209498683307504,
			timeout:       10 * time.Second,
			symbol:        "ZCN",
			provider:      "coinmarketcap",
			setup: func(testCaseName, symbol, provider string) {
				setupMockHttpResponse(&mockClient, provider, mockProviderUrl, "TestGetUSD", testCaseName, "GET", symbol, http.StatusOK, getProviderJsonResponse(t, provider))
			},
			response: getCoinmarketcapResponse(),
		},
		{
			name:        "TestCoinmarketcapWrongSymbol",
			expectedErr: fmt.Errorf("429, failed to get coin data from provider coinmarketcap for symbol \"wrong\""),
			timeout:     10 * time.Second,
			symbol:      "wrong",
			provider:    "coinmarketcap",
			setup: func(testCaseName, symbol, provider string) {
				setupMockHttpResponse(&mockClient, provider, mockProviderUrl, "TestGetUSD", testCaseName, "GET", symbol, http.StatusTooManyRequests, getProviderJsonResponse(t, provider))
			},
			response: getBancorResponse(),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(tc.name, tc.symbol, tc.provider)
			}
			var value float64
			var err error
			if tc.response != nil {
				value, err = tc.response(tc.name, mockProviderUrl, tc.provider, tc.symbol, tc.timeout)
			}

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, value, tc.expectedValue)
			}
		})
	}
}

func setupMockHttpResponse(
	mockClient *mocks.HttpClient, provider, mockProviderUrl, funcName, testCaseName, httpMethod, symbol string,
	statusCode int, body []byte) {
	url := funcName + testCaseName + mockProviderUrl + provider
	mockClient.On("Do", mock.MatchedBy(func(req *http.Request) bool {
		return req.Method == httpMethod &&
			strings.Contains(req.URL.String(), url) && (req.URL.Query().Get("symbol") == symbol)
	})).Return(
		&http.Response{
			StatusCode: statusCode,
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}, nil).Once()
}

func getBancorResponse() func(testCaseName, mockProviderURL, providerName, symbol string, timeout time.Duration) (float64, error) {
	return func(testCaseName, mockProviderURL, providerName, symbol string, timeout time.Duration) (float64, error) {
		var br bancorResponse
		ctx, cancel := context.WithTimeout(context.TODO(), timeout)
		go func() {
			<-ctx.Done()
			cancel()
		}()

		reqUrl := "TestGetUSD" + testCaseName + mockProviderURL + providerName + "?symbol=" + symbol
		r := resty.New(resty.WithRetry(1))
		r.DoGet(ctx, reqUrl).Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("%v, failed to get coin data from provider %v for symbol \"%v\"", resp.StatusCode, providerName, symbol)
			}
			err = json.Unmarshal(respBody, &br)
			log.Printf("==Response: %v", br)
			if err != nil {
				return err
			}
			br.Raw = string(respBody)
			return nil
		})
		errs := r.Wait()
		if len(errs) != 0 {
			return 0.0, errs[0]
		}
		return br.Data.Rate["usd"].Value, nil
	}
}

func getCoinmarketcapResponse() func(testCaseName, mockProviderURL, providerName, symbol string, timeout time.Duration) (float64, error) {
	return func(testCaseName, mockProviderURL, providerName, symbol string, timeout time.Duration) (float64, error) {
		var cr coinmarketcapResponse
		ctx, cancel := context.WithTimeout(context.TODO(), timeout)
		go func() {
			<-ctx.Done()
			cancel()
		}()

		reqUrl := "TestGetUSD" + testCaseName + mockProviderURL + providerName + "?symbol=" + symbol
		r := resty.New(resty.WithRetry(1))
		r.DoGet(ctx, reqUrl).Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to get coin data from provider %v for symbol \"%v\"", providerName, symbol)
			}
			err = json.Unmarshal(respBody, &cr)
			if err != nil {
				return err
			}
			cr.Raw = string(respBody)
			return nil
		})
		errs := r.Wait()
		if len(errs) != 0 {
			return 0.0, errs[0]
		}
		if len(cr.Data[strings.ToUpper(symbol)]) == 0 {
			return 0.0, fmt.Errorf("coinmarketcap: symbol \"%v\" is not provided on coinmarketcap apis", symbol)
		}
		val := cr.Data[strings.ToUpper(symbol)][0].Quote["USD"].Price
		return val, nil
	}
}

func getProviderJsonResponse(t *testing.T, provider string) []byte {
	data, err := ioutil.ReadFile("mockresponses/" + provider + ".json")
	if err != nil {
		t.Fatal(err)
	}
	return data
}
