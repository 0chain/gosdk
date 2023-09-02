package zcnswap

import (
	"errors"
	thrown "github.com/0chain/errors"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/sys"
	"github.com/spf13/viper"
	"os"
)

//var Configuration SwapConfig

type SwapClient struct {
	BancorAddress    string
	UsdcTokenAddress string
	ZcnTokenAddress  string
	WalletMnemonic   string
}

func CreateSwapClient(cfg *viper.Viper) *SwapClient {
	return &BridgeClient{
		BridgeAddress:      cfg.GetString("bridge.bridge_address"),
		TokenAddress:       cfg.GetString("bridge.token_address"),
		AuthorizersAddress: cfg.GetString("bridge.authorizers_address"),
		EthereumAddress:    cfg.GetString("bridge.ethereum_address"),
		Password:           cfg.GetString("bridge.password"),
		EthereumNodeURL:    cfg.GetString("ethereum_node_url"),
		GasLimit:           cfg.GetUint64("bridge.gas_limit"),
		ConsensusThreshold: cfg.GetFloat64("bridge.consensus_threshold"),
		Homedir:            homedir,
	}
}

//
//func Init(file string) error {
//	var err error
//	Configuration, err = loadConfigFile(file)
//
//	return err
//}

func SetWalletMnemonic(mnemonic string) {
	Configuration.WalletMnemonic = mnemonic
}

// LoadConfigFile load and parse Config from file
func loadConfigFile(file string) (SwapConfig, error) {

	var cfg SwapConfig
	var err error

	_, err = sys.Files.Stat(file)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, thrown.Throw(conf.ErrMssingConfig, file)
		}
		return cfg, err
	}

	v := viper.New()

	v.SetConfigFile(file)

	if err := v.ReadInConfig(); err != nil {
		return cfg, thrown.Throw(conf.ErrBadParsing, err.Error())
	}

	return loadConfig(v)
}

// LoadConfig load and parse config
func loadConfig(v conf.Reader) (SwapConfig, error) {

	var cfg SwapConfig

	cfg.UsdcTokenAddress = v.GetString("zcnswap.usdc_token_address")
	cfg.BancorAddress = v.GetString("zcnswap.bancor_address")
	cfg.ZcnTokenAddress = v.GetString("zcnswap.zcn_token_address")

	return cfg, nil
}
