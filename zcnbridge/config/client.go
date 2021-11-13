package config

import (
	"flag"

	"github.com/0chain/gosdk/zcnbridge/log"
)

type ClientConfig struct {
	KeyFile     *string
	LogPath     *string
	ConfigFile  *string
	ConfigDir   *string
	Development *bool
}

func (c ClientConfig) LogDir() string {
	panic("implement me")
}

func (c ClientConfig) LogLvl() string {
	panic("implement me")
}

func (c ClientConfig) BlockWorker() string {
	panic("implement me")
}

func (c ClientConfig) SignatureScheme() string {
	panic("implement me")
}

var Client ClientConfig

// ParseClientConfig reads config from command line
func ParseClientConfig() {
	Client.Development = flag.Bool("development", true, "development mode")
	Client.KeyFile = flag.String("keys_file_0chain", "", "keys_file_0chain")
	Client.LogPath = flag.String("log_dir", "", "log_dir")
	Client.ConfigDir = flag.String("config_dir", "./config", "0chain config folder")
	Client.ConfigFile = flag.String("config_file", "0chain", "0chain config file")

	flag.Parse()

	validateRequiredFlags()
}

func validateRequiredFlags() {
	required := []string{
		"keys_file_0chain",
	}

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
