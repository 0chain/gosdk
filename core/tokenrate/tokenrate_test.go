package tokenrate

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGetUSD(t *testing.T) {
	for _, tc := range []struct {
		name        string
		expectedErr error
		setup       func() context.Context
	}{
		{
			name:        "context deadline exceeded",
			expectedErr: context.DeadlineExceeded,
			setup:       getRequestContext(10 * time.Microsecond),
		},
		{
			name:        "all success case",
			expectedErr: nil,
			setup:       getRequestContext(100 * time.Second),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setup()
			_, err := GetUSD(ctx, "eth")
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
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
