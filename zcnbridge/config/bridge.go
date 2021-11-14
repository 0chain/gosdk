package config

// BridgeConfig TODO: some of parameters are not initialized
type BridgeConfig struct {
	// Ethereum mnemonic
	Mnemonic string
	// Address of bridge contract
	BridgeAddress string
	// Address of WZCN wrapper token
	WzcnAddress string
	// URL of ethereum RPC node (infura or alchemy)
	EthereumNodeURL string
	// Owner ethereum wallet address
	EthereumAddress string
	// Chain ID (Ropsten, RinkeBy) // TODO: add description and initialization
	ChainID int
	// Gas limit to execute ethereum transaction
	GasLimit int
	// Value to execute ZCN smart contracts
	Value int64
}

var (
	Bridge BridgeConfig
)
