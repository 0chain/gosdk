// Provides low-level functions and types to work with different cryptographic schemes with a unified interface and provide cryptographic operations.
package zcncrypto

import (
	"encoding/json"

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
	// ClientID client unique identifier
	ClientID string `json:"client_id"`

	// ClientKey client public key
	ClientKey string `json:"client_key"`

	// Keys private and public key pair
	Keys []KeyPair `json:"keys"`

	// Mnemonic recovery phrase of the wallet
	Mnemonic string `json:"mnemonics"`

	// Version version of the wallet
	Version string `json:"version"`

	// DateCreated date of wallet creation
	DateCreated string `json:"date_created"`

	// Nonce nonce of the wallet
	Nonce int64 `json:"nonce"`
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

func IsMnemonicValid(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

func Sha3Sum256(data string) string {
	return encryption.Hash(data)
}
