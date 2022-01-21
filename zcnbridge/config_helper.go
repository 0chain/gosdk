package zcnbridge

import (
	"fmt"
	"os"
	"path"

	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/zcnbridge/chain"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/spf13/viper"
)

func GetConfigDir() string {
	var configDir string
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	configDir = home + "/.zcn"
	return configDir
}

func initChainFromConfig(filename string) {
	configDir := GetConfigDir()
	chainConfig := viper.New()
	chainConfig.AddConfigPath(configDir)
	chainConfig.SetConfigFile(path.Join(configDir, filename))

	if err := chainConfig.ReadInConfig(); err != nil {
		ExitWithError("Can't read config: ", err)
	}

	InitChainFromConfig(chainConfig)
}

func restoreChain() {
	config, err := conf.GetClientConfig()
	if err != nil {
		ExitWithError("Can't read config:", err)
	}

	RestoreChainFromConfig(config)
}

func readSDKConfig(sdkConfig *BridgeSDKConfig) *viper.Viper {
	cfg := viper.New()
	cfg.AddConfigPath(*sdkConfig.ConfigDir)
	cfg.SetConfigName(*sdkConfig.ConfigFile)
	err := cfg.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	log.InitLogging(*sdkConfig.Development, *sdkConfig.LogPath, *sdkConfig.LogLevel)

	return cfg
}

func RestoreChainFromConfig(cfg *conf.Config) {
	chain.SetServerChain(chain.NewChain(
		cfg.BlockWorker,
		cfg.SignatureScheme,
		cfg.MinSubmit,
		cfg.MinConfirmation,
	))
}

func InitChainFromConfig(reader conf.Reader) {
	chain.SetServerChain(chain.NewChain(
		reader.GetString("block_worker"),
		reader.GetString("signature_scheme"),
		reader.GetInt("min_submit"),
		reader.GetInt("min_confirmation"),
	))
}

func ExitWithError(v ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}
