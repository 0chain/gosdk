package tokenrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/0chain/gosdk/core/resty"
)

type coinmarketcapQuoteQuery struct {
	APIKey string
}

// js call is unsupported for coinmarketcap api due to core issue
// https://coinmarketcap.com/api/documentation/v1/#section/Quick-Start-Guide
// Note: Making HTTP requests on the client side with Javascript is currently prohibited through CORS configuration. This is to protect your API Key which should not be visible to users of your application so your API Key is not stolen. Secure your API Key by routing calls through your own backend service.
func createCoinmarketcapQuoteQuery() quoteQuery {

	coinmarketcapAPIKEY, ok := os.LookupEnv("COINMARKETCAP_API_KEY")
	if !ok {
		coinmarketcapAPIKEY = "7e386213-56ef-4a7e-af17-806496c20d3b"
	}

	return &coinmarketcapQuoteQuery{
		APIKey: coinmarketcapAPIKEY,
	}
}

func (qq *coinmarketcapQuoteQuery) getUSD(ctx context.Context, symbol string) (float64, error) {

	var result coinmarketcapResponse

	r := resty.New(resty.WithHeader(map[string]string{
		"X-CMC_PRO_API_KEY": qq.APIKey,
	}))

	s := strings.ToUpper(symbol)

	r.DoGet(ctx, "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest?symbol="+s).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusOK {
				return errors.New("coinmarketcap: " + strconv.Itoa(resp.StatusCode) + resp.Status)
			}

			err = json.Unmarshal(respBody, &result)
			if err != nil {
				return err
			}

			result.Raw = string(respBody)

			return nil

		})

	errs := r.Wait()
	if len(errs) > 0 {
		return 0, errs[0]
	}

	zcn, ok := result.Data[s]

	if !ok || len(zcn) == 0 {
		return 0, errors.New("coinmarketcap: " + symbol + " is not provided on coinmarketcap apis")
	}

	rate, ok := zcn[0].Quote["USD"]
	if ok {
		if rate.Price > 0 {
			return rate.Price, nil
		}

		return 0, fmt.Errorf("coinmarketcap: invalid response %s", result.Raw)
	}

	return 0, errors.New("coinmarketcap: " + symbol + " to USD quote is not provided on coinmarketcap apis")
}

//	{
//		"status": {
//				"timestamp": "2022-06-03T02:18:34.093Z",
//				"error_code": 0,
//				"error_message": null,
//				"elapsed": 50,
//				"credit_count": 1,
//				"notice": null
//		},
//		"data": {
//				"ZCN": [
//						{
//								"id": 2882,
//								"name": "0Chain",
//								"symbol": "ZCN",
//								"slug": "0chain",
//								"num_market_pairs": 8,
//								"date_added": "2018-07-02T00:00:00.000Z",
//								"tags": [
//										{
//												"slug": "platform",
//												"name": "Platform",
//												"category": "PROPERTY"
//										},
//										{
//												"slug": "ai-big-data",
//												"name": "AI & Big Data",
//												"category": "PROPERTY"
//										},
//										{
//												"slug": "distributed-computing",
//												"name": "Distributed Computing",
//												"category": "PROPERTY"
//										},
//										{
//												"slug": "filesharing",
//												"name": "Filesharing",
//												"category": "PROPERTY"
//										},
//										{
//												"slug": "iot",
//												"name": "IoT",
//												"category": "PROPERTY"
//										},
//										{
//												"slug": "storage",
//												"name": "Storage",
//												"category": "PROPERTY"
//										}
//								],
//								"max_supply": 400000000,
//								"circulating_supply": 48400982,
//								"total_supply": 200000000,
//								"platform": {
//										"id": 1027,
//										"name": "Ethereum",
//										"symbol": "ETH",
//										"slug": "ethereum",
//										"token_address": "0xb9ef770b6a5e12e45983c5d80545258aa38f3b78"
//								},
//								"is_active": 1,
//								"cmc_rank": 782,
//								"is_fiat": 0,
//								"self_reported_circulating_supply": 115000000,
//								"self_reported_market_cap": 25409234.858036295,
//								"last_updated": "2022-06-03T02:17:00.000Z",
//								"quote": {
//										"USD": {
//												"price": 0.2209498683307504,
//												"volume_24h": 28807.79174117,
//												"volume_change_24h": -78.341,
//												"percent_change_1h": 0.09600341,
//												"percent_change_24h": 0.1834049,
//												"percent_change_7d": 24.08736297,
//												"percent_change_30d": -43.56084388,
//												"percent_change_60d": -63.69787917,
//												"percent_change_90d": -27.17695342,
//												"market_cap": 10694190.59997902,
//												"market_cap_dominance": 0.0008,
//												"fully_diluted_market_cap": 88379947.33,
//												"last_updated": "2022-06-03T02:17:00.000Z"
//										}
//								}
//						}
//				]
//		}
//	}
type coinmarketcapResponse struct {
	Data map[string][]coinmarketcapCurrency `json:"data"`
	Raw  string                             `json:"-"`
}

type coinmarketcapCurrency struct {
	Quote map[string]coinmarketcapQuote `json:"quote"`
}

type coinmarketcapQuote struct {
	Price float64 `json:"price"`
}
