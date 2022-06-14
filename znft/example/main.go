package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/0chain/gosdk/znft"
)

var mnemonic string

func main() {
	mnemonic = "use your mnemonic to import account to a local storage"

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
	address := "beef"
	ownerAddress := "beef"

	// Creating NFTs

	factorySession, err := cfg.CreateFactoryERC721Session(context.Background(), address)
	if err != nil {
		panic(err)
	}

	data := []byte("")
	max := new(big.Int).SetInt64(10000)
	err = factorySession.CreateToken(
		ownerAddress,
		"TestFixedPriceToken",
		"dNFT",
		"https://api.0chain.net/v1/file/abcdefgh/",
		max,
		data,
	)
	if err != nil {
		panic(err)
	}

	// Reading and writing to NFT smart contracts

	// Create session of NFT token
	sessionRandom, err := cfg.CreateStorageERC721RandomSession(context.Background(), address)
	if err != nil {
		panic(err)
	}

	// Start using session
	price, err := sessionRandom.Price()
	if err != nil {
		panic(err)
	}
	log.Print(price)

	// ERC721Fixed

	// Create session of NFT token
	sessionFixed, err := cfg.CreateStorageERC721FixedSession(context.Background(), address)
	if err != nil {
		panic(err)
	}

	// Start using session
	price, err = sessionFixed.Price()
	if err != nil {
		panic(err)
	}
	log.Print(price)

	// ERC721Pack

	// Create session of NFT token
	sessionPack, err := cfg.CreateStorageERC721PackSession(context.Background(), address)
	if err != nil {
		panic(err)
	}

	// Start using session
	price, err = sessionPack.Price()
	if err != nil {
		panic(err)
	}
	log.Print(price)

	// ERC721

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
