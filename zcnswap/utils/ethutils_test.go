package utils

import (
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestEthUtils(t *testing.T) {
	t.Run("Correct percentage", func(t *testing.T) {
		value := new(big.Int).SetInt64(int64(2100))
		res := AddPercents(value.Uint64(), 10)
		require.Equal(t, 2310, int(res.Int64()))
	})
}
