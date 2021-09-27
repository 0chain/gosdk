package zcncrypto

import (
	"encoding/json"
	"fmt"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/tyler-smith/go-bip39"
)

const CryptoVersion = "1.0"

// KeyPair private and publickey
type KeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// Wallet structure
type Wallet struct {
	ClientID    string    `json:"client_id"`
	ClientKey   string    `json:"client_key"`
	Keys        []KeyPair `json:"keys"`
	Mnemonic    string    `json:"mnemonics"`
	Version     string    `json:"version"`
	DateCreated string    `json:"date_created"`
}

//SignatureScheme - an encryption scheme for signing and verifying messages
type SignatureScheme interface {
	// Generate fresh keys
	GenerateKeys() (*Wallet, error)
	// Generate fresh keys based on eth wallet
	GenerateKeysWithEth(mnemonic, password string) (*Wallet, error)

	// Generate keys from mnemonic for recovery
	RecoverKeys(mnemonic string) (*Wallet, error)

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
}

// SplitSignatureScheme splits the primary key into number of parts.
type SplitSignatureScheme interface {
	SignatureScheme
	SplitKeys(numSplits int) (*Wallet, error)
}

// NewSignatureScheme creates an instance for using signature functions
func NewSignatureScheme(sigScheme string) SignatureScheme {
	switch sigScheme {
	case "ed25519":
		return NewED255190chainScheme()
	case "bls0chain":
		return NewBLS0ChainScheme()
	default:
		panic(fmt.Sprintf("unknown signature scheme: %v", sigScheme))
	}
}

// Marshal returns json string
func (w *Wallet) Marshal() (string, error) {
	ws, err := json.Marshal(w)
	if err != nil {
		return "", errors.New("wallet_marshal", "Invalid Wallet")
	}
	return string(ws), nil
}

func IsMnemonicValid(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

func Sha3Sum256(data string) string {
	return encryption.Hash(data)
}
