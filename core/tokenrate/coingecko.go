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
		envName := "COINGECKO_COINID_" + strings.ToUpper(symbol)
		id, ok := os.LookupEnv(envName)
		if !ok {
			return 0, errors.New("coingecko: please configure coinid on environment variable [" + envName + "] first")
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
				return errors.New("coingecko: " + strconv.Itoa(resp.StatusCode) + resp.Status)
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

	var rate float64

	h, ok := result.MarketData.High24h["usd"]
	if ok {
		l, ok := result.MarketData.Low24h["usd"]
		if ok {
			rate = (h + l) / 2
			if rate > 0 {
				return rate, nil
			}
		}
	}

	rate, ok = result.MarketData.CurrentPrice["usd"]

	if ok {
		if rate > 0 {
			return rate, nil
		}

		return 0, fmt.Errorf("coingecko: invalid response %s", result.Raw)
	}

	return 0, fmt.Errorf("coingecko: %s price is not provided on coingecko apis", symbol)
}

type coingeckoResponse struct {
	MarketData coingeckoMarketData `json:"market_data"`
	Raw        string              `json:"-"`
}

type coingeckoMarketData struct {
	CurrentPrice map[string]float64 `json:"current_price"`
	High24h      map[string]float64 `json:"high_24h"`
	Low24h       map[string]float64 `json:"low_24h"`
}
