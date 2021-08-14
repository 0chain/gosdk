package conf

import (
	"errors"
	"net/url"
	"os"
	"strings"

	thrown "github.com/0chain/errors"
	"github.com/spf13/viper"
)

const (
	// DefaultMinSubmit default value for min_submit
	DefaultMinSubmit = 50
	// DefaultMinConfirmation default value for min_confirmation
	DefaultMinConfirmation = 50
	// DefaultMaxTxnQuery default value for max_txn_query
	DefaultMaxTxnQuery = 5
	// DefaultConfirmationChainLength default value for confirmation_chain_length
	DefaultConfirmationChainLength = 3
	// DefaultQuerySleepTime default value for query_sleep_time
	DefaultQuerySleepTime = 5
)

// Config settings from ~/.zcn/config.yaml
// block_worker: http://198.18.0.98:9091
// signature_scheme: bls0chain
// min_submit: 50
// min_confirmation: 50
// confirmation_chain_length: 3
// max_txn_query: 5
// query_sleep_time: 5
// # # OPTIONAL - Uncomment to use/ Add more if you want
// # preferred_blobbers:
// #   - http://one.devnet-0chain.net:31051
// #   - http://one.devnet-0chain.net:31052
// #   - http://one.devnet-0chain.net:31053
type Config struct {
	// BlockWorker the url of 0dns's network api
	BlockWorker string
	// PreferredBlobbers preferred blobbers on new allocation
	PreferredBlobbers []string

	// MinSubmit mininal submit from blobber
	MinSubmit int
	// MinConfirmation mininal confirmation from sharders
	MinConfirmation int
	// CconfirmationChainLength minial confirmation chain length
	ConfirmationChainLength int

	// additional settings depending network latency
	// MaxTxnQuery maximum transcation query from sharders
	MaxTxnQuery int
	// QuerySleepTime sleep time before transcation query
	QuerySleepTime int

	// SignatureScheme signature scheme
	SignatureScheme string
	// ChainID which blockchain it is working
	ChainID string
}

// LoadConfigFile load and parse Config from file
func LoadConfigFile(file string) (Config, error) {

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

	return LoadConfig(v)
}

// LoadConfig load and parse config
func LoadConfig(v Reader) (Config, error) {

	var cfg Config

	blockWorker := strings.TrimSpace(v.GetString("block_worker"))

	if !isURL(blockWorker) {
		return cfg, thrown.Throw(ErrInvalidValue, "block_worker="+blockWorker)
	}

	minSubmit := v.GetInt("min_submit")
	if minSubmit < 1 {
		minSubmit = DefaultMinSubmit
	} else if minSubmit > 100 {
		minSubmit = 100
	}

	minCfm := v.GetInt("min_confirmation")

	if minCfm < 1 {
		minCfm = DefaultMinConfirmation
	} else if minCfm > 100 {
		minCfm = 100
	}

	CfmChainLength := v.GetInt("confirmation_chain_length")

	if CfmChainLength < 1 {
		CfmChainLength = DefaultConfirmationChainLength
	}

	// additional settings depending network latency
	maxTxnQuery := v.GetInt("max_txn_query")
	if maxTxnQuery < 1 {
		maxTxnQuery = DefaultMaxTxnQuery
	}

	querySleepTime := v.GetInt("query_sleep_time")
	if querySleepTime < 1 {
		querySleepTime = DefaultQuerySleepTime
	}

	cfg.BlockWorker = blockWorker
	cfg.PreferredBlobbers = v.GetStringSlice("preferred_blobbers")
	cfg.MinSubmit = minSubmit
	cfg.MinConfirmation = minCfm
	cfg.ConfirmationChainLength = CfmChainLength
	cfg.MaxTxnQuery = maxTxnQuery
	cfg.QuerySleepTime = querySleepTime

	cfg.SignatureScheme = v.GetString("signature_scheme")
	cfg.ChainID = v.GetString("chain_id")

	return cfg, nil

}

func isURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}
