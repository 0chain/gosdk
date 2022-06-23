package tokenrate

import (
	"context"
	"errors"
)

var ErrNoAvailableQuoteQuery = errors.New("token: no available quote query service")
var quotes []quoteQuery

func init() {

	quotes = []quoteQuery{
		&bancorQuoteQuery{},
		&coingeckoQuoteQuery{},
		createCoinmarketcapQuoteQuery(),
		//more query services
	}

}

func GetUSD(ctx context.Context, symbol string) (float64, error) {

	for _, q := range quotes {
		r, err := q.getUSD(ctx, symbol)
		if err == nil {
			return r, nil
		}
	}

	return 0, ErrNoAvailableQuoteQuery
}

type quoteQuery interface {
	getUSD(ctx context.Context, symbol string) (float64, error)
}
