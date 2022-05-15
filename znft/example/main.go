package main

import (
	"context"

	"github.com/0chain/gosdk/znft"
)

func main() {
	cfg := &znft.Configuration{
		FactoryAddress:                   "",
		FactoryModuleERC721Address:       "",
		FactoryModuleERC721FixedAddress:  "",
		FactoryModuleERC721RandomAddress: "",
		EthereumNodeURL:                  "",
		WalletAddress:                    "",
		VaultPassword:                    "",
		Homedir:                          "",
		GasLimit:                         0,
		Value:                            0,
	}

	tr := cfg.CreateSignedTransactionFromKeyStore(context.Background(), 0)
	_ = tr
}
