package config

type BridgeConfig struct {
	Mnemonic        string // Ethereum mnemonic
	BridgeAddress   string
	WzcnAddress     string
	EthereumNodeURL string
	ChainID         int
	GasLimit        int
	Value           int
}

var (
	Bridge BridgeConfig
)
