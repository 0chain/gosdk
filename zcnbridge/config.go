package zcnbridge

import (
	"flag"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	ether "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
)

type BridgeSDKConfig struct {
	LogLevel    *string
	LogPath     *string
	ConfigFile  *string
	ConfigDir   *string
	Development *bool
}

type EthereumConfig struct {
	// URL of ethereum RPC node (infura or alchemy)
	EthereumNodeURL string
	// Address of Ethereum bridge contract
	BridgeAddress string
	// Address of WZCN Ethereum wrapped token
	WzcnAddress string

	// Ethereum chain ID
	ChainID string
	// Gas limit to execute ethereum transaction
	GasLimit uint64
	// Value to execute ZCN smart contracts
	Value int64
}

type BridgeOwnerConfig struct {
	EthereumConfig
	// Deployer of all bridge contracts
	EthereumMnemonic string
	// Address of Ethereum authorizers contract
	AuthorizersAddress string
}

// BridgeClientConfig initializes Ethereum zcnWallet and params
type BridgeClientConfig struct {
	EthereumConfig
	// Ethereum mnemonic (derivation of Ethereum owner, public and private key)
	EthereumMnemonic string
	// Authorizers required to confirm (in percents)
	ConsensusThreshold int
}

type Instance struct {
	zcnWallet *wallet.Wallet
	ethWallet *EthereumWallet
	startTime common.Timestamp
	nonce     int64
}

type BridgeClient struct {
	*BridgeClientConfig
	*Instance
}

type BridgeOwner struct {
	*BridgeOwnerConfig
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
		BridgeOwnerConfig: &BridgeOwnerConfig{
			EthereumConfig: EthereumConfig{
				EthereumNodeURL: cfg.GetString("owner.EthereumNodeURL"),
				// BridgeAddress:   cfg.GetString("owner.BridgeAddress"),
				WzcnAddress: cfg.GetString("owner.WzcnAddress"),
				ChainID:     cfg.GetString("owner.ChainID"),
				GasLimit:    cfg.GetUint64("owner.GasLimit"),
				Value:       cfg.GetInt64("owner.Value"),
			},
			EthereumMnemonic:   cfg.GetString("owner.OwnerEthereumMnemonic"),
			AuthorizersAddress: cfg.GetString("owner.AuthorizersAddress"),
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
			EthereumConfig: EthereumConfig{
				EthereumNodeURL: cfg.GetString("bridge.EthereumNodeURL"),
				BridgeAddress:   cfg.GetString("bridge.BridgeAddress"),
				WzcnAddress:     cfg.GetString("bridge.WzcnAddress"),
				ChainID:         cfg.GetString("bridge.ChainID"),
				GasLimit:        cfg.GetUint64("bridge.GasLimit"),
				Value:           cfg.GetInt64("bridge.Value"),
			},
			EthereumMnemonic:   cfg.GetString("bridge.ClientEthereumMnemonic"),
			ConsensusThreshold: cfg.GetInt("bridge.ConsensusThreshold"),
		},
		Instance: &Instance{
			startTime: common.Now(),
		},
	}
}

// GetClientEthereumAddress returns ethereum zcnWallet string
func (b *BridgeClient) GetClientEthereumAddress() ether.Address {
	return b.ethWallet.Address
}

// GetClientEthereumWallet returns ethereum zcnWallet string
func (b *BridgeClient) GetClientEthereumWallet() *EthereumWallet {
	return b.ethWallet
}

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
	bridgeClient.SetupEthereumWallet()

	return bridgeClient
}

// SetupBridgeOwnerSDK Use this from standalone application
func SetupBridgeOwnerSDK(cfg *BridgeSDKConfig) *BridgeOwner {
	bridgeOwner := CreateBridgeOwner(readSDKConfig(cfg))
	bridgeOwner.SetupEthereumWallet()

	return bridgeOwner
}

// GetEthereumWallet returns owner ethereum zcnWallet
func (b *BridgeOwner) GetEthereumWallet() *EthereumWallet {
	return b.ethWallet
}
