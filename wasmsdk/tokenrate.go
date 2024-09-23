package main

import (
	"context"

	"github.com/0chain/gosdk/core/tokenrate"
)

// getUSDRate gets the USD rate for the given crypto symbol
func getUSDRate(symbol string) (float64, error) { //nolint:unused
	return tokenrate.GetUSD(context.TODO(), symbol)
}
