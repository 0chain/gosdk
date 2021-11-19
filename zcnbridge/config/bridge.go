package config

// BridgeConfig TODO: some of parameters are not initialized
type BridgeConfig struct {
	// Ethereum mnemonic (derivation of Ethereum owner, public and private key)
	Mnemonic string
	// Address of Ethereum bridge contract
	BridgeAddress string
	// Address of WZCN Ethereum wrapped token
	WzcnAddress string
	// URL of ethereum RPC node (infura or alchemy)
	EthereumNodeURL string
	// Gas limit to execute ethereum transaction
	GasLimit uint64
	// Value to execute ZCN smart contracts
	Value int64
}

var (
	Bridge BridgeConfig
)
