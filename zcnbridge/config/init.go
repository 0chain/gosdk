package config

import (
	"fmt"

	"github.com/0chain/gosdk/zcnbridge"

	"github.com/0chain/gosdk/zcnbridge/chain"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/spf13/viper"
)

// SetupBridge Use this from standalone application
func SetupBridge(configDir, configFile string, development bool, logPath string) *zcnbridge.Bridge {
	setDefaults()
	ReadConfig(configDir, configFile)
	setupLogging(development, logPath)
	SetupChainFromConfig()
	bridge := zcnbridge.SetupBridgeFromConfig()

	return bridge
}

func setDefaults() {
	viper.SetDefault("logging.level", "info")
}

func ReadConfig(configPath, configName string) {
	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func setupLogging(development bool, logPath string) {
	log.InitLogging(development, logPath, viper.GetString("logging.level"))
}

func SetupChainFromConfig() {
	chain.SetServerChain(chain.NewChain(
		viper.GetString("block_worker"),
		viper.GetString("signature_scheme"),
		viper.GetInt("min_submit"),
		viper.GetInt("min_confirmation"),
	))
}
