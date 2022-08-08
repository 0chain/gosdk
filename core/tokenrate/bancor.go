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
		if rate == 0 {
			rate, ok = result.Data.Rate["usd"]
			if ok {
				if rate == 0 {
					return 0, fmt.Errorf("bancor: invalid response %s", result.Raw)
				}
			}
		}

		return rate, nil
	}

	return 0, fmt.Errorf("bancor: %s price is not provided on bancor apis", symbol)
}

type bancorResponse struct {
	Data bancorMarketData `json:"data"`
	Raw  string           `json:"-"`
}

type bancorMarketData struct {
	Rate       map[string]float64 `json:"rate"`
	Rate24hAgo map[string]float64 `json:"rate24hAgo"`
}
