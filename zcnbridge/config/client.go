package config

import (
	"flag"

	"github.com/0chain/gosdk/zcnbridge/chain"

	"github.com/spf13/viper"

	"github.com/0chain/gosdk/zcnbridge/log"
)

type ClientConfig struct {
	KeyFileDir  *string
	KeyFile     *string
	LogPath     *string
	ConfigFile  *string
	ConfigDir   *string
	Development *bool
}

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

var Client ClientConfig

// ParseClientConfig reads config from command line
func ParseClientConfig() {
	Client.Development = flag.Bool("development", true, "development mode")
	Client.KeyFileDir = flag.String("zcn_dir", ".zcn", "zcn home folder")
	Client.KeyFile = flag.String("wallet_config", "wallet.json", "wallet config")
	Client.LogPath = flag.String("log_dir", "./logs", "log folder")
	Client.ConfigDir = flag.String("config_dir", "./config", "0chain config folder")
	Client.ConfigFile = flag.String("config_file", "0chain", "0chain config file")

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
