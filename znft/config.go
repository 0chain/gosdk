package znft

const (
	ConfigFile   = "config.yaml"
	WalletConfig = "wallet.json"
	WalletDir    = "wallets"
)

type Configuration struct {
	FactoryAddress                   string // FactoryAddress address
	FactoryModuleERC721Address       string // FactoryModuleERC721Address address
	FactoryModuleERC721FixedAddress  string // FactoryModuleERC721FixedAddress address
	FactoryModuleERC721RandomAddress string // FactoryModuleERC721RandomAddress address
	EthereumNodeURL                  string // EthereumNodeURL URL of ethereum RPC node (infura or alchemy)
	WalletAddress                    string // WalletAddress client address
	VaultPassword                    string // VaultPassword used to sign transactions on behalf of the client
	Homedir                          string // Homedir is a client config folder
	GasLimit                         uint64 // GasLimit limit to execute ethereum transaction
	Value                            int64  // Value to execute Ethereum smart contracts (default = 0)
}
