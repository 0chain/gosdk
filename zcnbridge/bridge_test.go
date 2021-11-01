package zcnbridge_test

import (
	"github.com/0chain/gosdk/zcnbridge"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMain(m *testing.M) {
}

func TestInitTestBridge(t *testing.T) {
	t.Run("Burn WZCN in Ether RPC", func(t *testing.T) {
		zcnbridge.InitBridge() // TODO: Fill in

		transaction, err := zcnbridge.IncreaseBurnerAllowance(10000000)

		require.NoError(t, err)
		require.NotNil(t, transaction)
		require.NotEmpty(t, transaction.Hash())
		transaction.Hash().Hex()
	})
}
