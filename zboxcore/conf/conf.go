// pakcage conf provide config helpers for ~/.zcn/config.yaml, ～/.zcn/network.yaml and ~/.zcn/wallet.json

package conf

import (
	"errors"
	"net/url"
	"os"
	"path"
	"strings"

	thrown "github.com/0chain/errors"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	// DefaultConfigFileName default config file in ~/.zcn
	DefaultConfigFileName = "config.yaml"

	// Config current config instance for zbox
	Config ZConfig
)

var (
	// ErrMssingConfig config file is missing
	ErrMssingConfig = errors.New("[conf]missing config file")
	// ErrInvalidValue invalid value in config
	ErrInvalidValue = errors.New("[conf]invalid value")
	// ErrBadFormat fail to parse config via spf13/viper
	ErrBadFormat = errors.New("[conf]bad format")
)

func init() {
	LoadDefault()
}

// LoadDefault load and parse config from ~/.zcn/config.yaml
func LoadDefault() error {
	return Load(DefaultConfigFileName)
}

// Load load and parse config file in ~/.zcn folder. it is ~/.zcn/config.yaml if file is invalid.
// Example:
//   conf.Load("stream.yaml"), it will load settings from ~/.zcn/stream.yaml
func Load(fileName string) error {
	file := path.Join(getConfigDir(), fileName)
	_, err := os.Stat(file)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return thrown.Throw(ErrMssingConfig, err.Error())
		}
		return err
	}

	cfg, err := loadConfigFile(file)

	if err != nil {
		return err
	}

	Config = cfg
	return nil
}

func loadConfigFile(file string) (ZConfig, error) {

	var cfg ZConfig

	v := viper.New()

	v.SetConfigFile(file)

	if err := v.ReadInConfig(); err != nil {
		return cfg, thrown.Throw(ErrBadFormat, err.Error())
	}

	blockWorker := strings.TrimSpace(v.GetString("block_worker"))

	if isURL(blockWorker) == false {
		return cfg, thrown.Throw(ErrInvalidValue, "block_worker="+blockWorker)
	}

	minSubmit := v.GetInt("min_submit")

	if minSubmit < 1 {
		minSubmit = 50
	} else if minSubmit > 100 {
		minSubmit = 100
	}

	minCfm := v.GetInt("min_confirmation")

	if minCfm < 1 {
		minCfm = 50
	} else if minCfm > 100 {
		minCfm = 100
	}

	CfmChainLength := v.GetInt("confirmation_chain_length")

	if CfmChainLength < 1 {
		CfmChainLength = 3
	}

	// additional settings depending network latency

	maxTxnQuery := v.GetInt("max_txn_query")
	if maxTxnQuery < 1 {
		maxTxnQuery = 5
	}

	querySleepTime := v.GetInt("query_sleep_time")
	if querySleepTime < 1 {
		querySleepTime = 5
	}

	cfg.BlockWorker = blockWorker
	cfg.PreferredBlobbers = v.GetStringSlice("preferred_blobbers")
	cfg.MinSubmit = minSubmit
	cfg.MinConfirmation = minCfm
	cfg.ConfirmationChainLength = CfmChainLength
	cfg.MaxTxnQuery = maxTxnQuery
	cfg.QuerySleepTime = querySleepTime

	return cfg, nil

}

// getConfigDir get config directory , default is ~/.zcn/
func getConfigDir() string {

	var configDir string
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	configDir = home + string(os.PathSeparator) + ".zcn"

	os.MkdirAll(configDir, 0744)

	return configDir
}

func isURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}
