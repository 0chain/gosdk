package zcnbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/0chain/gosdk/zcnbridge/ethereum/bancornetwork"
	"github.com/0chain/gosdk/zcnbridge/ethereum/zcntoken"

	sdkcommon "github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zcnbridge/ethereum"
	"github.com/0chain/gosdk/zcnbridge/ethereum/authorizers"
	binding "github.com/0chain/gosdk/zcnbridge/ethereum/bridge"
	bridgemocks "github.com/0chain/gosdk/zcnbridge/mocks"
	"github.com/0chain/gosdk/zcnbridge/transaction"
	transactionmocks "github.com/0chain/gosdk/zcnbridge/transaction/mocks"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/0chain/gosdk/zcnbridge/zcnsc"
	"github.com/0chain/gosdk/zcncore"
	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	ethereumAddress = "0xD8c9156e782C68EE671C09b6b92de76C97948432"

	alchemyEthereumNodeURL = "https://eth-mainnet.g.alchemy.com/v2/9VanLUbRE0pLmDHwCHGJlhs9GHosrfD9"
	infuraEthereumNodeURL  = "https://mainnet.infura.io/v3/7238211010344719ad14a89db874158c"
	value                  = uint64(1e+20)

	password = "02289b9"

	authorizerDelegatedAddress = "0xa149B58b7e1390D152383BB03dBc79B390F648e2"

	bridgeAddress      = "0x7bbbEa24ac1751317D7669f05558632c4A9113D7"
	tokenAddress       = "0xb9EF770B6A5e12E45983C5D80545258aA38F3B78"
	authorizersAddress = "0xEAe8229c0E457efBA1A1769e7F8c20110fF68E61"

	sourceAddress = "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"

	zcnTxnID = "b26abeb31fcee5d2e75b26717722938a06fa5ce4a5b5e68ddad68357432caace"
	amount   = 1
	txnFee   = 1
	nonce    = 1

	ethereumTxnID = "0x3b59971c2aa294739cd73912f0c5a7996aafb796238cf44408b0eb4af0fbac82"

	clientId = "d6e9b3222434faa043c683d1a939d6a0fa2818c4d56e794974d64a32005330d3"
)

var (
	testKeyStoreLocation = path.Join(".", EthereumWalletStorageDir)
)

var (
	ethereumSignatures = []*ethereum.AuthorizerSignature{
		{
			ID:        "0x2ec8F26ccC678c9faF0Df20208aEE3AF776160CD",
			Signature: []byte("0xEAe8229c0E457efBA1A1769e7F8c20110fF68E61"),
		},
	}

	zcnScSignatures = []*zcnsc.AuthorizerSignature{
		{
			ID:        "0x2ec8F26ccC678c9faF0Df20208aEE3AF776160CD",
			Signature: "0xEAe8229c0E457efBA1A1769e7F8c20110fF68E61",
		},
	}
)

type ethereumClientMock struct {
	mock.TestingT
}

func (ecm *ethereumClientMock) Cleanup(callback func()) {
	callback()
}

type transactionMock struct {
	mock.TestingT
}

func (tem *transactionMock) Cleanup(callback func()) {
	callback()
}

type transactionProviderMock struct {
	mock.TestingT
}

func (tem *transactionProviderMock) Cleanup(callback func()) {
	callback()
}

type keyStoreMock struct {
	mock.TestingT
}

func (ksm *keyStoreMock) Cleanup(callback func()) {
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

func getBridgeClient(bancorAPIURL, ethereumNodeURL string, ethereumClient EthereumClient, transactionProvider transaction.TransactionProvider, keyStore KeyStore) *BridgeClient {
	cfg := viper.New()

	tempConfigFile, err := os.CreateTemp(".", "config.yaml")
	if err != nil {
		log.Fatalln(err)
	}

	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			log.Fatalln(err)
		}
	}(tempConfigFile.Name())

	cfg.SetConfigFile(tempConfigFile.Name())

	cfg.SetDefault("bridge.bridge_address", bridgeAddress)
	cfg.SetDefault("bridge.token_address", tokenAddress)
	cfg.SetDefault("bridge.authorizers_address", authorizersAddress)
	cfg.SetDefault("bridge.ethereum_address", ethereumAddress)
	cfg.SetDefault("bridge.password", password)
	cfg.SetDefault("bridge.gas_limit", 0)
	cfg.SetDefault("bridge.consensus_threshold", 0)

	return NewBridgeClient(
		cfg.GetString("bridge.bridge_address"),
		cfg.GetString("bridge.token_address"),
		cfg.GetString("bridge.authorizers_address"),
		ethereumNodeURL,
		cfg.GetString("ethereum_node_url"),
		cfg.GetString("bridge.password"),
		cfg.GetUint64("bridge.gas_limit"),
		cfg.GetFloat64("bridge.consensus_threshold"),
		bancorAPIURL,
		ethereumClient,
		transactionProvider,
		keyStore,
	)
}

func prepareEthereumClientGeneralMockCalls(ethereumClient *mock.Mock) {
	ethereumClient.On("EstimateGas", mock.Anything, mock.Anything).Return(uint64(400000), nil)
	ethereumClient.On("ChainID", mock.Anything).Return(big.NewInt(400000), nil)
	ethereumClient.On("PendingNonceAt", mock.Anything, mock.Anything).Return(uint64(nonce), nil)
	ethereumClient.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(400000), nil)
	ethereumClient.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
}

func getTransaction(t mock.TestingT) *transactionmocks.Transaction {
	return transactionmocks.NewTransaction(&transactionMock{t})
}

func prepareTransactionGeneralMockCalls(transaction *mock.Mock) {
	transaction.On("ExecuteSmartContract", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(zcnTxnID, nil)
	transaction.On("Verify", mock.Anything).Return(nil)
}

func getTransactionProvider(t mock.TestingT) *transactionmocks.TransactionProvider {
	return transactionmocks.NewTransactionProvider(&transactionProviderMock{t})
}

func prepareTransactionProviderGeneralMockCalls(transactionProvider *mock.Mock, transaction *transactionmocks.Transaction) {
	transactionProvider.On("NewTransactionEntity", mock.Anything).Return(transaction, nil)
}

func getKeyStore(t mock.TestingT) *bridgemocks.KeyStore {
	return bridgemocks.NewKeyStore(&keyStoreMock{t})
}

func prepareKeyStoreGeneralMockCalls(keyStore *bridgemocks.KeyStore) {
	ks := keystore.NewKeyStore(testKeyStoreLocation, keystore.StandardScryptN, keystore.StandardScryptP)

	keyStore.On("Find", mock.Anything).Return(accounts.Account{Address: common.HexToAddress(ethereumAddress)}, nil)
	keyStore.On("TimedUnlock", mock.Anything, mock.Anything, mock.Anything).Run(
		func(args mock.Arguments) {
			err := ks.TimedUnlock(args.Get(0).(accounts.Account), args.Get(1).(string), args.Get(2).(time.Duration))
			if err != nil {
				log.Fatalln(err)
			}
		},
	).Return(nil)
	keyStore.On("SignHash", mock.Anything, mock.Anything).Return([]byte(ethereumAddress), nil)

	keyStore.On("GetEthereumKeyStore").Return(ks)
}

func prepareBancorMockServer() string {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := fmt.Fprintln(w, `{"data":{"dltId":"0xb9EF770B6A5e12E45983C5D80545258aA38F3B78","symbol":"ZCN","decimals":10,"rate":{"bnt":"0.175290404525335519","usd":"0.100266","eur":"0.094499","eth":"1"},"rate24hAgo":{"bnt":"0.175290404525335519","usd":"0.100266","eur":"0.094499","eth":"0.000064086171894462"}},"timestamp":{"ethereum":{"block":18333798,"timestamp":1697107211}}}`)
			if err != nil {
				log.Fatalln(err)
			}
		}))

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		load := time.NewTicker(time.Millisecond * 500)

		for range load.C {
			select {
			case <-sigs:
				load.Stop()

				ts.Close()

				close(sigs)
			default:
			}
		}
	}()

	return ts.URL
}

func Test_ZCNBridge(t *testing.T) {
	ethereumClient := getEthereumClient(t)
	prepareEthereumClientGeneralMockCalls(&ethereumClient.Mock)

	tx := getTransaction(t)
	prepareTransactionGeneralMockCalls(&tx.Mock)

	transactionProvider := getTransactionProvider(t)
	prepareTransactionProviderGeneralMockCalls(&transactionProvider.Mock, tx)

	keyStore := getKeyStore(t)
	prepareKeyStoreGeneralMockCalls(keyStore)

	bancorMockServerURL := prepareBancorMockServer()

	bridgeClient := getBridgeClient(bancorMockServerURL, alchemyEthereumNodeURL, ethereumClient, transactionProvider, keyStore)

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
			ZCNTxnID:   zcnTxnID,
			Amount:     amount,
			To:         ethereumAddress,
			Nonce:      nonce,
			Signatures: ethereumSignatures,
		})
		require.NoError(t, err)

		var sigs [][]byte
		for _, signature := range ethereumSignatures {
			sigs = append(sigs, signature.Signature)
		}

		to := common.HexToAddress(bridgeAddress)
		fromAddress := common.HexToAddress(ethereumAddress)

		abi, err := binding.BridgeMetaData.GetAbi()
		require.NoError(t, err)

		pack, err := abi.Pack("mint", common.HexToAddress(ethereumAddress),
			big.NewInt(amount),
			DefaultClientIDEncoder(zcnTxnID),
			big.NewInt(nonce),
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
		_, err := bridgeClient.BurnWZCN(context.Background(), amount)
		require.NoError(t, err)

		to := common.HexToAddress(bridgeAddress)
		fromAddress := common.HexToAddress(ethereumAddress)

		abi, err := binding.BridgeMetaData.GetAbi()
		require.NoError(t, err)

		pack, err := abi.Pack("burn", big.NewInt(amount), DefaultClientIDEncoder(zcncore.GetClientWalletID()))
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
		payload := &zcnsc.MintPayload{
			EthereumTxnID:     ethereumTxnID,
			Amount:            sdkcommon.Balance(amount),
			Nonce:             nonce,
			Signatures:        zcnScSignatures,
			ReceivingClientID: clientId,
		}

		_, err := bridgeClient.MintZCN(context.Background(), payload)
		require.NoError(t, err)

		require.True(t, tx.AssertCalled(
			t,
			"ExecuteSmartContract",
			context.Background(),
			wallet.ZCNSCSmartContractAddress,
			wallet.MintFunc,
			payload,
			uint64(0),
		))
	})

	t.Run("should check configuration used by BurnZCN", func(t *testing.T) {
		_, err := bridgeClient.BurnZCN(context.Background(), amount, txnFee)
		require.NoError(t, err)

		require.True(t, tx.AssertCalled(
			t,
			"ExecuteSmartContract",
			context.Background(),
			wallet.ZCNSCSmartContractAddress,
			wallet.BurnFunc,
			zcnsc.BurnPayload{
				EthereumAddress: ethereumAddress,
			},
			uint64(amount),
		))
	})

	t.Run("should check configuration used by AddEthereumAuthorizer", func(t *testing.T) {
		_, err := bridgeClient.AddEthereumAuthorizer(context.Background(), common.HexToAddress(authorizerDelegatedAddress))
		require.NoError(t, err)

		to := common.HexToAddress(authorizersAddress)
		fromAddress := common.HexToAddress(ethereumAddress)

		abi, err := authorizers.AuthorizersMetaData.GetAbi()
		require.NoError(t, err)

		pack, err := abi.Pack("addAuthorizers", common.HexToAddress(authorizerDelegatedAddress))
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

	t.Run("should check configuration used by RemoveAuthorizer", func(t *testing.T) {
		_, err := bridgeClient.RemoveEthereumAuthorizer(context.Background(), common.HexToAddress(authorizerDelegatedAddress))
		require.NoError(t, err)

		to := common.HexToAddress(authorizersAddress)
		fromAddress := common.HexToAddress(ethereumAddress)

		abi, err := authorizers.AuthorizersMetaData.GetAbi()
		require.NoError(t, err)

		pack, err := abi.Pack("removeAuthorizers", common.HexToAddress(authorizerDelegatedAddress))
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

	t.Run("should check configuration used by IncreaseBurnerAllowance", func(t *testing.T) {
		_, err := bridgeClient.IncreaseBurnerAllowance(context.Background(), amount)
		require.NoError(t, err)

		spenderAddress := common.HexToAddress(bridgeAddress)

		to := common.HexToAddress(tokenAddress)
		fromAddress := common.HexToAddress(ethereumAddress)

		abi, err := zcntoken.TokenMetaData.GetAbi()
		require.NoError(t, err)

		pack, err := abi.Pack("increaseApproval", spenderAddress, big.NewInt(amount))
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

	t.Run("should check configuration used by Swap", func(t *testing.T) {
		// 1. Predefined deadline period
		deadlinePeriod := time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)

		_, err := bridgeClient.Swap(context.Background(), SourceTokenETHAddress, amount, big.NewInt(amount), deadlinePeriod)
		require.NoError(t, err)

		// 2. Trade deadline
		deadline := big.NewInt(deadlinePeriod.Unix())

		// 3. Swap amount parameter
		swapAmount := big.NewInt(amount)

		// 4. User's Ethereum wallet address.
		beneficiary := common.HexToAddress(ethereumAddress)

		// 5. Source zcntoken address parameter
		from := common.HexToAddress(sourceAddress)

		// 6. Target zcntoken address parameter
		to := common.HexToAddress(tokenAddress)

		// 7. Max trade zcntoken amount
		maxAmount := big.NewInt(amount)

		// 8. Bancor network smart contract address
		contractAddress := common.HexToAddress(BancorNetworkAddress)

		abi, err := bancornetwork.BancorMetaData.GetAbi()
		require.NoError(t, err)

		pack, err := abi.Pack("tradeByTargetAmount", from, to, swapAmount, maxAmount, deadline, beneficiary)
		require.NoError(t, err)

		require.True(t, ethereumClient.AssertCalled(
			t,
			"EstimateGas",
			context.Background(),
			eth.CallMsg{
				To:    &contractAddress,
				From:  beneficiary,
				Data:  pack,
				Value: maxAmount,
			},
		))
	})

	t.Run("should check configuration used by CreateSignedTransactionFromKeyStore", func(t *testing.T) {
		bridgeClient.CreateSignedTransactionFromKeyStore(ethereumClient, 400000)

		require.True(t, ethereumClient.AssertCalled(
			t,
			"PendingNonceAt",
			context.Background(),
			common.HexToAddress(ethereumAddress)))

		require.True(t, keyStore.AssertCalled(
			t,
			"TimedUnlock",
			accounts.Account{
				Address: common.HexToAddress(ethereumAddress),
			},
			password,
			time.Second*2,
		))
	})

	t.Run("should check if gas price estimation works with correct ethereum node url", func(t *testing.T) {
		bridgeClient := getBridgeClient(bancorMockServerURL, alchemyEthereumNodeURL, ethereumClient, transactionProvider, keyStore)

		result, err := bridgeClient.EstimateGasPrice(context.Background(), tokenAddress, bridgeAddress, value)
		require.NoError(t, err)
	})

	t.Run("should check if gas price estimation works with incorrect ethereum node url", func(t *testing.T) {
		bridgeClient := getBridgeClient(bancorMockServerURL, infuraEthereumNodeURL, ethereumClient, transactionProvider, keyStore)

		result, err := bridgeClient.EstimateGasPrice(context.Background(), tokenAddress, bridgeAddress, value)
		require.Error(t, err)
	})
}
