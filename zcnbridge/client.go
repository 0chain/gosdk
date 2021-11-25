package zcnbridge

import (
	"flag"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zcnbridge/chain"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	ether "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
)

type ClientConfig struct {
	WalletFileConfig *string
	LogPath          *string
	ConfigFile       *string
	ConfigDir        *string
	Development      *bool
}

// BridgeConfig initializes Ethereum wallet and params
type BridgeConfig struct {
	// Ethereum mnemonic (derivation of Ethereum owner, public and private key)
	Mnemonic string
	// Address of Ethereum bridge contract
	BridgeAddress string
	// Address of WZCN Ethereum wrapped token
	WzcnAddress string
	// URL of ethereum RPC node (infura or alchemy)
	EthereumNodeURL string
	// Ethereum chain ID
	ChainID string
	// Gas limit to execute ethereum transaction
	GasLimit uint64
	// Value to execute ZCN smart contracts
	Value int64
	// Authorizers required to confirm (in percents)
	ConsensusThreshold int
}

type Instance struct {
	wallet         *wallet.Wallet
	ethereumWallet *EthereumWallet
	startTime      common.Timestamp
	nonce          int64
}

type Bridge struct {
	BridgeConfig
	Instance
}

func (c ClientConfig) LogDir() string {
	return *c.LogPath
}

func (c ClientConfig) LogLvl() string {
	return viper.GetString("logging.level")
}

func (c ClientConfig) BlockWorker() string {
	return chain.GetServerChain().BlockWorker
}

func (c ClientConfig) SignatureScheme() string {
	return chain.GetServerChain().SignatureScheme
}

func (c ClientConfig) WalletFile() string {
	return *c.WalletFileConfig
}

// ReadClientConfigFromCmd reads config from command line
func ReadClientConfigFromCmd() *ClientConfig {
	cmd := &ClientConfig{}
	cmd.Development = flag.Bool("development", true, "development mode")
	cmd.WalletFileConfig = flag.String("wallet_config", "wallet.json", "wallet config")
	cmd.LogPath = flag.String("log_dir", "./logs", "log folder")
	cmd.ConfigDir = flag.String("config_dir", "./config", "config folder")
	cmd.ConfigFile = flag.String("config_file", "bridge", "config file")

	flag.Parse()

	return cmd
}

func SetupBridgeFromConfig() *Bridge {
	return &Bridge{
		BridgeConfig: BridgeConfig{
			Mnemonic:           viper.GetString("bridge.Mnemonic"),
			BridgeAddress:      viper.GetString("bridge.BridgeAddress"),
			WzcnAddress:        viper.GetString("bridge.WzcnAddress"),
			EthereumNodeURL:    viper.GetString("bridge.EthereumNodeURL"),
			ChainID:            viper.GetString("bridge.ChainID"),
			GasLimit:           viper.GetUint64("bridge.GasLimit"),
			Value:              viper.GetInt64("bridge.Value"),
			ConsensusThreshold: viper.GetInt("bridge.ConsensusThreshold"),
		},
	}
}

// GetEthereumAddress returns ethereum wallet string
func (b *Bridge) GetEthereumAddress() ether.Address {
	return b.ethereumWallet.Address
}

// GetEthereumWallet returns ethereum wallet string
func (b *Bridge) GetEthereumWallet() *EthereumWallet {
	return b.ethereumWallet
}

// ID returns id of Node.
func (b *Bridge) ID() string {
	return b.wallet.ID()
}

func (b *Bridge) IncrementNonce() int64 {
	b.nonce++
	return b.nonce
}
