package tokenrate

import (
	"context"
	"errors"
	"fmt"
	"sync"
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
	mu := &sync.RWMutex{}
	errs := make([]error, 0, len(quotes))
	finalErrCh := make(chan error)
	resultCh := make(chan float64)

	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	for _, q := range quotes {
		go func(q quoteQuery) {
			errCh := make(chan error)
			go q.getUSD(ctx, symbol, errCh, resultCh)
			select {
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					finalErrCh <- ctx.Err()
				}
				cancel()
				return
			case e := <-errCh:
				mu.Lock()
				errs = append(errs, e)
				mu.Unlock()
				if len(errs) >= len(quotes) {
					finalErrCh <- fmt.Errorf("%w: %s", ErrNoAvailableQuoteQuery, errs)
					cancel()
					return
				}
			}
		}(q)
	}

	select {
	case e := <-finalErrCh:
		return 0, e
	case r := <-resultCh:
		return r, nil
	}
}

type quoteQuery interface {
	getUSD(ctx context.Context, symbol string, errCh chan error, resultCh chan float64)
}
