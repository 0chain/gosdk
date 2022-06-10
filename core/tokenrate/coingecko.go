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

type coingeckoQuoteQuery struct {
}

func (qq *coingeckoQuoteQuery) getUSD(ctx context.Context, symbol string) (float64, error) {

	var result coingeckoResponse

	s := strings.ToLower(symbol)
	var coinID string
	//
	switch s {
	case "zcn":
		coinID = "0chain"
	case "eth":
		coinID = "ethereum"
	default:
		id, ok := os.LookupEnv("COINGECKO_COINID_" + strings.ToUpper(symbol))
		if !ok {
			return 0, errors.New("token: please configurate coinid for " + s + " first")
		}
		coinID = id

	}

	r := resty.New()
	r.DoGet(ctx, "https://api.coingecko.com/api/v3/coins/"+coinID+"?localization=false").
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

	rate, ok := result.MarketData.CurrentPrice["usd"]

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
