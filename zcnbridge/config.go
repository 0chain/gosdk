package zcnbridge

import (
	"context"
	"fmt"
	"math/big"
	"path"

	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/zcnbridge/transaction"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/spf13/viper"
)

const (
	ZChainsClientConfigName  = "config.yaml"
	ZChainWalletConfigName   = "wallet.json"
	EthereumWalletStorageDir = "wallets"
)

type BridgeSDKConfig struct {
	LogLevel        *string
	LogPath         *string
	ConfigChainFile *string
	ConfigDir       *string
	Development     *bool
}

type BridgeClient struct {
	KeyStore
	transaction.TransactionProvider
	EthereumClient

	BridgeAddress,
	TokenAddress,
	AuthorizersAddress,
	EthereumAddress,
	Password string

	ConsensusThreshold float64
	GasLimit           uint64
}

// EthereumClient describes Ethereum JSON-RPC client generealized interface
type EthereumClient interface {
	bind.ContractBackend

	ChainID(ctx context.Context) (*big.Int, error)
}

// createBridgeClient initializes new bridge client with the help of the given
// Ethereum JSON-RPC client and locally-defined confiruration.
func createBridgeClient(cfg *viper.Viper, ethereumClient EthereumClient, transactionProvider transaction.TransactionProvider, keyStore KeyStore) *BridgeClient {
	return &BridgeClient{
		BridgeAddress:       cfg.GetString("bridge.bridge_address"),
		TokenAddress:        cfg.GetString("bridge.token_address"),
		AuthorizersAddress:  cfg.GetString("bridge.authorizers_address"),
		EthereumAddress:     cfg.GetString("bridge.ethereum_address"),
		Password:            cfg.GetString("bridge.password"),
		GasLimit:            cfg.GetUint64("bridge.gas_limit"),
		ConsensusThreshold:  cfg.GetFloat64("bridge.consensus_threshold"),
		EthereumClient:      ethereumClient,
		TransactionProvider: transactionProvider,
		KeyStore:            keyStore,
	}
}

// SetupBridgeClientSDK Use this from standalone application
// 0Chain SDK initialization is required
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

	ks := NewKeyStore(path.Join(homedir, EthereumWalletStorageDir))

	bridgeClient := createBridgeClient(chainCfg, ethereumClient, transactionProvider, ks)
	return bridgeClient
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
