package zcnbridge_test

import (
	"context"
	"testing"

	"github.com/0chain/gosdk/zcnbridge"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// 1. IncreaseAllowance (client)
	// 2. Burn (client)
	// 3. Wait for transaction to complete (client)
	// 3. Retrieve all authorizers (client)
	// 4. Send transaction hash to all authorizers (client)
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

		res := zcnbridge.GetTransactionStatus(transaction.Hash().Hex())
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

		res := zcnbridge.GetTransactionStatus(transaction.Hash().Hex())
		require.Equal(t, 1, res)
	})
}

func TestBurnTicketCollection(t *testing.T) {
	t.Run("Burn WZCN in Ether RPC", func(t *testing.T) {
		zcnbridge.InitBridge() // TODO: Fill in the configuration

		transaction, err := zcnbridge.BurnWZCN(10000000, "123")

		require.NoError(t, err)
		require.NotNil(t, transaction)
		require.NotEmpty(t, transaction.Hash())
		t.Logf("Transaction hash: %s", transaction.Hash().Hex())

		res := zcnbridge.GetTransactionStatus(transaction.Hash().Hex())
		require.Equal(t, 1, res)

		payload, err := zcnbridge.CreateMintPayload(
			context.TODO(),
			transaction.Hash().Hex(),
			wallet.ZCNSCSmartContractAddress,
			wallet.ClientID,
			wallet.ConsensusThresh,
		)

		require.NoError(t, err)
		require.NotNil(t, payload)

		require.NotEmpty(t, payload.ReceivingClientID)
		require.NotEmpty(t, payload.EthereumTxnID)
		require.NotEqual(t, 0, payload.Nonce)
		require.NotEqual(t, 0, payload.Amount)
		require.NotEqual(t, 0, len(payload.Signatures))
	})
}
