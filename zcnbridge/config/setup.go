package config

import (
	"fmt"

	"github.com/0chain/gosdk/zcnbridge/chain"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/spf13/viper"
)

func Setup() {
	setDefault()
	setupConfig(Client.ConfigDir, Client.ConfigFile)
	setupLogging()
	setupChainConfiguration()
}

func setDefault() {
	viper.SetDefault("logging.level", "info")
}

func setupConfig(configPath, configName *string) {
	viper.AddConfigPath(*configPath)
	viper.SetConfigName(*configName)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func setupLogging() {
	log.InitLogging(
		*Client.Development,
		*Client.LogPath,
		viper.GetString("logging.level"),
	)
}

func setupChainConfiguration() {
	chain.SetServerChain(chain.NewChain(
		viper.GetString("server_chain.id"),
		viper.GetString("server_chain.owner"),
		viper.GetString("block_worker"),
		viper.GetString("server_chain.signature_scheme"),
	))
}
