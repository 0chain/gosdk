package conf

import (
	"errors"
	"strings"
)

var (
	//  global client config
	cfg *Config
	//  global sharders and miners
	network *Network
)

var (
	//ErrNilConfig config is nil
	ErrNilConfig = errors.New("[conf]config is nil")

	// ErrMssingConfig config file is missing
	ErrMssingConfig = errors.New("[conf]missing config file")
	// ErrInvalidValue invalid value in config
	ErrInvalidValue = errors.New("[conf]invalid value")
	// ErrBadParsing fail to parse config via spf13/viper
	ErrBadParsing = errors.New("[conf]bad parsing")

	// ErrConfigNotInitialized config is not initialized
	ErrConfigNotInitialized = errors.New("[conf]conf.cfg is not initialized. please initialize it by conf.InitClientConfig")
)

// GetClientConfig get global client config
func GetClientConfig() (*Config, error) {
	if cfg == nil {
		return nil, ErrConfigNotInitialized
	}

	return cfg, nil
}

// InitClientConfig set global client config
func InitClientConfig(c *Config) {
	cfg = c
}

// InitChainNetwork set global chain network
func InitChainNetwork(n *Network) {
	if n == nil {
		return
	}

	normalizeURLs(n)

	if network == nil {
		network = n
		return
	}

	network.Sharders = n.Sharders
	network.Miners = n.Miners
}

func normalizeURLs(network *Network) {
	if network == nil {
		return
	}

	for i := 0; i < len(network.Miners); i++ {
		network.Miners[i] = strings.TrimSuffix(network.Miners[i], "/")
	}

	for i := 0; i < len(network.Sharders); i++ {
		network.Sharders[i] = strings.TrimSuffix(network.Sharders[i], "/")
	}
}
