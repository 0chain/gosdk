package tokenrate

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGetUSD(t *testing.T) {
	for _, tc := range []struct {
		name        string
		expectedErr error
		setup       func() context.Context
		symbol      string
	}{
		{
			name:        "context deadline exceeded",
			expectedErr: context.DeadlineExceeded,
			setup:       getRequestContext(10 * time.Millisecond),
			symbol:      "eth",
		},
		{
			name:        "all success case",
			expectedErr: nil,
			setup:       getRequestContext(10 * time.Second),
			symbol:      "eth",
		},
		{
			name:        "error wrong symbol",
			expectedErr: getErrorForWrongSymbol(),
			setup:       getRequestContext(10 * time.Second),
			symbol:      "wrong",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setup()
			val, err := GetUSD(ctx, tc.symbol)
			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Greater(t, val, 0.0)
			}
		})
	}
}

func getRequestContext(d time.Duration) func() context.Context {
	return func() context.Context {
		ctx, cancel := context.WithTimeout(context.TODO(), d)
		go func() {
			<-ctx.Done()
			cancel()
		}()
		return ctx
	}
}

func getErrorForWrongSymbol() error {
	errs := make([]error, 0, 3)
	errs = append(errs, fmt.Errorf("bancor: please configure dlt_id on environment variable [%v] first", "BANCOR_DLTID_WRONG"))
	errs = append(errs, fmt.Errorf("coingecko: please configure coinid on environment variable [%v] first", "COINGECKO_COINID_WRONG"))
	errs = append(errs, fmt.Errorf("coinmarketcap: wrong is not provided on coinmarketcap apis"))
	return fmt.Errorf("%w: %s", ErrNoAvailableQuoteQuery, errs)
}
