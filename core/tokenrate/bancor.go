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

type bancorQuoteQuery struct {
}

func (qq *bancorQuoteQuery) getUSD(ctx context.Context, symbol string) (float64, error) {

	var result bancorResponse

	s := strings.ToLower(symbol)
	var dltId string
	//
	switch s {
	case "zcn":
		dltId = "0xb9EF770B6A5e12E45983C5D80545258aA38F3B78"
	case "eth":
		dltId = "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"
	default:
		evnName := "BANCOR_DLTID_" + strings.ToUpper(symbol)
		id, ok := os.LookupEnv(evnName)
		if !ok {
			return 0, errors.New("bancor: please configure dlt_id on environment variable [" + evnName + "] first")
		}
		dltId = id
	}

	r := resty.New()
	r.DoGet(ctx, "https://api-v3.bancor.network/tokens?dlt_id="+dltId).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {

			if err != nil {
				return err
			}

			if resp.StatusCode != http.StatusOK {
				return errors.New("bancor: " + strconv.Itoa(resp.StatusCode) + resp.Status)
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

	rate, ok := result.Data.Rate24hAgo["usd"]

	if ok {

		if rate.Value > 0 {
			return rate.Value, nil
		}

		//rate24ago is invalid, try get current rate
		rate, ok = result.Data.Rate["usd"]
		if ok && rate.Value > 0 {
			return rate.Value, nil
		}
	}

	return 0, fmt.Errorf("bancor: %s price is not provided on bancor apis", symbol)
}

// {
// 	"data": {
// 			"dltId": "0xb9EF770B6A5e12E45983C5D80545258aA38F3B78",
// 			"symbol": "ZCN",
// 			"decimals": 10,
// 			"rate": {
// 					"bnt": "0.271257342312491431",
// 					"usd": "0.118837",
// 					"eur": "0.121062",
// 					"eth": "0.000089243665620809"
// 			},
// 			"rate24hAgo": {
// 					"bnt": "0.273260935543748855",
// 					"usd": "0.120972",
// 					"eur": "0.126301",
// 					"eth": "0.000094001761827049"
// 			}
// 	},
// 	"timestamp": {
// 			"ethereum": {
// 					"block": 15644407,
// 					"timestamp": 1664519843
// 			}
// 	}
// }

type bancorResponse struct {
	Data bancorMarketData `json:"data"`
	Raw  string           `json:"-"`
}

type bancorMarketData struct {
	Rate       map[string]Float64 `json:"rate"`
	Rate24hAgo map[string]Float64 `json:"rate24hAgo"`
}

type Float64 struct {
	Value float64
}

func (s *Float64) UnmarshalJSON(data []byte) error {

	if data == nil {
		s.Value = 0
		return nil
	}

	js := strings.Trim(string(data), "\"")

	v, err := strconv.ParseFloat(js, 32)
	if err != nil {
		return err
	}

	s.Value = v
	return nil

}
