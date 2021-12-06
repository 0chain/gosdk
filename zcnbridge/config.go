package zcnbridge

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/zcnbridge/chain"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	ether "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
)

type ClientConfig struct {
	LogPath     *string
	ConfigFile  *string
	ConfigDir   *string
	Development *bool
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
	*BridgeConfig
	*Instance
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

// ReadClientConfigFromCmd reads config from command line
func ReadClientConfigFromCmd() *ClientConfig {
	cmd := &ClientConfig{}
	cmd.Development = flag.Bool("development", true, "development mode")
	cmd.LogPath = flag.String("log_dir", "./logs", "log folder")
	cmd.ConfigDir = flag.String("config_dir", "./config", "config folder")
	cmd.ConfigFile = flag.String("config_file", "bridge", "config file")

	flag.Parse()

	return cmd
}

func SetupBridgeFromConfig() *Bridge {
	return &Bridge{
		BridgeConfig: &BridgeConfig{
			Mnemonic:           viper.GetString("bridge.Mnemonic"),
			BridgeAddress:      viper.GetString("bridge.BridgeAddress"),
			WzcnAddress:        viper.GetString("bridge.WzcnAddress"),
			EthereumNodeURL:    viper.GetString("bridge.EthereumNodeURL"),
			ChainID:            viper.GetString("bridge.ChainID"),
			GasLimit:           viper.GetUint64("bridge.GasLimit"),
			Value:              viper.GetInt64("bridge.Value"),
			ConsensusThreshold: viper.GetInt("bridge.ConsensusThreshold"),
		},
		Instance: &Instance{
			startTime: common.Now(),
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

// PublicKey returns public key of Node
func (b *Bridge) PublicKey() string {
	return b.wallet.PublicKey()
}

func (b *Bridge) PrivateKey() string {
	return b.wallet.PrivateKey()
}

func (b *Bridge) IncrementNonce() int64 {
	b.nonce++
	return b.nonce
}

func getConfigDir() string {
	var configDir string
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	configDir = home + "/.zcn"
	return configDir
}

func initChain() {
	nodeConfig := viper.New()
	configDir := getConfigDir()
	nodeConfig.AddConfigPath(configDir)
	nodeConfig.SetConfigFile(path.Join(configDir, "config.yaml"))

	if err := nodeConfig.ReadInConfig(); err != nil {
		ExitWithError("Can't read config:", err)
	}

	InitChainFromConfig(nodeConfig)
}

func restoreChain() {
	config, err := conf.GetClientConfig()
	if err != nil {
		ExitWithError("Can't read config:", err)
	}

	RestoreFromConfig(config)
}

// SetupBridge Use this from standalone application
func SetupBridge(configDir, configFile string, development bool, logPath string) *Bridge {
	setDefaults()
	ReadConfig(configDir, configFile)
	setupLogging(development, logPath)
	bridge := SetupBridgeFromConfig()

	return bridge
}

func setDefaults() {
	viper.SetDefault("bridge.loglevel", "info")
}

func ReadConfig(configPath, configName string) {
	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func setupLogging(development bool, logPath string) {
	log.InitLogging(development, logPath, viper.GetString("bridge.loglevel"))
}

func RestoreFromConfig(cfg *conf.Config) {
	chain.SetServerChain(chain.NewChain(
		cfg.BlockWorker,
		cfg.SignatureScheme,
		cfg.MinSubmit,
		cfg.MinConfirmation,
	))
}

func InitChainFromConfig(reader conf.Reader) {
	chain.SetServerChain(chain.NewChain(
		reader.GetString("block_worker"),
		reader.GetString("signature_scheme"),
		reader.GetInt("min_submit"),
		reader.GetInt("min_confirmation"),
	))
}

func ExitWithError(v ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}
