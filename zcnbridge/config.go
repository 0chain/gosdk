package zcnbridge

import (
	"flag"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/spf13/viper"
)

type BridgeSDKConfig struct {
	LogLevel    *string
	LogPath     *string
	ConfigFile  *string
	ConfigDir   *string
	Development *bool
}

type ContractsRegistry struct {
	// Address of Ethereum bridge contract
	BridgeAddress string
	// Address of WZCN Ethereum wrapped token
	WzcnAddress string
	// Address of Ethereum authorizers contract
	AuthorizersAddress string
}

type BridgeConfig struct {
	ConsensusThreshold int
}

type EthereumConfig struct {
	// URL of ethereum RPC node (infura or alchemy)
	EthereumNodeURL string
	// Ethereum chain ID
	ChainID string
	// Gas limit to execute ethereum transaction
	GasLimit uint64
	// Value to execute ZCN smart contracts
	Value int64
}

type BridgeClientConfig struct {
	ContractsRegistry
	EthereumConfig
	Address  string
	Password string
}

type Instance struct {
	zcnWallet *wallet.Wallet
	startTime common.Timestamp
	nonce     int64
}

type BridgeClient struct {
	*BridgeConfig
	*BridgeClientConfig
	*Instance
}

type BridgeOwner struct {
	*BridgeClientConfig
	*Instance
}

// ReadClientConfigFromCmd reads config from command line
// Bridge has several configs:
// Chain config at ~/.zcn/config.json
// User 0Chain wallet config at ~/.zcn/wallet.json
// User EthBridge config ~/.zcn/bridge.json
// Owner EthBridge config ~/.zcn/bridgeowner.json
func ReadClientConfigFromCmd() *BridgeSDKConfig {
	// reading from bridge.yaml
	cmd := &BridgeSDKConfig{}
	cmd.Development = flag.Bool("development", true, "development mode")
	cmd.LogPath = flag.String("log_dir", "./logs", "log folder")
	cmd.ConfigDir = flag.String("config_dir", "./config", "config folder")
	cmd.ConfigFile = flag.String("config_file", "bridge", "config file")
	cmd.LogLevel = flag.String("loglevel", "debug", "log level")

	flag.Parse()

	return cmd
}

func CreateBridgeOwner(cfg *viper.Viper) *BridgeOwner {
	owner := cfg.Get("owner")
	if owner == nil {
		ExitWithError("Can't read config with `owner` key")
	}

	return &BridgeOwner{
		BridgeClientConfig: &BridgeClientConfig{
			ContractsRegistry: ContractsRegistry{
				BridgeAddress:      cfg.GetString("owner.BridgeAddress"),
				WzcnAddress:        cfg.GetString("owner.WzcnAddress"),
				AuthorizersAddress: cfg.GetString("owner.AuthorizersAddress"),
			},
			EthereumConfig: EthereumConfig{
				EthereumNodeURL: cfg.GetString("owner.EthereumNodeURL"),
				ChainID:         cfg.GetString("owner.ChainID"),
				GasLimit:        cfg.GetUint64("owner.GasLimit"),
				Value:           cfg.GetInt64("owner.Value"),
			},
			Address:  cfg.GetString("owner.address"),
			Password: cfg.GetString("owner.password"),
		},
		Instance: &Instance{
			startTime: common.Now(),
		},
	}
}

func CreateBridgeClient(cfg *viper.Viper) *BridgeClient {
	bridge := cfg.Get("bridge")
	if bridge == nil {
		ExitWithError("Can't read config with `bridge` key")
	}

	return &BridgeClient{
		BridgeClientConfig: &BridgeClientConfig{
			ContractsRegistry: ContractsRegistry{
				BridgeAddress:      cfg.GetString("bridge.BridgeAddress"),
				WzcnAddress:        cfg.GetString("bridge.WzcnAddress"),
				AuthorizersAddress: cfg.GetString("bridge.AuthorizersAddress"),
			},
			EthereumConfig: EthereumConfig{
				EthereumNodeURL: cfg.GetString("bridge.EthereumNodeURL"),
				ChainID:         cfg.GetString("bridge.ChainID"),
				GasLimit:        cfg.GetUint64("bridge.GasLimit"),
				Value:           cfg.GetInt64("bridge.Value"),
			},
			Address: cfg.GetString("bridge.address"),
		},
		BridgeConfig: &BridgeConfig{
			ConsensusThreshold: cfg.GetInt("bridge.ConsensusThreshold"),
		},
		Instance: &Instance{
			startTime: common.Now(),
		},
	}
}

//// GetClientEthereumAddress returns ethereum zcnWallet string
//func (b *BridgeClient) GetClientEthereumAddress() ether.Address {
//	return b.ethWallet.Address
//}

//// GetClientEthereumWallet returns ethereum zcnWallet string
//func (b *BridgeClient) GetClientEthereumWallet() *EthereumWallet {
//	return b.ethWallet
//}

// ID returns id of Node.
func (b *BridgeClient) ID() string {
	return b.zcnWallet.ID()
}

// ID returns id of Node.
func (b *BridgeOwner) ID() string {
	return b.zcnWallet.ID()
}

// PublicKey returns public key of Node
func (b *BridgeClient) PublicKey() string {
	return b.zcnWallet.PublicKey()
}

func (b *BridgeClient) PrivateKey() string {
	return b.zcnWallet.PrivateKey()
}

func (b *BridgeClient) IncrementNonce() int64 {
	b.nonce++
	return b.nonce
}

// SetupBridgeClientSDK Use this from standalone application
func SetupBridgeClientSDK(cfg *BridgeSDKConfig) *BridgeClient {
	initChainFromConfig("config.yaml")

	bridgeClient := CreateBridgeClient(readSDKConfig(cfg))
	bridgeClient.SetupZCNSDK(*cfg.LogPath, *cfg.LogLevel)
	bridgeClient.SetupZCNWallet("wallet.json")
	//bridgeClient.SetupEthereumWallet()

	return bridgeClient
}

// SetupBridgeOwnerSDK Use this from standalone application
func SetupBridgeOwnerSDK(cfg *BridgeSDKConfig) *BridgeOwner {
	bridgeOwner := CreateBridgeOwner(readSDKConfig(cfg))
	//bridgeOwner.SetupEthereumWallet()

	return bridgeOwner
}

//// GetEthereumWallet returns owner ethereum zcnWallet
//func (b *BridgeOwner) GetEthereumWallet() *EthereumWallet {
//	return b.ethWallet
//}
