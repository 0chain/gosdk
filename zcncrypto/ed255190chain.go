package zcncrypto

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/0chain/gosdk/encryption"
	"github.com/tyler-smith/go-bip39"
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
func (ed *ED255190chainScheme) GenerateKeys() error {
	// Check for recovery
	if len(ed.mnemonic) == 0 {
		entropy, err := bip39.NewEntropy(256)
		if err != nil {
			return fmt.Errorf("Getting entropy failed")
		}
		ed.mnemonic, err = bip39.NewMnemonic(entropy)
		if err != nil {
			return fmt.Errorf("Getting mnemonic failed")
		}
	}
	seed := bip39.NewSeed(ed.mnemonic, "0chain-client-ed25519-key")
	r := bytes.NewReader(seed)
	public, private, err := ed25519.GenerateKey(r)
	fmt.Println("public key:", hex.EncodeToString(public))
	fmt.Println("private key:", hex.EncodeToString(private))
	if err != nil {
		return err
	}
	ed.privateKey = private
	ed.publicKey = public
	return nil
}

func (ed *ED255190chainScheme) RecoverKeys(mnemonic string) error {
	if mnemonic == "" {
		return errors.New("Set mnemonic key failed")
	}
	if len(ed.privateKey) > 0 || len(ed.publicKey) > 0 {
		return errors.New("Cannot recover when there are keys")
	}
	ed.mnemonic = mnemonic
	ed.GenerateKeys()
	return nil
}

func (ed *ED255190chainScheme) GetPublicKey() (string, error) {
	if len(ed.publicKey) == 0 {
		return "", errors.New("Key Not Found")
	}
	return hex.EncodeToString(ed.publicKey), nil
}

func (ed *ED255190chainScheme) GetPublicKeyWithIdx(i int) (string, error) {
	if i < 1 {
		return ed.GetPublicKey()
	}
	return "", errors.New("Get publickey invalid index")
}

func (ed *ED255190chainScheme) GetSecretKeyWithIdx(i int) (string, error) {
	if len(ed.privateKey) == 0 {
		return "", errors.New("Key Not Found")
	}
	if i > 0 {
		return "", errors.New("Get private invalid index")
	}
	return hex.EncodeToString(ed.privateKey), nil
}

func (ed *ED255190chainScheme) GetMnemonic() (string, error) {
	if len(ed.mnemonic) == 0 {
		return "", fmt.Errorf("Mnemonic found only if key is generated")
	}
	return ed.mnemonic, nil
}

func (ed *ED255190chainScheme) SetPrivateKey(privateKey string) error {
	if len(ed.privateKey) > 0 {
		return errors.New("Cannot set private key when there is a public key")
	}
	if len(ed.publicKey) > 0 {
		return errors.New("Private key already exists")
	}
	var err error
	ed.privateKey, err = hex.DecodeString(privateKey)
	return err
}

func (ed *ED255190chainScheme) SetPublicKey(publicKey string) error {
	if len(ed.publicKey) > 0 {
		return errors.New("cannot set public key when there is a private key")
	}
	if len(ed.privateKey) > 0 {
		return errors.New("Public key already exists")
	}
	var err error
	ed.publicKey, err = hex.DecodeString(publicKey)
	return err
}

func (ed *ED255190chainScheme) Sign(hash string) (string, error) {
	if len(ed.privateKey) == 0 {
		return "", errors.New("private key does not exists for signing")
	}
	rawHash := encryption.RawHash(hash)
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
	data, err := hex.DecodeString(encryption.Hash(msg))
	if err != nil {
		return false, err
	}
	return ed25519.Verify(ed.publicKey, data, sign), nil
}

func (ed *ED255190chainScheme) Add(signature, msg string) (string, error) {
	return "", nil
}
