package wallet

import (
	"encoding/hex"
	"encoding/json"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/zmagmacore/crypto"
)

type (
	// Wallet represents a wallet that stores keys and additional info.
	Wallet struct {
		ZCNWallet *zcncrypto.Wallet
	}
)

// New creates initialized Wallet.
func New(publicKey, privateKey []byte) *Wallet {
	var (
		publicKeyHex, privateKeyHex = hex.EncodeToString(publicKey), hex.EncodeToString(privateKey)
	)
	return &Wallet{
		ZCNWallet: &zcncrypto.Wallet{
			ClientID:  crypto.Hash(publicKey),
			ClientKey: publicKeyHex,
			Keys: []zcncrypto.KeyPair{
				{
					PublicKey:  publicKeyHex,
					PrivateKey: privateKeyHex,
				},
			},
			Version: zcncrypto.CryptoVersion,
		},
	}
}

// PublicKey returns the public key.
func (w *Wallet) PublicKey() string {
	return w.ZCNWallet.Keys[0].PublicKey
}

// PrivateKey returns the public key.
func (w *Wallet) PrivateKey() string {
	return w.ZCNWallet.Keys[0].PrivateKey
}

// ID returns the client id.
//
// NOTE: client id represents hex encoded SHA3-256 hash of the raw public key.
func (w *Wallet) ID() string {
	return w.ZCNWallet.ClientID
}

// StringJSON returns marshalled to JSON string Wallet.ZCNWallet.
func (w *Wallet) StringJSON() (string, error) {
	byt, err := json.Marshal(w.ZCNWallet)
	if err != nil {
		return "", err
	}

	return string(byt), err
}
