package zcnbridge

import (
	"fmt"
	"path"

	hdw "github.com/miguelmota/go-ethereum-hdwallet"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/pkg/errors"
)

// ListAccount List available accounts
func ListAccount() {
	keyDir := path.Join(GetConfigDir(), "wallets")
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	config := &accounts.Config{InsecureUnlockAllowed: false}
	am := accounts.NewManager(config, ks)
	addresses := am.Accounts()
	fmt.Println(addresses)
}

// CreateKeyStorage create, restore or unlock key storage
func CreateKeyStorage(password string) error {
	keyDir := path.Join(GetConfigDir(), "wallets")
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.NewAccount(password)
	if err != nil {
		return errors.Wrap(err, "failed to create keystore")
	}
	fmt.Printf("Created account: %s", account.Address.Hex())

	return nil
}

func ImportAccount(mnemonic, password string) error {
	// 1. Create storage and account if it doesn't exist and add account to it

	keyDir := path.Join(GetConfigDir(), "wallets")
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)

	// 2. Init wallet

	wallet, err := hdw.NewFromMnemonic(mnemonic)
	if err != nil {
		return errors.Wrap(err, "failed to import from mnemonic")
	}

	pathD := hdw.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(pathD, true)
	if err != nil {
		return errors.Wrap(err, "failed parse derivation path")
	}

	key, err := wallet.PrivateKey(account)
	if err != nil {
		return errors.Wrap(err, "failed to get private key")
	}

	// 3. Find key

	acc, err := ks.Find(account)
	if err == nil {
		fmt.Printf("Account already exists %s\n, Path: %s", acc.Address.Hex(), acc.URL.Path)
		return nil
	}

	// 4. Import the key if it doesn't exist

	acc, err = ks.ImportECDSA(key, password)
	if err != nil {
		return errors.Wrap(err, "failed to get import private key")
	}

	fmt.Printf("Imported account %s to path: %s\n", acc.Address.Hex(), acc.URL.Path)

	return nil
}
