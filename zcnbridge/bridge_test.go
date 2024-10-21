package zcnbridge

import (
	"context"
	"encoding/hex"
	"encoding/json"
	coreClient "github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zcnbridge/ethereum/uniswapnetwork"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"log"
	"math/big"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/0chain/gosdk/zcnbridge/ethereum/zcntoken"

	sdkcommon "github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zcnbridge/ethereum"
	"github.com/0chain/gosdk/zcnbridge/ethereum/authorizers"
	binding "github.com/0chain/gosdk/zcnbridge/ethereum/bridge"
	bridgemocks "github.com/0chain/gosdk/zcnbridge/mocks"
	"github.com/0chain/gosdk/zcnbridge/zcnsc"
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

	alchemyEthereumNodeURL  = "https://eth-mainnet.g.alchemy.com/v2/9VanLUbRE0pLmDHwCHGJlhs9GHosrfD9"
	tenderlyEthereumNodeURL = "https://rpc.tenderly.co/fork/835ecb4e-1f60-4129-adc2-b0c741193839"
	infuraEthereumNodeURL   = "https://mainnet.infura.io/v3/7238211010344719ad14a89db874158c"

	password = "02289b9"

	authorizerDelegatedAddress = "0xa149B58b7e1390D152383BB03dBc79B390F648e2"

	bridgeAddress      = "0x7bbbEa24ac1751317D7669f05558632c4A9113D7"
	tokenAddress       = "0xb9EF770B6A5e12E45983C5D80545258aA38F3B78"
	authorizersAddress = "0xEAe8229c0E457efBA1A1769e7F8c20110fF68E61"
	uniswapAddress     = "0x4c12C2FeEDD86267d17dB64BaB2cFD12cD8611f5"

	zcnTxnID = "b26abeb31fcee5d2e75b26717722938a06fa5ce4a5b5e68ddad68357432caace"
	amount   = 1
	txnFee   = 1
	nonce    = 1

	ethereumTxnID = "0x3b59971c2aa294739cd73912f0c5a7996aafb796238cf44408b0eb4af0fbac82" //nolint:unused

	clientId = "d6e9b3222434faa043c683d1a939d6a0fa2818c4d56e794974d64a32005330d3"
)

var (
	uniswapSmartContractCode = "60806040526004361061002d5760003560e01c806318ae74a41461003957806397a40b341461006957610034565b3661003457005b600080fd5b610053600480360381019061004e9190610781565b6100a6565b60405161006091906107bd565b60405180910390f35b34801561007557600080fd5b50610090600480360381019061008b91906107d8565b61017e565b60405161009d91906107bd565b60405180910390f35b60008060008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663fb3bdb4134856100ef61061a565b33426040518663ffffffff1660e01b81526004016101109493929190610917565b60006040518083038185885af115801561012e573d6000803e3d6000fd5b50505050506040513d6000823e3d601f19601f820116820180604052508101906101589190610ad1565b90508060018151811061016e5761016d610b1a565b5b6020026020010151915050919050565b6000600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166323b872dd3330856040518463ffffffff1660e01b81526004016101df93929190610b49565b6020604051808303816000875af11580156101fe573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906102229190610bb8565b50600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663095ea7b360008054906101000a900473ffffffffffffffffffffffffffffffffffffffff16846040518363ffffffff1660e01b81526004016102a0929190610be5565b6020604051808303816000875af11580156102bf573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906102e39190610bb8565b506060600367ffffffffffffffff81111561030157610300610979565b5b60405190808252806020026020018201604052801561032f5781602001602082028036833780820191505090505b50905073a0b86991c6218b36c1d19d4a2e9eb0ce3606eb488160008151811061035b5761035a610b1a565b5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505073c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2816001815181106103be576103bd610b1a565b5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505073b9ef770b6a5e12e45983c5d80545258aa38f3b788160028151811061042157610420610b1a565b5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505060008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16638803dbee86868533426040518663ffffffff1660e01b81526004016104bf959493929190610c0e565b6000604051808303816000875af11580156104de573d6000803e3d6000fd5b505050506040513d6000823e3d601f19601f820116820180604052508101906105079190610ad1565b9050838160008151811061051e5761051d610b1a565b5b602002602001015110156105f457600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663a9059cbb338360008151811061057f5761057e610b1a565b5b6020026020010151876105929190610c97565b6040518363ffffffff1660e01b81526004016105af929190610be5565b6020604051808303816000875af11580156105ce573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105f29190610bb8565b505b8060028151811061060857610607610b1a565b5b60200260200101519250505092915050565b60606000600267ffffffffffffffff81111561063957610638610979565b5b6040519080825280602002602001820160405280156106675781602001602082028036833780820191505090505b50905073c02aaa39b223fe8d0a0e5c4f27ead9083c756cc28160008151811061069357610692610b1a565b5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505073b9ef770b6a5e12e45983c5d80545258aa38f3b78816001815181106106f6576106f5610b1a565b5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508091505090565b6000604051905090565b600080fd5b600080fd5b6000819050919050565b61075e8161074b565b811461076957600080fd5b50565b60008135905061077b81610755565b92915050565b60006020828403121561079757610796610741565b5b60006107a58482850161076c565b91505092915050565b6107b78161074b565b82525050565b60006020820190506107d260008301846107ae565b92915050565b600080604083850312156107ef576107ee610741565b5b60006107fd8582860161076c565b925050602061080e8582860161076c565b9150509250929050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b600061086f82610844565b9050919050565b61087f81610864565b82525050565b60006108918383610876565b60208301905092915050565b6000602082019050919050565b60006108b582610818565b6108bf8185610823565b93506108ca83610834565b8060005b838110156108fb5781516108e28882610885565b97506108ed8361089d565b9250506001810190506108ce565b5085935050505092915050565b61091181610864565b82525050565b600060808201905061092c60008301876107ae565b818103602083015261093e81866108aa565b905061094d6040830185610908565b61095a60608301846107ae565b95945050505050565b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6109b182610968565b810181811067ffffffffffffffff821117156109d0576109cf610979565b5b80604052505050565b60006109e3610737565b90506109ef82826109a8565b919050565b600067ffffffffffffffff821115610a0f57610a0e610979565b5b602082029050602081019050919050565b600080fd5b600081519050610a3481610755565b92915050565b6000610a4d610a48846109f4565b6109d9565b90508083825260208201905060208402830185811115610a7057610a6f610a20565b5b835b81811015610a995780610a858882610a25565b845260208401935050602081019050610a72565b5050509392505050565b600082601f830112610ab857610ab7610963565b5b8151610ac8848260208601610a3a565b91505092915050565b600060208284031215610ae757610ae6610741565b5b600082015167ffffffffffffffff811115610b0557610b04610746565b5b610b1184828501610aa3565b91505092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b6000606082019050610b5e6000830186610908565b610b6b6020830185610908565b610b7860408301846107ae565b949350505050565b60008115159050919050565b610b9581610b80565b8114610ba057600080fd5b50565b600081519050610bb281610b8c565b92915050565b600060208284031215610bce57610bcd610741565b5b6000610bdc84828501610ba3565b91505092915050565b6000604082019050610bfa6000830185610908565b610c0760208301846107ae565b9392505050565b600060a082019050610c2360008301886107ae565b610c3060208301876107ae565b8181036040830152610c4281866108aa565b9050610c516060830185610908565b610c5e60808301846107ae565b9695505050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610ca28261074b565b9150610cad8361074b565b9250828203905081811115610cc557610cc4610c68565b5b9291505056fea26469706673582212207de082f4e5f623e928f9b99a8e233f194bacc23969b40ea49a470ecfd2a1fb8464736f6c63430008140033"
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

	zcnScSignatures = []*zcnsc.AuthorizerSignature{ //nolint:unused
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

type transactionMock struct { //nolint:unused
	mock.TestingT
}

func (tem *transactionMock) Cleanup(callback func()) { //nolint:unused
	callback()
}

type transactionProviderMock struct { //nolint:unused
	mock.TestingT
}

func (tem *transactionProviderMock) Cleanup(callback func()) { //nolint:unused
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

func getBridgeClient(ethereumNodeURL string, ethereumClient EthereumClient, keyStore KeyStore) *BridgeClient {
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
	cfg.SetDefault("bridge.uniswap_address", uniswapAddress)
	cfg.SetDefault("bridge.ethereum_address", ethereumAddress)
	cfg.SetDefault("bridge.password", password)
	cfg.SetDefault("bridge.gas_limit", 0)
	cfg.SetDefault("bridge.consensus_threshold", 0)

	return NewBridgeClient(
		cfg.GetString("bridge.bridge_address"),
		cfg.GetString("bridge.token_address"),
		cfg.GetString("bridge.authorizers_address"),
		cfg.GetString("bridge.uniswap_address"),
		cfg.GetString("bridge.ethereum_address"),
		ethereumNodeURL,
		cfg.GetString("bridge.password"),
		cfg.GetUint64("bridge.gas_limit"),
		cfg.GetFloat64("bridge.consensus_threshold"),

		ethereumClient,
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

func prepareTransactionGeneralMockCalls(transaction *mock.Mock) { //nolint:unused
	transaction.On("ExecuteSmartContract", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(zcnTxnID, nil)
	transaction.On("Verify", mock.Anything).Return(nil)
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

func Test_ZCNBridge(t *testing.T) {
	ethereumClient := getEthereumClient(t)
	prepareEthereumClientGeneralMockCalls(&ethereumClient.Mock)

	keyStore := getKeyStore(t)
	prepareKeyStoreGeneralMockCalls(keyStore)

	bridgeClient := getBridgeClient(alchemyEthereumNodeURL, ethereumClient, keyStore)

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

		rawAbi, err := binding.BridgeMetaData.GetAbi()
		require.NoError(t, err)

		pack, err := rawAbi.Pack("mint", common.HexToAddress(ethereumAddress),
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
		coreClient.SetWallet(zcncrypto.Wallet{
			ClientID: clientId,
		})
		_, err := bridgeClient.BurnWZCN(context.Background(), amount)
		require.NoError(t, err)

		to := common.HexToAddress(bridgeAddress)
		fromAddress := common.HexToAddress(ethereumAddress)

		rawAbi, err := binding.BridgeMetaData.GetAbi()
		require.NoError(t, err)

		pack, err := rawAbi.Pack("burn", big.NewInt(amount), DefaultClientIDEncoder(coreClient.ClientID()))
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

	//TODO:JAYASHTODO
	//t.Run("should check configuration used by MintZCN", func(t *testing.T) {
	//	payload := &zcnsc.MintPayload{
	//		EthereumTxnID:     ethereumTxnID,
	//		Amount:            sdkcommon.Balance(amount),
	//		Nonce:             nonce,
	//		Signatures:        zcnScSignatures,
	//		ReceivingClientID: clientId,
	//	}
	//
	//	_, err := bridgeClient.MintZCN(context.Background(), payload)
	//	require.NoError(t, err)
	//
	//	require.True(t, tx.AssertCalled(
	//		t,
	//		"ExecuteSmartContract",
	//		context.Background(),
	//		wallet.ZCNSCSmartContractAddress,
	//		wallet.MintFunc,
	//		payload,
	//		uint64(0),
	//	))
	//})
	//
	//t.Run("should check configuration used by BurnZCN", func(t *testing.T) {
	//	_, _, err := bridgeClient.BurnZCN(amount)
	//	require.NoError(t, err)
	//
	//	require.True(t, tx.AssertCalled(
	//		t,
	//		"ExecuteSmartContract",
	//		context.Background(),
	//		wallet.ZCNSCSmartContractAddress,
	//		wallet.BurnFunc,
	//		zcnsc.BurnPayload{
	//			EthereumAddress: ethereumAddress,
	//		},
	//		uint64(amount),
	//	))
	//})

	t.Run("should check configuration used by AddEthereumAuthorizer", func(t *testing.T) {
		_, err := bridgeClient.AddEthereumAuthorizer(context.Background(), common.HexToAddress(authorizerDelegatedAddress))
		require.NoError(t, err)

		to := common.HexToAddress(authorizersAddress)
		fromAddress := common.HexToAddress(ethereumAddress)

		rawAbi, err := authorizers.AuthorizersMetaData.GetAbi()
		require.NoError(t, err)

		pack, err := rawAbi.Pack("addAuthorizers", common.HexToAddress(authorizerDelegatedAddress))
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

		rawAbi, err := authorizers.AuthorizersMetaData.GetAbi()
		require.NoError(t, err)

		pack, err := rawAbi.Pack("removeAuthorizers", common.HexToAddress(authorizerDelegatedAddress))
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

		rawAbi, err := zcntoken.TokenMetaData.GetAbi()
		require.NoError(t, err)

		pack, err := rawAbi.Pack("increaseApproval", spenderAddress, big.NewInt(amount))
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

	t.Run("should check configuration used by SwapETH", func(t *testing.T) {
		uniswapSmartContractCodeRaw, err := hex.DecodeString(uniswapSmartContractCode)
		require.NoError(t, err)

		ethereumClient.On("PendingCodeAt", mock.Anything, mock.Anything).Return(uniswapSmartContractCodeRaw, nil)

		_, err = bridgeClient.SwapETH(context.Background(), amount, amount)
		require.NoError(t, err)

		// 1. To address parameter.
		to := common.HexToAddress(bridgeClient.UniswapAddress)

		// 2. From address parameter.
		from := common.HexToAddress(bridgeClient.EthereumAddress)

		// 3. Swap amount parameter
		swapAmount := big.NewInt(amount)

		var rawAbi *abi.ABI

		rawAbi, err = uniswapnetwork.UniswapMetaData.GetAbi()
		require.NoError(t, err)

		var pack []byte

		pack, err = rawAbi.Pack("swapETHForZCNExactAmountOut", swapAmount)
		require.NoError(t, err)

		require.True(t, ethereumClient.AssertCalled(
			t,
			"EstimateGas",
			context.Background(),
			eth.CallMsg{
				To:       &to,
				From:     from,
				Data:     pack,
				Value:    swapAmount,
				GasPrice: big.NewInt(400000),
			},
		))
	})

	t.Run("should check configuration used by SwapUSDC", func(t *testing.T) {
		uniswapSmartContractCodeRaw, err := hex.DecodeString(uniswapSmartContractCode)
		require.NoError(t, err)

		ethereumClient.On("PendingCodeAt", mock.Anything, mock.Anything).Return(uniswapSmartContractCodeRaw, nil)

		_, err = bridgeClient.SwapUSDC(context.Background(), amount, amount)
		require.NoError(t, err)

		// 1. To address parameter.
		to := common.HexToAddress(bridgeClient.UniswapAddress)

		// 2. From address parameter.
		from := common.HexToAddress(bridgeClient.EthereumAddress)

		// 3. Swap amount parameter
		swapAmount := big.NewInt(amount)

		var rawAbi *abi.ABI

		rawAbi, err = uniswapnetwork.UniswapMetaData.GetAbi()
		require.NoError(t, err)

		var pack []byte

		pack, err = rawAbi.Pack("swapUSDCForZCNExactAmountOut", swapAmount, swapAmount)
		require.NoError(t, err)

		require.True(t, ethereumClient.AssertCalled(
			t,
			"EstimateGas",
			context.Background(),
			eth.CallMsg{
				To:       &to,
				From:     from,
				Data:     pack,
				Value:    big.NewInt(0),
				GasPrice: big.NewInt(400000),
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

	t.Run("should check if gas price estimation works with correct alchemy ethereum node url", func(t *testing.T) {
		bridgeClient = getBridgeClient(alchemyEthereumNodeURL, ethereumClient, keyStore)

		_, err := bridgeClient.EstimateGasPrice(context.Background())
		require.Contains(t, err.Error(), "Must be authenticated!")
	})

	t.Run("should check if gas price estimation works with correct tenderly ethereum node url", func(t *testing.T) {
		bridgeClient = getBridgeClient(tenderlyEthereumNodeURL, ethereumClient, keyStore)

		_, err := bridgeClient.EstimateGasPrice(context.Background())
		require.NoError(t, err)
	})

	t.Run("should check if gas price estimation works with incorrect ethereum node url", func(t *testing.T) {
		bridgeClient = getBridgeClient(infuraEthereumNodeURL, ethereumClient, keyStore)

		_, err := bridgeClient.EstimateGasPrice(context.Background())
		require.Error(t, err)
	})
}
