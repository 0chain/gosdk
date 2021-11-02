package zcnbridge_test

import (
	"github.com/0chain/gosdk/zcnbridge"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMain(m *testing.M) {
	// 1. IncreaseAllowance
	// 2. Burn
	// 3. Retrieve all authorizers
	// 4. Send transaction hash to all authorizers
	// 5. Poll for authorizers response
	// 6. Collect tickers from all authorizers
	// 7. Send tickets to ZCNSC to the chain
}

func TestInitTestBridge(t *testing.T) {
	t.Run("Increase Allowance for bridge contract to transfer tokens to token pool", func(t *testing.T) {
		zcnbridge.InitBridge() // TODO: Fill in the configuration

		transaction, err := zcnbridge.IncreaseBurnerAllowance(10000000)

		require.NoError(t, err)
		require.NotNil(t, transaction)
		require.NotEmpty(t, transaction.Hash())
		t.Logf("Transaction hash: %s", transaction.Hash().Hex())

		res := zcnbridge.TransactionStatus(transaction.Hash().Hex())
		require.Equal(t, 1, res)
	})
}

func TestTransactionStatus(t *testing.T) {
	t.Run("Burn WZCN in Ether RPC", func(t *testing.T) {
		zcnbridge.InitBridge() // TODO: Fill in the configuration

		transaction, err := zcnbridge.BurnWZCN(10000000, "123")

		require.NoError(t, err)
		require.NotNil(t, transaction)
		require.NotEmpty(t, transaction.Hash())
		t.Logf("Transaction hash: %s", transaction.Hash().Hex())

		res := zcnbridge.TransactionStatus(transaction.Hash().Hex())
		require.Equal(t, 1, res)
	})
}
