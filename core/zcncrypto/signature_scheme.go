// Provides low-level functions and types to work with different cryptographic schemes with a unified interface and provide cryptographic operations.
package zcncrypto

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/tyler-smith/go-bip39"
)

// CryptoVersion - version of the crypto library
const CryptoVersion = "1.0"

// KeyPair private and publickey
type KeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// Wallet represents client wallet information
type Wallet struct {
	ClientID      string    `json:"client_id"`
	ClientKey     string    `json:"client_key"`
	PeerPublicKey string    `json:"peer_public_key"` // Peer public key exists only in split wallet
	Keys          []KeyPair `json:"keys"`
	Mnemonic      string    `json:"mnemonics"`
	Version       string    `json:"version"`
	DateCreated   string    `json:"date_created"`
	Nonce         int64     `json:"nonce"`
	IsSplit       bool      `json:"is_split"`
}

// SignatureScheme - an encryption scheme for signing and verifying messages
type SignatureScheme interface {
	// Generate fresh keys
	GenerateKeys() (*Wallet, error)
	// Generate fresh keys based on eth wallet
	GenerateKeysWithEth(mnemonic, password string) (*Wallet, error)

	// Generate keys from mnemonic for recovery
	RecoverKeys(mnemonic string) (*Wallet, error)
	GetMnemonic() string

	// Signing  - Set private key to sign
	SetPrivateKey(privateKey string) error
	Sign(hash string) (string, error)

	// Signature verification - Set public key to verify
	SetPublicKey(publicKey string) error
	GetPublicKey() string
	GetPrivateKey() string
	Verify(signature string, msg string) (bool, error)

	// Combine signature for schemes BLS
	Add(signature, msg string) (string, error)

	// implement SplitSignatureScheme

	SplitKeys(numSplits int) (*Wallet, error)

	GetPrivateKeyAsByteArray() ([]byte, error)

	// // implement ThresholdSignatureScheme

	SetID(id string) error
	GetID() string
}

// Marshal returns json string
func (w *Wallet) Marshal() (string, error) {
	ws, err := json.Marshal(w)
	if err != nil {
		return "", errors.New("wallet_marshal", "Invalid Wallet")
	}
	return string(ws), nil
}

func (w *Wallet) Sign(hash, scheme string) (string, error) {
	sigScheme := NewSignatureScheme(scheme)
	err := sigScheme.SetPrivateKey(w.Keys[0].PrivateKey)
	if err != nil {
		return "", err
	}
	return sigScheme.Sign(hash)
}

// SetSplitKeys sets split keys and wipes out mnemonic and original primary keys
func (w *Wallet) SetSplitKeys(sw *Wallet) {
	*w = *sw
}

func (w *Wallet) SaveTo(file string) error {
	d, err := json.Marshal(w)
	if err != nil {
		return err
	}

	fmt.Println("Saving wallet to file: ", string(d))

	return os.WriteFile(file, d, 0644)
}

func IsMnemonicValid(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

func Sha3Sum256(data string) string {
	return encryption.Hash(data)
}
