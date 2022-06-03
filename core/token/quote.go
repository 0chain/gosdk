package token

import (
	"context"
	"errors"
	"log"
)

var ErrNoAvailableQuoteQuery = errors.New("token: no available quote query service")
var quotes []QuoteQuery

func init() {

	quotes = []QuoteQuery{
		&coingeckoQuoteQuery{},
		createCoinmarketcapQuoteQuery(),
		//more query services
	}

}

func GetTokenUSDRate(ctx context.Context, symbol string) (float64, error) {

	for _, q := range quotes {
		r, err := q.GetUSDRate(ctx, symbol)
		if err == nil {
			return r, nil
		}
		log.Println("token: ", err)
	}

	return 0, ErrNoAvailableQuoteQuery
}

type QuoteQuery interface {
	GetUSDRate(ctx context.Context, symbol string) (float64, error)
}
