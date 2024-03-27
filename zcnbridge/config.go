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
	ZChainsClientConfigName  = "config.yaml"
	ZChainWalletConfigName   = "wallet.json"
	EthereumWalletStorageDir = "wallets"
)

const (
	BancorNetworkAddress   = "0xeEF417e1D5CC832e619ae18D2F140De2999dD4fB"
	SourceTokenETHAddress  = "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"
	SourceTokenUSDCAddress = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
	SourceTokenEURCAddress = "0x1aBaEA1f7C830bD89Acc67eC4af516284b1bC33c"
	SourceTokenBNTAddress  = "0x1f573d6fb3f13d689ff844b4ce37794d79a7ff1c"
)

const BancorAPIURL = "https://api-v3.bancor.network"

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

type BridgeClient struct {
	keyStore            KeyStore
	transactionProvider transaction.TransactionProvider
	ethereumClient      EthereumClient

	BridgeAddress,
	TokenAddress,
	AuthorizersAddress,
	NFTConfigAddress,
	EthereumAddress,
	EthereumNodeURL,
	Password string

	BancorAPIURL string

	ConsensusThreshold float64
	GasLimit           uint64
}

// NewBridgeClient creates BridgeClient with the given parameters.
func NewBridgeClient(
	bridgeAddress,
	tokenAddress,
	authorizersAddress,
	ethereumAddress,
	ethereumNodeURL,
	password string,
	gasLimit uint64,
	consensusThreshold float64,
	bancorAPIURL string,
	ethereumClient EthereumClient,
	transactionProvider transaction.TransactionProvider,
	keyStore KeyStore) *BridgeClient {
	return &BridgeClient{
		BridgeAddress:       bridgeAddress,
		TokenAddress:        tokenAddress,
		AuthorizersAddress:  authorizersAddress,
		EthereumAddress:     ethereumAddress,
		EthereumNodeURL:     ethereumNodeURL,
		Password:            password,
		GasLimit:            gasLimit,
		ConsensusThreshold:  consensusThreshold,
		BancorAPIURL:        bancorAPIURL,
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
		chainCfg.GetString("bridge.ethereum_address"),
		ethereumNodeURL,
		chainCfg.GetString("bridge.password"),
		chainCfg.GetUint64("bridge.gas_limit"),
		chainCfg.GetFloat64("bridge.consensus_threshold"),
		BancorAPIURL,
		ethereumClient,
		transactionProvider,
		keyStore,
	)
}
