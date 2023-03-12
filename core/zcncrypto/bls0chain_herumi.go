//go:build !js && !wasm
// +build !js,!wasm

package zcncrypto

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/0chain/errors"
	"github.com/tyler-smith/go-bip39"

	"github.com/0chain/gosdk/core/encryption"
)

func init() {

}

// HerumiScheme - a signature scheme for BLS0Chain Signature
type HerumiScheme struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Mnemonic   string `json:"mnemonic"`

	id  ID
	Ids string `json:"threshold_scheme_id"`
}

// NewHerumiScheme - create a MiraclScheme object
func NewHerumiScheme() *HerumiScheme {
	return &HerumiScheme{
		id: BlsSignerInstance.NewID(),
	}
}

// GenerateKeys  generate fresh keys
func (b0 *HerumiScheme) GenerateKeys() (*Wallet, error) {
	return b0.generateKeys("0chain-client-split-key")
}

// GenerateKeysWithEth  generate fresh keys based on eth wallet
func (b0 *HerumiScheme) GenerateKeysWithEth(mnemonic, password string) (*Wallet, error) {
	if len(mnemonic) == 0 {
		return nil, fmt.Errorf("Mnemonic phase is mandatory.")
	}
	b0.Mnemonic = mnemonic

	_, err := bip39.NewSeedWithErrorChecking(b0.Mnemonic, password)
	if err != nil {
		return nil, fmt.Errorf("Wrong mnemonic phase.")
	}

	return b0.generateKeys(password)
}

// RecoverKeys recovery keys from mnemonic
func (b0 *HerumiScheme) RecoverKeys(mnemonic string) (*Wallet, error) {
	if mnemonic == "" {
		return nil, errors.New("recover_keys", "Set mnemonic key failed")
	}
	if b0.PublicKey != "" || b0.PrivateKey != "" {
		return nil, errors.New("recover_keys", "Cannot recover when there are keys")
	}
	b0.Mnemonic = mnemonic
	return b0.GenerateKeys()
}

func (b0 *HerumiScheme) GetMnemonic() string {
	if b0 == nil {
		return ""
	}

	return b0.Mnemonic
}

// SetPrivateKey  set private key to sign
func (b0 *HerumiScheme) SetPrivateKey(privateKey string) error {
	if b0.PublicKey != "" {
		return errors.New("set_private_key", "cannot set private key when there is a public key")
	}
	if b0.PrivateKey != "" {
		return errors.New("set_private_key", "private key already exists")
	}
	b0.PrivateKey = privateKey
	//ToDo: b0.publicKey should be set here?
	return nil
}

func (b0 *HerumiScheme) GetPrivateKey() string {
	return b0.PrivateKey
}

func (b0 *HerumiScheme) SplitKeys(numSplits int) (*Wallet, error) {
	if b0.PrivateKey == "" {
		return nil, errors.New("split_keys", "primary private key not found")
	}
	primaryFr := BlsSignerInstance.NewFr()
	primarySk := BlsSignerInstance.NewSecretKey()
	err := primarySk.DeserializeHexStr(b0.PrivateKey)
	if err != nil {
		return nil, err
	}
	err = primaryFr.SetLittleEndian(primarySk.GetLittleEndian())

	if err != nil {
		return nil, err
	}

	// New Wallet
	w := &Wallet{}
	w.Keys = make([]KeyPair, numSplits)
	sk := BlsSignerInstance.NewSecretKey()
	for i := 0; i < numSplits-1; i++ {
		tmpSk := BlsSignerInstance.NewSecretKey()
		tmpSk.SetByCSPRNG()
		w.Keys[i].PrivateKey = tmpSk.SerializeToHexStr()
		pub := tmpSk.GetPublicKey()
		w.Keys[i].PublicKey = pub.SerializeToHexStr()
		sk.Add(tmpSk)
	}
	aggregateSk := BlsSignerInstance.NewFr()
	err = aggregateSk.SetLittleEndian(sk.GetLittleEndian())

	if err != nil {
		return nil, err
	}

	//Subtract the aggregated private key from the primary private key to derive the last split private key
	lastSk := BlsSignerInstance.NewFr()
	BlsSignerInstance.FrSub(lastSk, primaryFr, aggregateSk)

	// Last key
	lastSecretKey := BlsSignerInstance.NewSecretKey()
	err = lastSecretKey.SetLittleEndian(lastSk.Serialize())
	if err != nil {
		return nil, err
	}
	w.Keys[numSplits-1].PrivateKey = lastSecretKey.SerializeToHexStr()
	w.Keys[numSplits-1].PublicKey = lastSecretKey.GetPublicKey().SerializeToHexStr()

	// Generate client ID and public
	w.ClientKey = primarySk.GetPublicKey().SerializeToHexStr()
	w.ClientID = encryption.Hash(primarySk.GetPublicKey().Serialize())
	w.Mnemonic = b0.Mnemonic
	w.Version = CryptoVersion
	w.DateCreated = time.Now().Format(time.RFC3339)

	return w, nil
}

// Sign sign message
func (b0 *HerumiScheme) Sign(hash string) (string, error) {
	sig, err := b0.rawSign(hash)
	if err != nil {
		return "", err
	}
	return sig.SerializeToHexStr(), nil
}

// SetPublicKey - implement interface
func (b0 *HerumiScheme) SetPublicKey(publicKey string) error {
	if b0.PrivateKey != "" {
		return errors.New("set_public_key", "cannot set public key when there is a private key")
	}
	if b0.PublicKey != "" {
		return errors.New("set_public_key", "public key already exists")
	}
	b0.PublicKey = publicKey
	return nil
}

// GetPublicKey - implement interface
func (b0 *HerumiScheme) GetPublicKey() string {
	return b0.PublicKey
}

// Verify - implement interface
func (b0 *HerumiScheme) Verify(signature, msg string) (bool, error) {
	if b0.PublicKey == "" {
		return false, errors.New("verify", "public key does not exists for verification")
	}
	sig := BlsSignerInstance.NewSignature()
	pk := BlsSignerInstance.NewPublicKey()
	err := sig.DeserializeHexStr(signature)
	if err != nil {
		return false, err
	}
	rawHash, err := hex.DecodeString(msg)
	if err != nil {
		return false, err
	}
	if rawHash == nil {
		return false, errors.New("verify", "failed hash while signing")
	}
	err = pk.DeserializeHexStr(b0.PublicKey)
	if err != nil {
		return false, err
	}

	return sig.Verify(pk, string(rawHash)), nil
}

func (b0 *HerumiScheme) Add(signature, msg string) (string, error) {
	sign := BlsSignerInstance.NewSignature()
	err := sign.DeserializeHexStr(signature)
	if err != nil {
		return "", err
	}
	signature1, err := b0.rawSign(msg)
	if err != nil {
		return "", errors.Wrap(err, "BLS signing failed")
	}
	sign.Add(signature1)
	return sign.SerializeToHexStr(), nil
}

// GetPrivateKeyAsByteArray - converts private key into byte array
func (b0 *HerumiScheme) GetPrivateKeyAsByteArray() ([]byte, error) {
	if len(b0.PrivateKey) == 0 {
		return nil, errors.New("get_private_key_as_byte_array", "cannot convert empty private key to byte array")
	}
	privateKeyBytes, err := hex.DecodeString(b0.PrivateKey)
	if err != nil {
		return nil, err
	}
	return privateKeyBytes, nil
}

// SetID sets ID in HexString format
func (b0 *HerumiScheme) SetID(id string) error {
	if b0.id == nil {
		b0.id = BlsSignerInstance.NewID()
	}
	b0.Ids = id
	return b0.id.SetHexString(id)
}

// GetID gets ID in hex string format
func (b0 *HerumiScheme) GetID() string {
	if b0.id == nil {
		b0.id = BlsSignerInstance.NewID()
	}
	return b0.id.GetHexString()
}

func (b0 *HerumiScheme) generateKeys(password string) (*Wallet, error) {
	// Check for recovery
	if len(b0.Mnemonic) == 0 {
		entropy, err := bip39.NewEntropy(256)
		if err != nil {
			return nil, errors.Wrap(err, "Generating entropy failed")
		}
		b0.Mnemonic, err = bip39.NewMnemonic(entropy)
		if err != nil {
			return nil, errors.Wrap(err, "Generating mnemonic failed")
		}
	}

	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(b0.Mnemonic, password)
	r := bytes.NewReader(seed)
	BlsSignerInstance.SetRandFunc(r)

	// New Wallet
	w := &Wallet{}
	w.Keys = make([]KeyPair, 1)

	// Generate pair
	sk := BlsSignerInstance.NewSecretKey()
	sk.SetByCSPRNG()
	w.Keys[0].PrivateKey = sk.SerializeToHexStr()
	pub := sk.GetPublicKey()
	w.Keys[0].PublicKey = pub.SerializeToHexStr()

	b0.PrivateKey = w.Keys[0].PrivateKey
	b0.PublicKey = w.Keys[0].PublicKey
	w.ClientKey = w.Keys[0].PublicKey
	w.ClientID = encryption.Hash(pub.Serialize())
	w.Mnemonic = b0.Mnemonic
	w.Version = CryptoVersion
	w.DateCreated = time.Now().Format(time.RFC3339)

	// Revert the Random function to default
	BlsSignerInstance.SetRandFunc(nil)
	return w, nil
}

func (b0 *HerumiScheme) rawSign(hash string) (Signature, error) {
	sk := BlsSignerInstance.NewSecretKey()
	if b0.PrivateKey == "" {
		return nil, errors.New("raw_sign", "private key does not exists for signing")
	}
	rawHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}
	if rawHash == nil {
		return nil, errors.New("raw_sign", "failed hash while signing")
	}
	sk.SetByCSPRNG()
	err = sk.DeserializeHexStr(b0.PrivateKey)
	if err != nil {
		return nil, err
	}
	sig := sk.Sign(string(rawHash))
	return sig, nil
}
