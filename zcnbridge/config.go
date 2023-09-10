package zcnbridge

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"math/big"
	"path"

	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/zcnbridge/transaction"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/spf13/viper"
)

const (
	ZChainsClientConfigName  = "config.yaml"
	ZChainWalletConfigName   = "wallet.json"
	EthereumWalletStorageDir = "wallets"
)

const BancorNetworkAddress = "0xeEF417e1D5CC832e619ae18D2F140De2999dD4fB"

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
	Password string

	// Source token address for swap operation.
	SourceTokenAddress string

	ConsensusThreshold float64
	GasLimit           uint64
}

// NewBridgeClient creates BridgeClient with the given parameters.
func NewBridgeClient(
	bridgeAddress,
	tokenAddress,
	authorizersAddress,
	ethereumAddress,
	password,
	sourceTokenAddress string,
	gasLimit uint64,
	consensusThreshold float64,
	ethereumClient EthereumClient,
	transactionProvider transaction.TransactionProvider,
	keyStore KeyStore) *BridgeClient {
	return &BridgeClient{
		BridgeAddress:       bridgeAddress,
		TokenAddress:        tokenAddress,
		AuthorizersAddress:  authorizersAddress,
		EthereumAddress:     ethereumAddress,
		Password:            password,
		SourceTokenAddress:  sourceTokenAddress,
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
func SetupBridgeClientSDK(cfg *BridgeSDKConfig) *BridgeClient {
	log.InitLogging(*cfg.Development, *cfg.LogPath, *cfg.LogLevel)

	chainCfg := initChainConfig(cfg)

	ethereumClient, err := ethclient.Dial(chainCfg.GetString("ethereum_node_url"))
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
		chainCfg.GetString("bridge.password"),
		chainCfg.GetString("bridge.swap.source_token_address"),
		chainCfg.GetUint64("bridge.gas_limit"),
		chainCfg.GetFloat64("bridge.consensus_threshold"),
		ethereumClient,
		transactionProvider,
		keyStore,
	)
}
