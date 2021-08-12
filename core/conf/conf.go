// pakcage conf provide config helpers for ~/.zcn/config.yaml, ～/.zcn/network.yaml and ~/.zcn/wallet.json

package conf

import (
	"errors"
	"net/url"
	"os"
	"strings"

	thrown "github.com/0chain/errors"
	"github.com/spf13/viper"
)

var (
	// ErrMssingConfig config file is missing
	ErrMssingConfig = errors.New("[conf]missing config file")
	// ErrInvalidValue invalid value in config
	ErrInvalidValue = errors.New("[conf]invalid value")
	// ErrBadParsing fail to parse config via spf13/viper
	ErrBadParsing = errors.New("[conf]bad parsing")
)

// LoadFile load config from file
// Example:
//   conf.Load("~/.zcn/config.yaml"), it will load settings from ~/.zcn/config.yaml
func LoadFile(file string) (Config, error) {

	var cfg Config
	var err error

	_, err = os.Stat(file)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, thrown.Throw(ErrMssingConfig, file)
		}
		return cfg, err
	}

	v := viper.New()

	v.SetConfigFile(file)

	if err := v.ReadInConfig(); err != nil {
		return cfg, thrown.Throw(ErrBadParsing, err.Error())
	}

	return Load(v)
}

// Load load and parse config
func Load(v ConfigReader) (Config, error) {

	var cfg Config

	blockWorker := strings.TrimSpace(v.GetString("block_worker"))

	if isURL(blockWorker) == false {
		return cfg, thrown.Throw(ErrInvalidValue, "block_worker="+blockWorker)
	}

	minSubmit := v.GetInt("min_submit")

	if minSubmit < 1 {
		minSubmit = 3
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

func isURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}
