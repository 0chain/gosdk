package config

import (
	"flag"

	"github.com/0chain/gosdk/zcnbridge/chain"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/spf13/viper"
)

type ClientConfig struct {
	WalletFileConfig *string
	LogPath          *string
	ConfigFile       *string
	ConfigDir        *string
	Development      *bool
}

var cmd ClientConfig

func (c ClientConfig) LogDir() string {
	return *c.LogPath
}

func (c ClientConfig) LogLvl() string {
	return viper.GetString("logging.level")
}

func (c ClientConfig) BlockWorker() string {
	return chain.GetServerChain().BlockWorker
}

func (c ClientConfig) SignatureScheme() string {
	return chain.GetServerChain().SignatureScheme
}

func GetSDKConfig() ChainConfig {
	return cmd
}

func GetWalletFileConfig() string {
	return *cmd.WalletFileConfig
}

// ReadClientConfigFromCmd reads config from command line
func ReadClientConfigFromCmd() {
	cmd.Development = flag.Bool("development", true, "development mode")
	cmd.WalletFileConfig = flag.String("wallet_config", "wallet.json", "wallet config")
	cmd.LogPath = flag.String("log_dir", "./logs", "log folder")
	cmd.ConfigDir = flag.String("config_dir", "./config", "0chain config folder")
	cmd.ConfigFile = flag.String("config_file", "0chain", "0chain config file")

	flag.Parse()

	validateRequiredFlags()
}

func validateRequiredFlags() {
	required := []string{}

	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	for _, req := range required {
		if !seen[req] {
			msg := "missing required: '" + req + "' argument or flag"
			log.Logger.Fatal(msg)
			panic(msg)
		}
	}
}
