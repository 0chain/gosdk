package token

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/0chain/gosdk/core/resty"
)

type coingeckoQuoteQuery struct {
}

func (qq *coingeckoQuoteQuery) GetRate(ctx context.Context, currency string) (float64, error) {

	var result coingeckoResponse

	r := resty.New()
	r.DoGet(ctx, "https://api.coingecko.com/api/v3/coins/0chain?localization=false").
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

	symbol := strings.ToLower(currency)
	rate, ok := result.MarketData.CurrentPrice[symbol]

	if ok {
		return rate, nil
	}

	return 0, errors.New("token: " + symbol + " price is not provided on coingecko apis")
}

type coingeckoResponse struct {
	MarketData coingeckoMarketData `json:"market_data"`
}

type coingeckoMarketData struct {
	CurrentPrice map[string]float64 `json:"current_price"`
}
