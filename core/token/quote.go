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
	}

}

func GetTokenRate(ctx context.Context, currency string) (float64, error) {

	for _, q := range quotes {
		r, err := q.GetRate(ctx, currency)
		if err == nil {
			return r, nil
		}
		log.Println("token: ", err)
	}

	return 0, ErrNoAvailableQuoteQuery
}

type QuoteQuery interface {
	GetRate(ctx context.Context, currency string) (float64, error)
}
