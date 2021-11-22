package config

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
}

var (
	Bridge BridgeConfig
)

//// InitBridge Sets up the wallet and node
//// Wallet setup reads keys from keyfile and registers in the 0chain
//func InitBridge() {
//	err := wallet.SetupSDK(config.GetSDKConfig())
//	if err != nil {
//		log.Logger.Fatal("failed to setup ZCNSDK", zap.Error(err))
//	}
//
//	walletConfig, err := wallet.SetupZCNWallet()
//	if err != nil {
//		log.Logger.Fatal("failed to setup wallet", zap.Error(err))
//	}
//
//	ethWalletConfig, err := wallet.SetupEthereumWallet()
//	if err != nil {
//		log.Logger.Fatal("failed to setup ethereum wallet", zap.Error(err))
//	}
//
//	node.Start(walletConfig, ethWalletConfig)
//}
