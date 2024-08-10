package zcnbridge

import (
	"context"
	"fmt"
	"math/big"
	"path"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/zcnbridge/transaction"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/spf13/viper"
)

const (
	TenderlyProvider = iota
	AlchemyProvider
	UnknownProvider
)

const (
	EthereumWalletStorageDir = "wallets"
)

const (
	UniswapRouterAddress = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"
	UsdcTokenAddress     = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
	WethTokenAddress     = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
)

// BridgeSDKConfig describes the configuration for the bridge SDK.
type BridgeSDKConfig struct {
	LogLevel        *string
	LogPath         *string
	ConfigChainFile *string
	ConfigDir       *string
	Development     *bool
}

// EthereumClient describes Ethereum JSON-RPC client generealized interface
type EthereumClient interface {
	bind.ContractBackend

	ChainID(ctx context.Context) (*big.Int, error)
}

// BridgeClient is a wrapper, which exposes Ethereum KeyStore methods used by DEX bridge.
type BridgeClient struct {
	keyStore            KeyStore
	transactionProvider transaction.TransactionProvider
	ethereumClient      EthereumClient

	BridgeAddress,
	TokenAddress,
	AuthorizersAddress,
	UniswapAddress,
	NFTConfigAddress,
	EthereumAddress,
	EthereumNodeURL,
	Password string

	BancorAPIURL string

	ConsensusThreshold float64
	GasLimit           uint64
}

// NewBridgeClient creates BridgeClient with the given parameters.
//   - bridgeAddress is the address of the bridge smart contract on the Ethereum network.
//   - tokenAddress is the address of the token smart contract on the Ethereum network.
//   - authorizersAddress is the address of the authorizers smart contract on the Ethereum network.
//   - authorizersAddress is the address of the authorizers smart contract on the Ethereum network.
//   - uniswapAddress is the address of the user's ethereum wallet (on UniSwap).
//   - ethereumAddress is the address of the user's ethereum wallet.
//   - ethereumNodeURL is the URL of the Ethereum node.
//   - password is the password for the user's ethereum wallet.
//   - gasLimit is the gas limit for the transactions.
//   - consensusThreshold is the consensus threshold, the minimum percentage of authorizers that need to agree on a transaction.
//   - ethereumClient is the Ethereum JSON-RPC client.
//   - transactionProvider provider interface for the transaction entity.
//   - keyStore is the Ethereum KeyStore instance.
func NewBridgeClient(
	bridgeAddress,
	tokenAddress,
	authorizersAddress,
	uniswapAddress,
	ethereumAddress,
	ethereumNodeURL,
	password string,
	gasLimit uint64,
	consensusThreshold float64,
	ethereumClient EthereumClient,
	transactionProvider transaction.TransactionProvider,
	keyStore KeyStore) *BridgeClient {
	return &BridgeClient{
		BridgeAddress:       bridgeAddress,
		TokenAddress:        tokenAddress,
		AuthorizersAddress:  authorizersAddress,
		UniswapAddress:      uniswapAddress,
		EthereumAddress:     ethereumAddress,
		EthereumNodeURL:     ethereumNodeURL,
		Password:            password,
		GasLimit:            gasLimit,
		ConsensusThreshold:  consensusThreshold,
		ethereumClient:      ethereumClient,
		transactionProvider: transactionProvider,
		keyStore:            keyStore,
	}
}

func initChainConfig(sdkConfig *BridgeSDKConfig) *viper.Viper {
	cfg := readConfig(sdkConfig, func() string {
		return *sdkConfig.ConfigChainFile
	})

	log.Logger.Info(fmt.Sprintf("Chain config has been initialized from %s", cfg.ConfigFileUsed()))

	return cfg
}

func readConfig(sdkConfig *BridgeSDKConfig, getConfigName func() string) *viper.Viper {
	cfg := viper.New()
	cfg.AddConfigPath(*sdkConfig.ConfigDir)
	cfg.SetConfigName(getConfigName())
	cfg.SetConfigType("yaml")
	err := cfg.ReadInConfig()
	if err != nil {
		log.Logger.Fatal(fmt.Errorf("%w: can't read config", err).Error())
	}
	return cfg
}

// SetupBridgeClientSDK initializes new bridge client.
// Meant to be used from standalone application with 0chain SDK initialized.
//   - cfg is the configuration for the bridge SDK.
func SetupBridgeClientSDK(cfg *BridgeSDKConfig) *BridgeClient {
	log.InitLogging(*cfg.Development, *cfg.LogPath, *cfg.LogLevel)

	chainCfg := initChainConfig(cfg)

	ethereumNodeURL := chainCfg.GetString("ethereum_node_url")

	ethereumClient, err := ethclient.Dial(ethereumNodeURL)
	if err != nil {
		Logger.Error(err)
	}

	transactionProvider := transaction.NewTransactionProvider()

	homedir := path.Dir(chainCfg.ConfigFileUsed())
	if homedir == "" {
		log.Logger.Fatal("err happened during home directory retrieval")
	}

	keyStore := NewKeyStore(path.Join(homedir, EthereumWalletStorageDir))

	return NewBridgeClient(
		chainCfg.GetString("bridge.bridge_address"),
		chainCfg.GetString("bridge.token_address"),
		chainCfg.GetString("bridge.authorizers_address"),
		chainCfg.GetString("bridge.uniswap_address"),
		chainCfg.GetString("bridge.ethereum_address"),
		ethereumNodeURL,
		chainCfg.GetString("bridge.password"),
		chainCfg.GetUint64("bridge.gas_limit"),
		chainCfg.GetFloat64("bridge.consensus_threshold"),
		ethereumClient,
		transactionProvider,
		keyStore,
	)
}
