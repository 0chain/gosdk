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
	"github.com/0chain/gosdk/zcnbridge/mocks"
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
	To       = "0x2ec8F26ccC678c9faF0Df20208aEE3AF776160CD"
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

func getEthereumClient(t mock.TestingT) *mocks.EthereumClient {
	return mocks.NewEthereumClient(&ethereumClientMock{t})
}

func getBridgeClient(ethereumClient EthereumClient) *BridgeClient {
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

	return CreateBridgeClient(cfg, ethereumClient)
}

func prepareGeneralMockCalls(m *mock.Mock) {
	m.On("EstimateGas", mock.Anything, mock.Anything).Return(uint64(400000), nil)
	m.On("ChainID", mock.Anything).Return(big.NewInt(400000), nil)
	m.On("PendingNonceAt", mock.Anything, mock.Anything).Return(uint64(Nonce), nil)
	m.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(400000), nil)
	m.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
}

func Test_ZCNBridge(t *testing.T) {
	ethereumClient := getEthereumClient(t)

	prepareGeneralMockCalls(&ethereumClient.Mock)

	bridgeClient := getBridgeClient(ethereumClient)

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

	t.Run("should check signature and other data formating in MintWZCN", func(t *testing.T) {
		ctx := context.Background()

		_, err := bridgeClient.MintWZCN(ctx, &ethereum.MintPayload{
			ZCNTxnID:   ZCNTxnID,
			Amount:     Amount,
			To:         To,
			Nonce:      Nonce,
			Signatures: Signatures,
		})
		require.NoError(t, err)

		var sigs [][]byte
		for _, signature := range Signatures {
			sigs = append(sigs, signature.Signature)
		}

		ethereumClient.AssertCalled(
			t,
			"prepareBridge",
			ctx,
			To,
			"mint",
			common.HexToAddress(To),
			big.NewInt(Amount),
			DefaultClientIDEncoder(ZCNTxnID),
			big.NewInt(Nonce),
			sigs)
	})

	t.Run("should check data formating in BurnWZCN", func(t *testing.T) {

	})

	t.Run("should check data used by BurnZCN", func(t *testing.T) {

	})

	t.Run("should check data used by IncreaseBurnerAllowance", func(t *testing.T) {

	})

	t.Run("should check data used by CreateSignedTransactionFromKeyStore", func(t *testing.T) {

	})
}
