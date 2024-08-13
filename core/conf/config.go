// Provides the data structures and methods to work with the configuration data structure.
// This includes parsing, loading, and saving the configuration data structure.
// It uses the viper library to parse and manage the configuration data structure.
package conf

import (
	"errors"
	"net/url"
	"os"
	"strings"

	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/core/sys"
	"github.com/spf13/viper"
)

const (
	// DefaultMinSubmit default value for min_submit
	DefaultMinSubmit = 10
	// DefaultMinConfirmation default value for min_confirmation
	DefaultMinConfirmation = 10
	// DefaultMaxTxnQuery default value for max_txn_query
	DefaultMaxTxnQuery = 5
	// DefaultConfirmationChainLength default value for confirmation_chain_length
	DefaultConfirmationChainLength = 3
	// DefaultQuerySleepTime default value for query_sleep_time
	DefaultQuerySleepTime = 5
	// DefaultSharderConsensous default consensous to take make SCRestAPI calls
	DefaultSharderConsensous = 3
)

// Config settings from ~/.zcn/config.yaml
//
//	block_worker: http://198.18.0.98:9091
//	signature_scheme: bls0chain
//	min_submit: 50
//	min_confirmation: 50
//	confirmation_chain_length: 3
//	max_txn_query: 5
//	query_sleep_time: 5
//	# # OPTIONAL - Uncomment to use/ Add more if you want
//	# preferred_blobbers:
//	#   - http://one.devnet-0chain.net:31051
//	#   - http://one.devnet-0chain.net:31052
//	#   - http://one.devnet-0chain.net:31053
type Config struct {
	// BlockWorker the url of 0dns's network api
	BlockWorker string `json:"block_worker,omitempty"`
	// PreferredBlobbers preferred blobbers on new allocation
	PreferredBlobbers []string `json:"preferred_blobbers,omitempty"`

	// MinSubmit mininal submit from blobber
	MinSubmit int `json:"min_submit,omitempty"`
	// MinConfirmation mininal confirmation from sharders
	MinConfirmation int `json:"min_confirmation,omitempty"`
	// CconfirmationChainLength minial confirmation chain length
	ConfirmationChainLength int `json:"confirmation_chain_length,omitempty"`

	// additional settings depending network latency
	// MaxTxnQuery maximum transcation query from sharders
	MaxTxnQuery int `json:"max_txn_query,omitempty"`
	// QuerySleepTime sleep time before transcation query
	QuerySleepTime int `json:"query_sleep_time,omitempty"`

	// SignatureScheme signature scheme
	SignatureScheme string `json:"signature_scheme,omitempty"`
	// ChainID which blockchain it is working
	ChainID string `json:"chain_id,omitempty"`

	VerifyOptimistic bool

	// Ethereum node: "https://ropsten.infura.io/v3/xxxxxxxxxxxxxxx"
	EthereumNode string `json:"ethereum_node,omitempty"`

	// ZboxHost 0box api host host: "https://0box.dev.0chain.net"
	ZboxHost string `json:"zbox_host"`
	// ZboxAppType app type name
	ZboxAppType string `json:"zbox_app_type"`
	// SharderConsensous is consensous for when quering for SCRestAPI calls
	SharderConsensous int `json:"sharder_consensous"`
}

// LoadConfigFile load and parse SDK Config from file
//   - file: config file path (full path)
func LoadConfigFile(file string) (Config, error) {

	var cfg Config
	var err error

	_, err = sys.Files.Stat(file)

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
	VerifyOptimisticString := v.GetString("verify_optimistic")
	if VerifyOptimisticString == "true" {
		cfg.VerifyOptimistic = true
	}

	sharderConsensous := v.GetInt("sharder_consensous")
	if sharderConsensous < 1 {
		sharderConsensous = DefaultSharderConsensous
	}

	cfg.BlockWorker = blockWorker
	cfg.PreferredBlobbers = v.GetStringSlice("preferred_blobbers")
	cfg.MinSubmit = minSubmit
	cfg.MinConfirmation = minCfm
	cfg.ConfirmationChainLength = CfmChainLength
	cfg.MaxTxnQuery = maxTxnQuery
	cfg.QuerySleepTime = querySleepTime
	cfg.SharderConsensous = sharderConsensous

	cfg.SignatureScheme = v.GetString("signature_scheme")
	cfg.ChainID = v.GetString("chain_id")

	return cfg, nil

}

func isURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}
