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
	t.Run("Burn WZCN in Ether RPC", func(t *testing.T) {
		zcnbridge.InitBridge() // TODO: Fill in

		transaction, err := zcnbridge.IncreaseBurnerAllowance(10000000)

		require.NoError(t, err)
		require.NotNil(t, transaction)
		require.NotEmpty(t, transaction.Hash())
		t.Logf("Transaction hash: %s", transaction.Hash().Hex())
	})
}

func TestTransactionStatus(t *testing.T) {

}
