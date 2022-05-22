package main

import (
	"context"
	"log"

	"github.com/0chain/gosdk/znft"
)

func main() {
	// Create this configuration
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

	// Create NFT with factory
	address := "beef"

	// Create session of NFT token
	session, err := cfg.CreateStorageERC721Session(context.Background(), address)
	if err != nil {
		panic(err)
	}

	// Start using session
	price, err := session.Price()
	if err != nil {
		return
	}

	log.Print(price)
}
