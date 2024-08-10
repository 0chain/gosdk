package zcnbridge

import (
	"fmt"
	"path"
	"time"

	hdw "github.com/0chain/gosdk/zcncore/ethhdwallet"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

// DetailedAccount describes detailed account
type DetailedAccount struct {
	EthereumAddress,
	PublicKey,
	PrivateKey accounts.Account
}

// KeyStore is a wrapper, which exposes Ethereum KeyStore methods used by DEX bridge.
type KeyStore interface {
	Find(accounts.Account) (accounts.Account, error)
	TimedUnlock(accounts.Account, string, time.Duration) error
	SignHash(account accounts.Account, hash []byte) ([]byte, error)
	GetEthereumKeyStore() *keystore.KeyStore
}

type keyStore struct {
	ks *keystore.KeyStore
}

// NewKeyStore creates new KeyStore wrapper instance
func NewKeyStore(path string) KeyStore {
	return &keyStore{
		ks: keystore.NewKeyStore(path, keystore.StandardScryptN, keystore.StandardScryptP),
	}
}

// Find forwards request to Ethereum KeyStore Find method
func (k *keyStore) Find(account accounts.Account) (accounts.Account, error) {
	return k.ks.Find(account)
}

// TimedUnlock forwards request to Ethereum KeyStore TimedUnlock method
func (k *keyStore) TimedUnlock(account accounts.Account, passPhrase string, timeout time.Duration) error {
	return k.ks.TimedUnlock(account, passPhrase, timeout)
}

// SignHash forwards request to Ethereum KeyStore SignHash method
func (k *keyStore) SignHash(account accounts.Account, hash []byte) ([]byte, error) {
	return k.ks.SignHash(account, hash)
}

// GetEthereumKeyStore returns Ethereum KeyStore instance
func (k *keyStore) GetEthereumKeyStore() *keystore.KeyStore {
	return k.ks
}

// ListStorageAccounts List available accounts
//   - homedir is the home directory
func ListStorageAccounts(homedir string) []common.Address {
	keyDir := path.Join(homedir, EthereumWalletStorageDir)
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	config := &accounts.Config{InsecureUnlockAllowed: false}
	am := accounts.NewManager(config, ks)
	addresses := am.Accounts()

	return addresses
}

// DeleteAccount deletes account from wallet
//   - homedir is the home directory
//   - address is the account address
func DeleteAccount(homedir, address string) bool {
	keyDir := path.Join(homedir, EthereumWalletStorageDir)
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	config := &accounts.Config{InsecureUnlockAllowed: false}
	am := accounts.NewManager(config, ks)

	wallet, err := am.Find(accounts.Account{
		Address: common.HexToAddress(address),
	})

	if err != nil && wallet == nil {
		fmt.Printf("failed to find account %s, error: %s", address, err)
		return false
	}

	return true
}

// AccountExists checks if account exists
//   - homedir is the home directory
//   - address is the account address
func AccountExists(homedir, address string) bool {
	keyDir := path.Join(homedir, EthereumWalletStorageDir)
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	config := &accounts.Config{InsecureUnlockAllowed: false}
	am := accounts.NewManager(config, ks)

	wallet, err := am.Find(accounts.Account{
		Address: common.HexToAddress(address),
	})

	if err != nil && wallet == nil {
		fmt.Printf("failed to find account %s, error: %s\n", address, err)
		return false
	}

	status, _ := wallet.Status()
	url := wallet.URL()

	fmt.Printf("Account exists. Status: %s, Path: %s\n", status, url)

	return true
}

// CreateKeyStorage create, restore or unlock key storage
//   - homedir is the home directory
//   - password is the password
func CreateKeyStorage(homedir, password string) error {
	keyDir := path.Join(homedir, EthereumWalletStorageDir)
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.NewAccount(password)
	if err != nil {
		return errors.Wrap(err, "failed to create keystore")
	}
	fmt.Printf("Created account: %s", account.Address.Hex())

	return nil
}

// AccountAddressIndex represents client account and address indexes for multi-key wallet. Used to derive ethereum account.
type AccountAddressIndex struct {
	AccountIndex int
	AddressIndex int

	// Bip32 flag indicates if the account is derived using BIP32 derivation path.
	Bip32 bool
}

// ImportAccount imports account using mnemonic
//   - homedir is the home directory
//   - mnemonic is the mnemonic phrase
//   - password is the password
//   - accountAddrIndex is the account and address indexes used for the derivation of the ethereum account
func ImportAccount(homedir, mnemonic, password string, accountAddrIndex ...AccountAddressIndex) (string, error) {
	// 1. Create storage and account if it doesn't exist and add account to it

	keyDir := path.Join(homedir, EthereumWalletStorageDir)
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)

	// 2. Init wallet

	wallet, err := hdw.NewFromMnemonic(mnemonic)
	if err != nil {
		return "", errors.Wrap(err, "failed to import from mnemonic")
	}

	var aai AccountAddressIndex
	if len(accountAddrIndex) > 0 {
		aai = accountAddrIndex[0]
	}

	var pathD accounts.DerivationPath
	if aai.Bip32 {
		pathD = hdw.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/%d", aai.AddressIndex))
	} else {
		pathD = hdw.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/%d'/0/%d", aai.AccountIndex, aai.AddressIndex))
	}

	account, err := wallet.Derive(pathD, true)
	if err != nil {
		return "", errors.Wrap(err, "failed parse derivation path")
	}

	key, err := wallet.PrivateKey(account)
	if err != nil {
		return "", errors.Wrap(err, "failed to get private key")
	}

	// 3. Find key

	acc, err := ks.Find(account)
	if err == nil {
		fmt.Printf("Account already exists: %s\nPath: %s\n\n", acc.Address.Hex(), acc.URL.Path)
		return acc.Address.Hex(), nil
	}

	// 4. Import the key if it doesn't exist

	acc, err = ks.ImportECDSA(key, password)
	if err != nil {
		return "", errors.Wrap(err, "failed to get import private key")
	}

	fmt.Printf("Imported account %s to path: %s\n", acc.Address.Hex(), acc.URL.Path)

	return acc.Address.Hex(), nil
}
