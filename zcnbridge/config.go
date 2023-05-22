package zcnbridge

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"

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
	EthereumAddress,
	Password,
	EthereumNodeURL,
	Homedir string

	ConsensusThreshold float64
	GasLimit           uint64
}

func CreateBridgeClient(cfg *viper.Viper, walletFile ...string) *BridgeClient {

	homedir := path.Dir(cfg.ConfigFileUsed())
	if homedir == "" {
		log.Logger.Fatal("homedir is required")
	}

	return &BridgeClient{
		BridgeAddress:      cfg.GetString("bridge.bridge_address"),
		TokenAddress:       cfg.GetString("bridge.token_address"),
		AuthorizersAddress: cfg.GetString("bridge.authorizers_address"),
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
func SetupBridgeClientSDK(cfg *BridgeSDKConfig, walletFile ...string) *BridgeClient {
	log.InitLogging(*cfg.Development, *cfg.LogPath, *cfg.LogLevel)
	bridgeClient := CreateBridgeClient(initChainConfig(cfg), walletFile...)
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
		_, file, line, ok := runtime.Caller(2)
		if !ok {
			file = "???"
			line = 0
		}
		f := filepath.Base(file)
		header := fmt.Sprintf("[ERROR] %s:%d: ", f, line)

		log.Logger.Fatal(fmt.Errorf("%w: %s: can't read config", err, header).Error())
	}
	return cfg
}
