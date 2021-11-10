// +build !js,!wasm

package zcncrypto

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/0chain/errors"
	"github.com/herumi/bls-go-binary/bls"
	"github.com/tyler-smith/go-bip39"

	"github.com/0chain/gosdk/core/encryption"
)

func init() {
	err := bls.Init(bls.CurveFp254BNb)
	if err != nil {
		panic(err)
	}
}

//MircalScheme - a signature scheme for BLS0Chain Signature
type MircalScheme struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Mnemonic   string `json:"mnemonic"`
}

//NewMircalScheme - create a MircalScheme object
func NewMircalScheme() *MircalScheme {
	return &MircalScheme{}
}

// GenerateKeys  generate fresh keys
func (b0 *MircalScheme) GenerateKeys() (*Wallet, error) {
	return b0.generateKeys("0chain-client-split-key")
}

// GenerateKeysWithEth  generate fresh keys based on eth wallet
func (b0 *MircalScheme) GenerateKeysWithEth(mnemonic, password string) (*Wallet, error) {
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
func (b0 *MircalScheme) RecoverKeys(mnemonic string) (*Wallet, error) {
	if mnemonic == "" {
		return nil, errors.New("recover_keys", "Set mnemonic key failed")
	}
	if b0.PublicKey != "" || b0.PrivateKey != "" {
		return nil, errors.New("recover_keys", "Cannot recover when there are keys")
	}
	b0.Mnemonic = mnemonic
	return b0.GenerateKeys()
}

//SetPrivateKey  set private key to sign
func (b0 *MircalScheme) SetPrivateKey(privateKey string) error {
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

func (b0 *MircalScheme) GetPrivateKey() string {
	return b0.PrivateKey
}

//Sign sign message
func (b0 *MircalScheme) Sign(hash string) (string, error) {
	sig, err := b0.rawSign(hash)
	if err != nil {
		return "", err
	}
	return sig.SerializeToHexStr(), nil
}

//SetPublicKey - implement interface
func (b0 *MircalScheme) SetPublicKey(publicKey string) error {
	if b0.PrivateKey != "" {
		return errors.New("set_public_key", "cannot set public key when there is a private key")
	}
	if b0.PublicKey != "" {
		return errors.New("set_public_key", "public key already exists")
	}
	b0.PublicKey = MiraclToHerumiPK(publicKey)
	return nil
}

//GetPublicKey - implement interface
func (b0 *MircalScheme) GetPublicKey() string {
	return b0.PublicKey
}

//Verify - implement interface
func (b0 *MircalScheme) Verify(signature, msg string) (bool, error) {
	if b0.PublicKey == "" {
		return false, errors.New("verify", "public key does not exists for verification")
	}
	var sig bls.Sign
	var pk bls.PublicKey
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
	pk.DeserializeHexStr(b0.PublicKey)
	return sig.Verify(&pk, string(rawHash)), nil
}

func (b0 *MircalScheme) Add(signature, msg string) (string, error) {
	var sign bls.Sign
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

func (b0 *MircalScheme) generateKeys(password string) (*Wallet, error) {
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
	bls.SetRandFunc(r)

	// New Wallet
	w := &Wallet{}
	w.Keys = make([]KeyPair, 1)

	// Generate pair
	var sk bls.SecretKey
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
	bls.SetRandFunc(nil)
	return w, nil
}

func (b0 *MircalScheme) rawSign(hash string) (*bls.Sign, error) {
	var sk bls.SecretKey
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
	sk.DeserializeHexStr(b0.PrivateKey)
	sig := sk.Sign(string(rawHash))
	return sig, nil
}
