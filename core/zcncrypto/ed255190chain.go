package zcncrypto

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
	"github.com/tyler-smith/go-bip39"
	"github.com/0chain/gosdk/core/encryption"
	"golang.org/x/crypto/ed25519"
)

//ED255190chainScheme - a signature scheme based on ED25519
type ED255190chainScheme struct {
	privateKey []byte
	publicKey  []byte
	mnemonic   string
}

//NewED25519Scheme - create a ED255219Scheme object
func NewED255190chainScheme() *ED255190chainScheme {
	return &ED255190chainScheme{}
}

//GenerateKeys - implement interface
func (ed *ED255190chainScheme) GenerateKeys() (*Wallet, error) {
	// Check for recovery
	if len(ed.mnemonic) == 0 {
		entropy, err := bip39.NewEntropy(256)
		if err != nil {
			return nil, fmt.Errorf("Getting entropy failed")
		}
		ed.mnemonic, err = bip39.NewMnemonic(entropy)
		if err != nil {
			return nil, fmt.Errorf("Getting mnemonic failed")
		}
	}

	seed := bip39.NewSeed(ed.mnemonic, "0chain-client-ed25519-key")
	r := bytes.NewReader(seed)
	public, private, err := ed25519.GenerateKey(r)
	if err != nil {
		return nil, fmt.Errorf("Generate keys failed - %s", err.Error())
	}
	// New Wallet
	w := &Wallet{}
	w.Keys = make([]KeyPair, 1)
	w.Keys[0].PublicKey = hex.EncodeToString(public)
	w.Keys[0].PrivateKey = hex.EncodeToString(private)
	w.ClientKey = w.Keys[0].PublicKey
	w.ClientID = encryption.Hash([]byte(public))
	w.Mnemonic = ed.mnemonic
	w.Version = cryptoVersion
	w.DateCreated = time.Now().String()
	return w, nil
}

func (ed *ED255190chainScheme) RecoverKeys(mnemonic string) (*Wallet, error) {
	if mnemonic == "" {
		return nil, errors.New("Set mnemonic key failed")
	}
	if len(ed.privateKey) > 0 || len(ed.publicKey) > 0 {
		return nil, errors.New("Cannot recover when there are keys")
	}
	ed.mnemonic = mnemonic
	return ed.GenerateKeys()
}

func (ed *ED255190chainScheme) SetPrivateKey(privateKey string) error {
	if len(ed.privateKey) > 0 {
		return errors.New("cannot set private key when there is a public key")
	}
	if len(ed.publicKey) > 0 {
		return errors.New("private key already exists")
	}
	var err error
	ed.privateKey, err = hex.DecodeString(privateKey)
	return err
}

func (ed *ED255190chainScheme) SetPublicKey(publicKey string) error {
	if len(ed.privateKey) > 0 {
		return errors.New("cannot set public key when there is a private key")
	}
	if len(ed.publicKey) > 0 {
		return errors.New("public key already exists")
	}
	var err error
	ed.publicKey, err = hex.DecodeString(publicKey)
	return err
}

func (ed *ED255190chainScheme) Sign(hash string) (string, error) {
	if len(ed.privateKey) == 0 {
		return "", errors.New("private key does not exists for signing")
	}
	rawHash, err := hex.DecodeString(hash)
	if err != nil {
		return "", err
	}
	if rawHash == nil {
		return "", errors.New("Failed hash while signing")
	}
	return hex.EncodeToString(ed25519.Sign(ed.privateKey, rawHash)), nil
}

func (ed *ED255190chainScheme) Verify(signature, msg string) (bool, error) {
	if len(ed.publicKey) == 0 {
		return false, errors.New("public key does not exists for verification")
	}
	sign, err := hex.DecodeString(signature)
	if err != nil {
		return false, err
	}
	data, err := hex.DecodeString(msg)
	if err != nil {
		return false, err
	}
	return ed25519.Verify(ed.publicKey, data, sign), nil
}

func (ed *ED255190chainScheme) Add(signature, msg string) (string, error) {
	return "", fmt.Errorf("Not supported by signature scheme")
}
