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
	"github.com/miracl/core/go/core/BN254"
)

func init() {

	res := BN254.Init()
	if res != 0 {
		panic("Failed to Initialize BN254\n")
	}
}

//MiraclScheme - a signature scheme for BLS0Chain Signature
type MiraclScheme struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Mnemonic   string `json:"mnemonic"`
}

//NewMiraclScheme - create a MiraclScheme object
func NewMiraclScheme() *MiraclScheme {
	return &MiraclScheme{}
}

// GenerateKeys  generate fresh keys
func (b0 *MiraclScheme) GenerateKeys() (*Wallet, error) {
	return b0.generateKeys("0chain-client-split-key")
}

// GenerateKeysWithEth  generate fresh keys based on eth wallet
func (b0 *MiraclScheme) GenerateKeysWithEth(mnemonic, password string) (*Wallet, error) {
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
func (b0 *MiraclScheme) RecoverKeys(mnemonic string) (*Wallet, error) {
	if mnemonic == "" {
		return nil, errors.New("recover_keys", "Set mnemonic key failed")
	}
	if b0.PublicKey != "" || b0.PrivateKey != "" {
		return nil, errors.New("recover_keys", "Cannot recover when there are keys")
	}
	b0.Mnemonic = mnemonic
	return b0.GenerateKeys()
}

func (b0 *MiraclScheme) GetMnemonic() string {
	if b0 == nil {
		return ""
	}

	return b0.Mnemonic
}

//SetPrivateKey  set private key to sign
func (b0 *MiraclScheme) SetPrivateKey(privateKey string) error {
	if b0.PublicKey != "" {
		return errors.New("set_private_key", "cannot set private key when there is a public key")
	}
	if b0.PrivateKey != "" {
		return errors.New("set_private_key", "private key already exists")
	}
	b0.PrivateKey = privateKey
	return nil
}

func (b0 *MiraclScheme) GetPrivateKey() string {
	return b0.PrivateKey
}

func (b0 *MiraclScheme) SplitKeys(numSplits int) (*Wallet, error) {
	if b0.PrivateKey == "" {
		return nil, errors.New("split_keys", "primary private key not found")
	}

	return nil, errors.New("non-implemented", "bls_split_keys")
}

//Sign sign message
func (b0 *MiraclScheme) Sign(hash string) (string, error) {

	if b0.PrivateKey == "" {
		return "", errors.New("raw_sign", "private key does not exists for signing")
	}

	rawHash, err := hex.DecodeString(hash)
	if err != nil {
		return "", err
	}
	if rawHash == nil {
		return "", errors.New("raw_sign", "failed hash while signing")
	}

	const BFS = BN254.BFS
	const G1S = BFS + 1 /* Group 1 Size */

	var SIG [G1S]byte

	S, err := hex.DecodeString(b0.PrivateKey)
	if err != nil {
		return "", err
	}

	BN254.Core_Sign(SIG[:], rawHash, S[:])

	return hex.EncodeToString(SIG[:]), nil

}

//SetPublicKey - implement interface
func (b0 *MiraclScheme) SetPublicKey(publicKey string) error {
	if b0.PrivateKey != "" {
		return errors.New("set_public_key", "cannot set public key when there is a private key")
	}
	if b0.PublicKey != "" {
		return errors.New("set_public_key", "public key already exists")
	}
	b0.PublicKey = publicKey
	//b0.PublicKey = MiraclToHerumiPK(publicKey)
	return nil
}

//GetPublicKey - implement interface
func (b0 *MiraclScheme) GetPublicKey() string {
	return b0.PublicKey
}

//Verify - implement interface
func (b0 *MiraclScheme) Verify(signature, msg string) (bool, error) {

	if b0.PublicKey == "" {
		return false, errors.New("verify", "public key does not exists for verification")
	}

	W, err := hex.DecodeString(b0.PublicKey)
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

	SIG, err := hex.DecodeString(signature)
	if err != nil {
		return false, err
	}

	res := BN254.Core_Verify(SIG[:], rawHash, W[:])

	if res == 0 {
		return true, nil
	}

	return false, errors.New("invalid_signature", "invalid signature")
}

func (b0 *MiraclScheme) Add(signature, msg string) (string, error) {
	return "", errors.New("non-implemented", "bls_add")
}

// GetPrivateKeyAsByteArray - converts private key into byte array
func (b0 *MiraclScheme) GetPrivateKeyAsByteArray() ([]byte, error) {
	if len(b0.PrivateKey) == 0 {
		return nil, errors.New("get_private_key_as_byte_array", "cannot convert empty private key to byte array")
	}
	privateKeyBytes, err := hex.DecodeString(b0.PrivateKey)
	if err != nil {
		return nil, err
	}
	return privateKeyBytes, nil
}

func (b0 *MiraclScheme) generateKeys(password string) (*Wallet, error) {
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

	w := &Wallet{}
	w.Keys = make([]KeyPair, 1)

	const BGS = BN254.BGS
	const BFS = BN254.BFS
	//	const G1S = BFS + 1   /* Group 1 Size */
	const G2S = 2*BFS + 1 /* Group 2 Size */

	var S [BGS]byte
	var W [G2S]byte
	//var SIG [G1S]byte
	IKM := make([]byte, 32)

	_, err := r.Read(IKM)
	if err != nil {
		return nil, err
	}

	res := BN254.KeyPairGenerate(IKM[:], S[:], W[:])
	if res != 0 {
		return nil, errors.New("bls_miracl_generate_keys", "Failed to generate keys")
		//	return
	}

	w.Keys[0].PrivateKey = hex.EncodeToString(S[:])
	w.Keys[0].PublicKey = hex.EncodeToString(W[:])

	b0.PrivateKey = w.Keys[0].PrivateKey
	b0.PublicKey = w.Keys[0].PublicKey
	w.ClientKey = w.Keys[0].PublicKey
	w.ClientID = encryption.Hash(W[:])
	w.Mnemonic = b0.Mnemonic
	w.Version = CryptoVersion
	w.DateCreated = time.Now().Format(time.RFC3339)

	return w, nil
}
