package zcncore

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTokensConversion(t *testing.T) {
	t.Run("Tokens to Eth", func(t *testing.T) {
		ethTokens := TokensToEth(4337488392000000000)
		require.Equal(t, 4.337488392, ethTokens)
	})

	t.Run("Eth to tokens", func(t *testing.T) {
		ethTokens := EthToTokens(4.337488392)
		require.Equal(t, int64(4337488392000000000), ethTokens)
	})
}