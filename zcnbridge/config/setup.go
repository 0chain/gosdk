package config

import (
	"fmt"

	"github.com/0chain/gosdk/zcnbridge/chain"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/spf13/viper"
)

// SetupBridge Use this from standalone application
func SetupBridge() {
	setDefaults()
	readConfig(cmd.ConfigDir, cmd.ConfigFile)
	setupLogging()
	setupChainConfiguration()
	setupBridge()
}

func setDefaults() {
	viper.SetDefault("logging.level", "info")
}

func readConfig(configPath, configName *string) {
	viper.AddConfigPath(*configPath)
	viper.SetConfigName(*configName)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func setupLogging() {
	log.InitLogging(
		*cmd.Development,
		*cmd.LogPath,
		viper.GetString("logging.level"),
	)
}

func setupChainConfiguration() {
	chain.SetServerChain(chain.NewChain(
		viper.GetString("block_worker"),
		viper.GetString("signature_scheme"),
		viper.GetInt("min_submit"),
		viper.GetInt("min_confirmation"),
	))
}

func setupBridge() {
	Bridge.BridgeAddress = viper.GetString("bridge.BridgeAddress")
	Bridge.Mnemonic = viper.GetString("bridge.Mnemonic")
	Bridge.EthereumNodeURL = viper.GetString("bridge.EthereumNodeURL")
	Bridge.Value = viper.GetInt64("bridge.Value")
	Bridge.GasLimit = viper.GetUint64("bridge.GasLimit")
	Bridge.WzcnAddress = viper.GetString("bridge.WzcnAddress")
	Bridge.ChainID = viper.GetString("bridge.ChainID")
}
