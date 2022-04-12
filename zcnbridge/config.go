package zcnbridge

import (
	"flag"
	"fmt"
	"path"

	"github.com/0chain/gosdk/zcncore"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/spf13/viper"
)

type BridgeSDKConfig struct {
	LogLevel         *string
	LogPath          *string
	ConfigBridgeFile *string
	ConfigChainFile  *string
	ConfigDir        *string
	Development      *bool
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
	ConsensusThreshold float64
}

type EthereumConfig struct {
	// URL of ethereum RPC node (infura or alchemy)
	EthereumNodeURL string
	// Gas limit to execute ethereum transaction
	GasLimit uint64
	// Value to execute Ethereum smart contracts (default = 0)
	Value int64
}

type BridgeClientConfig struct {
	ContractsRegistry
	EthereumConfig
	EthereumAddress string
	Password        string
	Homedir         string
}

type Instance struct {
	//zcnWallet *wallet.Wallet
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
	cmd.Development = flag.Bool("development", false, "development mode")
	cmd.LogPath = flag.String("logs", "./logs", "log folder")
	cmd.ConfigDir = flag.String("path", GetConfigDir(), "config home folder")
	cmd.ConfigBridgeFile = flag.String("bridge_config", BridgeClientConfigName, "bridge config file")
	cmd.ConfigChainFile = flag.String("chain_config", ZChainsClientConfigName, "chain config file")
	cmd.LogLevel = flag.String("loglevel", "debug", "log level")

	flag.Parse()

	return cmd
}

func CreateBridgeOwner(cfg *viper.Viper) *BridgeOwner {
	owner := cfg.Get(OwnerConfigKeyName)
	if owner == nil {
		ExitWithError("CreateBridgeOwner: can't read config with `owner` key")
	}

	fileUsed := cfg.ConfigFileUsed()
	homedir := path.Dir(fileUsed)
	if homedir == "" {
		ExitWithError("CreateBridgeOwner: homedir is required")
	}

	return &BridgeOwner{
		BridgeClientConfig: &BridgeClientConfig{
			ContractsRegistry: ContractsRegistry{
				BridgeAddress:      cfg.GetString(fmt.Sprintf("%s.BridgeAddress", OwnerConfigKeyName)),
				WzcnAddress:        cfg.GetString(fmt.Sprintf("%s.WzcnAddress", OwnerConfigKeyName)),
				AuthorizersAddress: cfg.GetString(fmt.Sprintf("%s.AuthorizersAddress", OwnerConfigKeyName)),
			},
			EthereumConfig: EthereumConfig{
				EthereumNodeURL: cfg.GetString(fmt.Sprintf("%s.EthereumNodeURL", OwnerConfigKeyName)),
				GasLimit:        cfg.GetUint64(fmt.Sprintf("%s.GasLimit", OwnerConfigKeyName)),
				Value:           cfg.GetInt64(fmt.Sprintf("%s.Value", OwnerConfigKeyName)),
			},
			EthereumAddress: cfg.GetString(fmt.Sprintf("%s.EthereumAddress", OwnerConfigKeyName)),
			Password:        cfg.GetString(fmt.Sprintf("%s.Password", OwnerConfigKeyName)),
			Homedir:         homedir,
		},
		Instance: &Instance{
			startTime: common.Now(),
		},
	}
}

func CreateBridgeClient(cfg *viper.Viper) *BridgeClient {
	fileUsed := cfg.ConfigFileUsed()
	homedir := path.Dir(fileUsed)
	if homedir == "" {
		ExitWithError("homedir is required")
	}

	bridge := cfg.Get(ClientConfigKeyName)
	if bridge == nil {
		ExitWithError(fmt.Sprintf("Can't read config with '%s' key", ClientConfigKeyName))
	}

	return &BridgeClient{
		BridgeClientConfig: &BridgeClientConfig{
			ContractsRegistry: ContractsRegistry{
				BridgeAddress:      cfg.GetString(fmt.Sprintf("%s.BridgeAddress", ClientConfigKeyName)),
				WzcnAddress:        cfg.GetString(fmt.Sprintf("%s.WzcnAddress", ClientConfigKeyName)),
				AuthorizersAddress: cfg.GetString(fmt.Sprintf("%s.AuthorizersAddress", ClientConfigKeyName)),
			},
			EthereumConfig: EthereumConfig{
				EthereumNodeURL: cfg.GetString(fmt.Sprintf("%s.EthereumNodeURL", ClientConfigKeyName)),
				GasLimit:        cfg.GetUint64(fmt.Sprintf("%s.GasLimit", ClientConfigKeyName)),
				Value:           cfg.GetInt64(fmt.Sprintf("%s.Value", ClientConfigKeyName)),
			},
			EthereumAddress: cfg.GetString(fmt.Sprintf("%s.EthereumAddress", ClientConfigKeyName)),
			Password:        cfg.GetString(fmt.Sprintf("%s.Password", ClientConfigKeyName)),
			Homedir:         homedir,
		},
		BridgeConfig: &BridgeConfig{
			ConsensusThreshold: cfg.GetFloat64(fmt.Sprintf("%s.ConsensusThreshold", ClientConfigKeyName)),
		},
		Instance: &Instance{
			startTime: common.Now(),
		},
	}
}

// ID returns id of Node.
func (b *BridgeClient) ID() string {
	return zcncore.GetClientWalletID()
}

// ID returns id of Node.
func (b *BridgeOwner) ID() string {
	return zcncore.GetClientWalletID()
}

func (b *BridgeClient) IncrementNonce() int64 {
	b.nonce++
	return b.nonce
}

// SetupBridgeClientSDK Use this from standalone application
// 0Chain SDK initialization is required
func SetupBridgeClientSDK(cfg *BridgeSDKConfig) *BridgeClient {
	log.InitLogging(*cfg.Development, *cfg.LogPath, *cfg.LogLevel)

	bridgeClient := CreateBridgeClient(initBridgeConfig(cfg))

	return bridgeClient
}

// SetupBridgeOwnerSDK Use this from standalone application to initialize bridge owner.
// 0Chain SDK initialization is not required in this case
func SetupBridgeOwnerSDK(cfg *BridgeSDKConfig) *BridgeOwner {
	bridgeOwner := CreateBridgeOwner(initBridgeConfig(cfg))

	return bridgeOwner
}
