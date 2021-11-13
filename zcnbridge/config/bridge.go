package config

type BridgeConfig struct {
	Mnemonic string // owner Mnemonic
	// ownerAddress  common.Address
	// publicKey     crypto.PublicKey
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
