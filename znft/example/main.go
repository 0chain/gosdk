package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/0chain/gosdk/znft"
)

var mnemonic string

func main() {
	mnemonic = "float actor page tip vital solution knock agent cage grant size funny"

	// Create this configuration
	cfg := &znft.Configuration{
		FactoryAddress:                   "0xD98602749e7f46036d496e8D3deb6eb9F90996a6",
		FactoryModuleERC721Address:       "",
		FactoryModuleERC721FixedAddress:  "",
		FactoryModuleERC721RandomAddress: "",
		EthereumNodeURL:                  "https://rinkeby.infura.io/v3/22cb2849f5f74b8599f3dc2a23085bd4",
		WalletAddress:                    "0x860FA46F170a87dF44D7bB867AA4a5D2813127c1",
		VaultPassword:                    "pass",
		Homedir:                          "", // optional, default = "~/.zcn"
		Value:                            0,  // mandatory, default > 0
	}

	if cfg.Homedir == "" {
		cfg.Homedir = znft.GetConfigDir()
	}

	if !znft.AccountExists(cfg.Homedir, cfg.WalletAddress) {
		_, err := znft.ImportAccount(cfg.Homedir, mnemonic, cfg.VaultPassword)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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
		panic(err)
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
		panic(err)
	}

	log.Print(price)
}
