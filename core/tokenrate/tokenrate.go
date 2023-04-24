package tokenrate

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

var ErrNoAvailableQuoteQuery = errors.New("token: no available quote query service")
var quotes []quoteQuery

func init() {

	//uniswap -> bancor -> coingecko -> coinmarketcap
	quotes = []quoteQuery{
		&uniswapQuoteQuery{},
		&bancorQuoteQuery{},
		&coingeckoQuoteQuery{},
		createCoinmarketcapQuoteQuery(),
		//more query services
	}

}

func GetUSD(ctx context.Context, symbol string) (float64, error) {
	var mu sync.Mutex
	done := false

	errs := make([]error, 0, len(quotes))
	errCh := make(chan error, len(quotes))
	successCh := make(chan float64)

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	for _, q := range quotes {
		go func(q quoteQuery) {
			val, err := q.getUSD(ctx, symbol)
			if err != nil {
				errCh <- err
			} else {
				mu.Lock()
				if !done {
					// we don't want to send result again if it has already been sent
					successCh <- val
					done = true
				}
				mu.Unlock()
			}
		}(q)
	}

	for {
		select {
		case e := <-errCh:
			errs = append(errs, e)
			if len(errs) >= len(quotes) {
				if errs[0] == context.DeadlineExceeded {
					return 0, context.DeadlineExceeded
				}
				sort.Slice(errs, func(i, j int) bool {
					return errs[i].Error() < errs[j].Error()
				})
				return 0, fmt.Errorf("%w: %s", ErrNoAvailableQuoteQuery, errs)
			}
		case r := <-successCh:
			return r, nil
		}
	}
}

type quoteQuery interface {
	getUSD(ctx context.Context, symbol string) (float64, error)
}
