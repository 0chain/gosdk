package tokenrate

import (
	"context"
	"errors"
	"fmt"
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

	errs := make([]error, 0, len(quotes))
	for _, q := range quotes {
		r, err := q.getUSD(ctx, symbol)
		if err == nil {
			return r, nil
		}

		errs = append(errs, err)
	}

	return 0, fmt.Errorf("%w: %s", ErrNoAvailableQuoteQuery, errs)
}

type quoteQuery interface {
	getUSD(ctx context.Context, symbol string) (float64, error)
}
