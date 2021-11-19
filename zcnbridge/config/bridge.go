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
