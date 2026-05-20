package conf

import (
	"errors"
	"regexp"
	"strings"
	"sync"
)

var (
	//  global client config
	cfg     *Config
	onceCfg sync.Once
	//  global sharders and miners
	network *Network
	// clientSecureTransport, when true, rewrites mainnet sharder URLs from
	// http://<host>:7171 to https://<host>/sharder01 so clients running under
	// secure-transport restrictions can reach them: browser wasm (mixed-content
	// blocking) and mobile (iOS App Transport Security / Android cleartext
	// blocked by default). Server-side consumers (zs3 gateway, zboxcli) leave
	// this off and keep using the direct http endpoints. Set via
	// SetClientSecureTransport from the wasm/mobile SDK init paths.
	clientSecureTransport bool
)

// sharderHTTPS rewrites http://<host>:7171 -> https://<host>/sharder01 (the
// mainnet sharders' TLS reverse-proxy path).
var sharderHTTPSRewriteRe = regexp.MustCompile(`^http://([^:/]+):7171$`)

// SetClientSecureTransport toggles the sharder http->https rewrite. Enable it
// from browser (wasm) and mobile SDK init; leave off for server-side SDK use.
func SetClientSecureTransport(enabled bool) {
	clientSecureTransport = enabled
}

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

// GetClientConfig get global client config from the SDK configuration
func GetClientConfig() (*Config, error) {
	if cfg == nil {
		return nil, ErrConfigNotInitialized
	}

	return cfg, nil
}

// InitClientConfig set global client SDK config
func InitClientConfig(c *Config) {
	onceCfg.Do(func() {
		sharderConsensous := c.SharderConsensous
		if sharderConsensous < 1 {
			sharderConsensous = DefaultSharderConsensous
		}
		cfg = c
		cfg.SharderConsensous = sharderConsensous
	})
}

// InitChainNetwork set global chain network for the SDK given its configuration
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
		if clientSecureTransport {
			network.Sharders[i] = sharderHTTPSRewriteRe.ReplaceAllString(
				network.Sharders[i], "https://$1/sharder01")
		}
	}
}
