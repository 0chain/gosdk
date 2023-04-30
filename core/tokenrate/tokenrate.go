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
		&uniswapQuoteQuery{},
		&bancorQuoteQuery{},
		&coingeckoQuoteQuery{},
		createCoinmarketcapQuoteQuery(),
		//more query services
	}

}

func GetUSD(ctx context.Context, symbol string) (float64, error) {

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	for _, q := range quotes {

		val, err := q.getUSD(ctx, symbol)

		if err != nil {
			return 0, err
		}

		if val > 0 {
			return val, nil
		}

	}

	return 0, ErrNoAvailableQuoteQuery
}

type quoteQuery interface {
	getUSD(ctx context.Context, symbol string) (float64, error)
}
