package zcnbridge

import (
	"fmt"
	"path"

	"github.com/0chain/gosdk/zcnbridge/log"

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
	BridgeAddress,
	TokenAddress,
	AuthorizersAddress,
	NFTConfigAddress,
	EthereumAddress,
	Password,
	EthereumNodeURL,
	Homedir string

	ConsensusThreshold float64
	GasLimit           uint64
}

func CreateBridgeClient(cfg *viper.Viper) *BridgeClient {

	homedir := path.Dir(cfg.ConfigFileUsed())
	if homedir == "" {
		log.Logger.Fatal("homedir is required")
	}

	return &BridgeClient{
		BridgeAddress:      cfg.GetString("bridge.bridge_address"),
		TokenAddress:       cfg.GetString("bridge.token_address"),
		AuthorizersAddress: cfg.GetString("bridge.authorizers_address"),
		NFTConfigAddress:   cfg.GetString("bridge.nft_config_address"),
		EthereumAddress:    cfg.GetString("bridge.ethereum_address"),
		Password:           cfg.GetString("bridge.password"),
		EthereumNodeURL:    cfg.GetString("ethereum_node_url"),
		GasLimit:           cfg.GetUint64("bridge.gas_limit"),
		ConsensusThreshold: cfg.GetFloat64("bridge.consensus_threshold"),
		Homedir:            homedir,
	}
}

// SetupBridgeClientSDK Use this from standalone application
// 0Chain SDK initialization is required
func SetupBridgeClientSDK(cfg *BridgeSDKConfig) *BridgeClient {
	log.InitLogging(*cfg.Development, *cfg.LogPath, *cfg.LogLevel)
	bridgeClient := CreateBridgeClient(initChainConfig(cfg))
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
