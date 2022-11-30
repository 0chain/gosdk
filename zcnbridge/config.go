package zcnbridge

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/zcncrypto"
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
	zcnWallet *zcncrypto.Wallet
	startTime common.Timestamp
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

func CreateBridgeOwner(cfg *viper.Viper, walletFile ...string) *BridgeOwner {
	owner := cfg.Get(OwnerConfigKeyName)
	if owner == nil {
		ExitWithError("CreateBridgeOwner: can't read config with `owner` key")
	}

	fileUsed := cfg.ConfigFileUsed()
	homedir := path.Dir(fileUsed)
	if homedir == "" {
		ExitWithError("CreateBridgeOwner: homedir is required")
	}

	wallet, err := loadWallet(homedir, walletFile...)
	if err != nil {
		ExitWithError("Error reading the wallet", err)
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
			zcnWallet: wallet,
		},
	}
}

func CreateBridgeClient(cfg *viper.Viper, walletFile ...string) *BridgeClient {
	fileUsed := cfg.ConfigFileUsed()
	homedir := path.Dir(fileUsed)
	if homedir == "" {
		ExitWithError("homedir is required")
	}

	bridge := cfg.Get(ClientConfigKeyName)
	if bridge == nil {
		ExitWithError(fmt.Sprintf("Can't read config with '%s' key", ClientConfigKeyName))
	}

	wallet, err := loadWallet(homedir, walletFile...)
	if err != nil {
		ExitWithError(err)
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
			zcnWallet: wallet,
		},
	}
}

type BridgeClientYaml struct {
	Password           string
	EthereumAddress    string
	BridgeAddress      string
	AuthorizersAddress string
	WzcnAddress        string
	EthereumNodeURL    string
	GasLimit           uint64
	Value              int64
	ConsensusThreshold float64
}

func CreateBridgeClientWithConfig(cfg BridgeClientYaml, wallet *zcncrypto.Wallet) *BridgeClient {
	return &BridgeClient{
		BridgeClientConfig: &BridgeClientConfig{
			ContractsRegistry: ContractsRegistry{
				BridgeAddress:      cfg.BridgeAddress,
				WzcnAddress:        cfg.WzcnAddress,
				AuthorizersAddress: cfg.AuthorizersAddress,
			},
			EthereumConfig: EthereumConfig{
				EthereumNodeURL: cfg.EthereumNodeURL,
				GasLimit:        cfg.GasLimit,
				Value:           cfg.Value,
			},
			EthereumAddress: cfg.EthereumAddress,
			Password:        cfg.Password,
			Homedir:         ".",
		},
		BridgeConfig: &BridgeConfig{
			ConsensusThreshold: cfg.ConsensusThreshold,
		},
		Instance: &Instance{
			startTime: common.Now(),
			zcnWallet: wallet,
		},
	}
}

func (b *BridgeClient) ClientID() string {
	return b.zcnWallet.ClientID
}

func (b *BridgeOwner) ClientID() string {
	return b.zcnWallet.ClientID
}

// SetupBridgeClientSDK Use this from standalone application
// 0Chain SDK initialization is required
func SetupBridgeClientSDK(cfg *BridgeSDKConfig, walletFile ...string) *BridgeClient {
	log.InitLogging(*cfg.Development, *cfg.LogPath, *cfg.LogLevel)
	bridgeClient := CreateBridgeClient(initBridgeConfig(cfg), walletFile...)
	return bridgeClient
}

// SetupBridgeOwnerSDK Use this from standalone application to initialize bridge owner.
// 0Chain SDK initialization is not required in this case
func SetupBridgeOwnerSDK(cfg *BridgeSDKConfig, walletFile ...string) *BridgeOwner {
	log.InitLogging(*cfg.Development, *cfg.LogPath, *cfg.LogLevel)
	bridgeOwner := CreateBridgeOwner(initBridgeConfig(cfg), walletFile...)
	return bridgeOwner
}

func loadWallet(homedir string, fileName ...string) (*zcncrypto.Wallet, error) {
	var walletPath string

	if len(fileName) != 0 {
		walletPath = path.Join(homedir, fileName[0])
	} else {
		walletPath = path.Join(homedir, ZChainWalletConfigName)
	}

	clientBytes, err := os.ReadFile(walletPath)
	if err != nil {
		ExitWithError("Error reading the wallet", err)
	}

	wallet := &zcncrypto.Wallet{}
	err = json.Unmarshal(clientBytes, &wallet)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}
