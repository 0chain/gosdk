package zcnbridge

import (
	"context"
	"encoding/json"
	"log"
	"math/big"
	"os"
	"strconv"
	"testing"

	sdkcommon "github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zcnbridge/ethereum"
	binding "github.com/0chain/gosdk/zcnbridge/ethereum/bridge"
	"github.com/0chain/gosdk/zcnbridge/ethereum/erc20"
	bridgemocks "github.com/0chain/gosdk/zcnbridge/mocks"
	"github.com/0chain/gosdk/zcnbridge/transaction"
	transactionmocks "github.com/0chain/gosdk/zcnbridge/transaction/mocks"
	"github.com/0chain/gosdk/zcnbridge/zcnsc"
	"github.com/0chain/gosdk/zcncore"
	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	ETHEREUM_MNEMONIC = "symbol alley celery diesel donate moral almost opinion achieve since diamond page"

	ETHEREUM_ADDRESS = "0xD8c9156e782C68EE671C09b6b92de76C97948432"
	PASSWORD         = "\"02289b9\""

	BRIDGE_ADDRESS     = "0x7bbbEa24ac1751317D7669f05558632c4A9113D7"
	TOKEN_ADDRESS      = "0x2ec8F26ccC678c9faF0Df20208aEE3AF776160CD"
	AUTHORIZER_ADDRESS = "0xEAe8229c0E457efBA1A1769e7F8c20110fF68E61"

	ZCNTxnID = "b26abeb31fcee5d2e75b26717722938a06fa5ce4a5b5e68ddad68357432caace"
	Amount   = 1e10
	TxnFee   = 1
	Nonce    = 1
)

var (
	Signatures = []*ethereum.AuthorizerSignature{
		{
			ID:        "0x2ec8F26ccC678c9faF0Df20208aEE3AF776160CD",
			Signature: []byte("0xEAe8229c0E457efBA1A1769e7F8c20110fF68E61"),
		},
	}
)

type ethereumClientMock struct {
	mock.TestingT
}

func (ecm *ethereumClientMock) Cleanup(callback func()) {
	callback()
}

type transactionProviderMock struct {
	mock.TestingT
}

func (tem *transactionProviderMock) Cleanup(callback func()) {
	callback()
}

type authorizerConfigTarget struct {
	Fee sdkcommon.Balance `json:"fee"`
}

type authorizerNodeTarget struct {
	ID        string                  `json:"id"`
	PublicKey string                  `json:"public_key"`
	URL       string                  `json:"url"`
	Config    *authorizerConfigTarget `json:"config"`
}

type authorizerConfigSource struct {
	Fee string `json:"fee"`
}

type authorizerNodeSource struct {
	ID     string                  `json:"id"`
	Config *authorizerConfigSource `json:"config"`
}

func (an *authorizerNodeTarget) decode(input []byte) error {
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(input, &objMap)
	if err != nil {
		return err
	}

	id, ok := objMap["id"]
	if ok {
		var idStr *string
		err = json.Unmarshal(*id, &idStr)
		if err != nil {
			return err
		}
		an.ID = *idStr
	}

	pk, ok := objMap["public_key"]
	if ok {
		var pkStr *string
		err = json.Unmarshal(*pk, &pkStr)
		if err != nil {
			return err
		}
		an.PublicKey = *pkStr
	}

	url, ok := objMap["url"]
	if ok {
		var urlStr *string
		err = json.Unmarshal(*url, &urlStr)
		if err != nil {
			return err
		}
		an.URL = *urlStr
	}

	rawCfg, ok := objMap["config"]
	if ok {
		var cfg = &authorizerConfigTarget{}
		err = cfg.decode(*rawCfg)
		if err != nil {
			return err
		}

		an.Config = cfg
	}

	return nil
}

func (c *authorizerConfigTarget) decode(input []byte) (err error) {
	const (
		Fee = "fee"
	)

	var objMap map[string]*json.RawMessage
	err = json.Unmarshal(input, &objMap)
	if err != nil {
		return err
	}

	fee, ok := objMap[Fee]
	if ok {
		var feeStr *string
		err = json.Unmarshal(*fee, &feeStr)
		if err != nil {
			return err
		}

		var balance, err = strconv.ParseInt(*feeStr, 10, 64)
		if err != nil {
			return err
		}

		c.Fee = sdkcommon.Balance(balance)
	}

	return nil
}

func getEthereumClient(t mock.TestingT) *bridgemocks.EthereumClient {
	return bridgemocks.NewEthereumClient(&ethereumClientMock{t})
}

func getBridgeClient(ethereumClient EthereumClient, transactionProvider transaction.TransactionProvider) *BridgeClient {
	cfg := viper.New()

	tempConfigFile, err := os.CreateTemp(".", "config.yaml")
	if err != nil {
		log.Fatalln(err)
	}

	defer os.Remove(tempConfigFile.Name())

	cfg.SetConfigFile(tempConfigFile.Name())

	cfg.SetDefault("bridge.bridge_address", BRIDGE_ADDRESS)
	cfg.SetDefault("bridge.token_address", TOKEN_ADDRESS)
	cfg.SetDefault("bridge.authorizers_address", AUTHORIZER_ADDRESS)
	cfg.SetDefault("bridge.ethereum_address", ETHEREUM_ADDRESS)
	cfg.SetDefault("bridge.password", PASSWORD)
	cfg.SetDefault("bridge.gas_limit", 0)
	cfg.SetDefault("bridge.consensus_threshold", 0)

	return createBridgeClient(cfg, ethereumClient, transactionProvider)
}

func prepareEthereumClientGeneralMockCalls(ethereumClient *mock.Mock) {
	ethereumClient.On("EstimateGas", mock.Anything, mock.Anything).Return(uint64(400000), nil)
	ethereumClient.On("ChainID", mock.Anything).Return(big.NewInt(400000), nil)
	ethereumClient.On("PendingNonceAt", mock.Anything, mock.Anything).Return(uint64(Nonce), nil)
	ethereumClient.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(400000), nil)
	ethereumClient.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
}

func getTransactionProvider(t mock.TestingT) *transactionmocks.TransactionProvider {
	return transactionmocks.NewTransactionProvider(&transactionProviderMock{t})
}

func prepareTransactionEntityGeneralMockCalls(transactionProvider *mock.Mock) {
	transactionProvider.On("ExecuteSmartContract", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(ZCNTxnID, nil)
	transactionProvider.On("Verify").Return(nil)
}

func Test_ZCNBridge(t *testing.T) {
	ethereumClient := getEthereumClient(t)
	prepareEthereumClientGeneralMockCalls(&ethereumClient.Mock)

	transactionProvider := getTransactionProvider(t)
	prepareTransactionEntityGeneralMockCalls(&transactionProvider.Mock)

	bridgeClient := getBridgeClient(ethereumClient, transactionProvider)

	t.Run("should update authorizer config.", func(t *testing.T) {
		source := &authorizerNodeSource{
			ID: "12345678",
			Config: &authorizerConfigSource{
				Fee: "999",
			},
		}
		target := &authorizerNodeTarget{}

		bytes, err := json.Marshal(source)
		require.NoError(t, err)

		err = target.decode(bytes)
		require.NoError(t, err)

		require.Equal(t, "", target.URL)
		require.Equal(t, "", target.PublicKey)
		require.Equal(t, "12345678", target.ID)
		require.Equal(t, sdkcommon.Balance(999), target.Config.Fee)
	})

	t.Run("should check configuration formating in MintWZCN", func(t *testing.T) {
		_, err := bridgeClient.MintWZCN(context.Background(), &ethereum.MintPayload{
			ZCNTxnID:   ZCNTxnID,
			Amount:     Amount,
			To:         ETHEREUM_ADDRESS,
			Nonce:      Nonce,
			Signatures: Signatures,
		})
		require.NoError(t, err)

		var sigs [][]byte
		for _, signature := range Signatures {
			sigs = append(sigs, signature.Signature)
		}

		to := common.HexToAddress(BRIDGE_ADDRESS)
		fromAddress := common.HexToAddress(ETHEREUM_ADDRESS)

		abi, err := binding.BridgeMetaData.GetAbi()
		require.NoError(t, err)

		pack, err := abi.Pack("mint", common.HexToAddress(ETHEREUM_ADDRESS),
			big.NewInt(Amount),
			DefaultClientIDEncoder(ZCNTxnID),
			big.NewInt(Nonce),
			sigs)
		require.NoError(t, err)

		require.True(t, ethereumClient.AssertCalled(
			t,
			"EstimateGas",
			context.Background(),
			eth.CallMsg{
				To:   &to,
				From: fromAddress,
				Data: pack,
			},
		))

	})

	t.Run("should check configuration formating in BurnWZCN", func(t *testing.T) {
		_, err := bridgeClient.BurnWZCN(context.Background(), Amount)
		require.NoError(t, err)

		to := common.HexToAddress(BRIDGE_ADDRESS)
		fromAddress := common.HexToAddress(ETHEREUM_ADDRESS)

		abi, err := binding.BridgeMetaData.GetAbi()
		require.NoError(t, err)

		pack, err := abi.Pack("burn", big.NewInt(Amount), DefaultClientIDEncoder(zcncore.GetClientWalletID()))
		require.NoError(t, err)

		require.True(t, ethereumClient.AssertCalled(
			t,
			"EstimateGas",
			context.Background(),
			eth.CallMsg{
				To:   &to,
				From: fromAddress,
				Data: pack,
			},
		))
	})

	t.Run("should check configuration used by MintZCN", func(t *testing.T) {
		// ethereumClient.Mock.On("createTransactionEntity")

		_, err := bridgeClient.MintZCN(context.Background(), &zcnsc.MintPayload{})
		require.NoError(t, err)

		// to := common.HexToAddress(BRIDGE_ADDRESS)
		// fromAddress := common.HexToAddress(ETHEREUM_ADDRESS)

		// abi, err := binding.BridgeMetaData.GetAbi()
		// require.NoError(t, err)

		// pack, err := abi.Pack("burn", big.NewInt(Amount), DefaultClientIDEncoder(zcncore.GetClientWalletID()))
		// require.NoError(t, err)

		// require.True(t, ethereumClient.AssertCalled(
		// 	t,
		// 	"EstimateGas",
		// 	context.Background(),
		// 	eth.CallMsg{
		// 		To:   &to,
		// 		From: fromAddress,
		// 		Data: pack,
		// 	},
		// ))
	})

	t.Run("should check configuration used by BurnZCN", func(t *testing.T) {
		_, err := bridgeClient.BurnZCN(context.Background(), Amount, TxnFee)
		require.NoError(t, err)

		// to := common.HexToAddress(BRIDGE_ADDRESS)
		// fromAddress := common.HexToAddress(ETHEREUM_ADDRESS)

		// abi, err := binding.BridgeMetaData.GetAbi()
		// require.NoError(t, err)

		// pack, err := abi.Pack("burn", big.NewInt(Amount), DefaultClientIDEncoder(zcncore.GetClientWalletID()))
		// require.NoError(t, err)

		// require.True(t, ethereumClient.AssertCalled(
		// 	t,
		// 	"EstimateGas",
		// 	context.Background(),
		// 	eth.CallMsg{
		// 		To:   &to,
		// 		From: fromAddress,
		// 		Data: pack,
		// 	},
		// ))
	})

	t.Run("should check configuration used by IncreaseBurnerAllowance", func(t *testing.T) {
		_, err := bridgeClient.IncreaseBurnerAllowance(context.Background(), Amount)
		require.NoError(t, err)

		spenderAddress := common.HexToAddress(BRIDGE_ADDRESS)

		to := common.HexToAddress(TOKEN_ADDRESS)
		fromAddress := common.HexToAddress(ETHEREUM_ADDRESS)

		abi, err := erc20.ERC20MetaData.GetAbi()
		require.NoError(t, err)

		pack, err := abi.Pack("increaseAllowance", spenderAddress, big.NewInt(Amount))
		require.NoError(t, err)

		require.True(t, ethereumClient.AssertCalled(
			t,
			"EstimateGas",
			context.Background(),
			eth.CallMsg{
				To:   &to,
				From: fromAddress,
				Data: pack,
			},
		))
	})

	t.Run("should check configuration used by CreateSignedTransactionFromKeyStore", func(t *testing.T) {
		bridgeClient.CreateSignedTransactionFromKeyStore(ethereumClient, 400000)

		require.True(t, ethereumClient.AssertCalled(
			t,
			"PendingNonceAt",
			context.Background(),
			common.HexToAddress(ETHEREUM_ADDRESS)))

		// TODO: check somehow used Password
	})
}
