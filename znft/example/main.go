package main

import (
	"context"
	"log"

	"github.com/0chain/gosdk/znft"
)

func main() {
	// Create this configuration
	cfg := &znft.Configuration{
		FactoryAddress:                   "0xD98602749e7f46036d496e8D3deb6eb9F90996a6",
		FactoryModuleERC721Address:       "",
		FactoryModuleERC721FixedAddress:  "",
		FactoryModuleERC721RandomAddress: "",
		EthereumNodeURL:                  "https://rinkeby.infura.io/v3/22cb2849f5f74b8599f3dc2a23085bd4",
		WalletAddress:                    "",
		VaultPassword:                    "",
		Homedir:                          "",
		GasLimit:                         0,
		Value:                            0,
	}

	// Create NFT with factory
	addressFixed := "beef"

	// Create session of NFT token
	sessionFixed, err := cfg.CreateStorageERC721FixedSession(context.Background(), addressFixed)
	if err != nil {
		panic(err)
	}

	// Start using session
	price, err := sessionFixed.Price()
	if err != nil {
		return
	}

	// Create NFT with factory
	address := "beef"

	// Create session of NFT token
	session, err := cfg.CreateStorageERC721Session(context.Background(), address)
	if err != nil {
		panic(err)
	}

	// Start using session
	price, err = session.Price()
	if err != nil {
		return
	}

	log.Print(price)
}
