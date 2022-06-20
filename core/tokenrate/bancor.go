package tokenrate

import (
	"context"
	"encoding/json"
	"errors"
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
		id, ok := os.LookupEnv("BANCOR_DLTID_" + strings.ToUpper(symbol))
		if !ok {
			return 0, errors.New("token: please configurate dlt_id for " + s + " first")
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
				return errors.New("token: " + strconv.Itoa(resp.StatusCode) + resp.Status)
			}

			err = json.Unmarshal(respBody, &result)
			if err != nil {
				return err
			}

			return nil

		})

	r.Wait()

	rate, ok := result.Data.Rate["usd"]

	if ok {
		return rate, nil
	}

	return 0, errors.New("token: " + symbol + " price is not provided on bancor apis")
}

type bancorResponse struct {
	Data bancorMarketData `json:"data"`
}

type bancorMarketData struct {
	Rate map[string]float64 `json:"rate"`
}
