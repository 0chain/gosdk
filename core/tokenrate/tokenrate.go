package tokenrate

import (
	"context"
	"errors"
)

var ErrNoAvailableQuoteQuery = errors.New("token: no available quote query service")
var quotes []quoteQuery

func init() {

	//priority: uniswap > bancor > coingecko > coinmarketcap
	quotes = []quoteQuery{
		&coingeckoQuoteQuery{},
		&bancorQuoteQuery{},
		&uniswapQuoteQuery{},
		createCoinmarketcapQuoteQuery(),
		//more query services
	}

}

func GetUSD(ctx context.Context, symbol string) (float64, error) {
	var err error

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	for _, q := range quotes {
		val, err := q.getUSD(ctx, symbol)

		if err != nil {
			continue
		}

		if val > 0 {
			return val, nil
		}
	}

	// All conversion APIs failed
	if err != nil {
		return 0, err
	}

	return 0, ErrNoAvailableQuoteQuery
}

type quoteQuery interface {
	getUSD(ctx context.Context, symbol string) (float64, error)
}
