package zcncrypto

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"0chain/gosdk/encryption"
	"github.com/herumi/bls/ffi/go/bls"
	"github.com/tyler-smith/go-bip39"
)

const numSplitKeys = 2

var GenG2 *bls.G2
var ErrKeyRead = errors.New("error reading the keys")
var ErrInvalidSignatureScheme = errors.New("invalid signature scheme")

func init() {
	err := bls.Init(bls.CurveFp254BNb)
	if err != nil {
		panic(err)
	}
	GenG2 = &bls.G2{}
	/* The following string is obtained by serializing the generator of G2 using temporary go binding as follows
		func (pub1 *PublicKey) GenG2() (pub2 *PublicKey) {
	        pub2 = new(PublicKey)
	        C.blsGetGeneratorOfG2(pub2.getPointer())
	        return pub2
	} */
	bytes, err := hex.DecodeString("28b1ce2dbb7eccc8ba6b0d29615ac81e33be4d5909602ac35d2cac774eb4cc119a0deec914a95ffcd4cdbe685608602e7f82de7651a2e95ba0c4dabb144a200f")
	if err != nil {
		panic(err)
	}
	GenG2.Deserialize(bytes)
}

//BLS0ChainScheme - a signature scheme for BLS0Chain Signature
type BLS0ChainScheme struct {
	sk       []string // Secret/Private key
	pk       []string // PublicKey
	mnemonic string
}

//NewBLS0ChainScheme - create a BLS0ChainScheme object
func NewBLS0ChainScheme() *BLS0ChainScheme {
	return &BLS0ChainScheme{}
}

//GenerateKeys - implement interface
func (b0 *BLS0ChainScheme) GenerateKeys() error {
	// Check for recovery
	if len(b0.mnemonic) == 0 {
		entropy, err := bip39.NewEntropy(256)
		if err != nil {
			return fmt.Errorf("Getting entropy failed")
		}
		b0.mnemonic, err = bip39.NewMnemonic(entropy)
		if err != nil {
			return fmt.Errorf("Getting mnemonic failed")
		}
	}

	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(b0.mnemonic, "0chain-client-split-key")
	var sk0, sk1 bls.SecretKey
	r := bytes.NewReader(seed)
	bls.SetRandFunc(r)
	sk0.SetByCSPRNG()
	sk1.SetByCSPRNG()

	b0.sk = make([]string, numSplitKeys)
	b0.pk = make([]string, numSplitKeys)
	b0.sk[0] = sk0.SerializeToHexStr()
	b0.sk[1] = sk1.SerializeToHexStr()
	b0.pk[0] = sk0.GetPublicKey().SerializeToHexStr()
	b0.pk[1] = sk1.GetPublicKey().SerializeToHexStr()
	bls.SetRandFunc(nil)
	return nil
}

func (b0 *BLS0ChainScheme) RecoverKeys(mnemonic string) error {
	if mnemonic == "" {
		return fmt.Errorf("Set mnemonic key failed")
	}
	if len(b0.pk) > 0 || len(b0.sk) > 0 {
		return errors.New("Cannot recover when there are keys")
	}
	b0.mnemonic = mnemonic
	b0.GenerateKeys()
	return nil
}

func (b0 *BLS0ChainScheme) GetPublicKey() (string, error) {
	if len(b0.sk) == 0 {
		return "", fmt.Errorf("Key Not Found")
	}
	// Get Public key for blockchain for GenerateKeys
	if len(b0.sk) == numSplitKeys {
		var pk0, pk1 bls.PublicKey
		pk0.DeserializeHexStr(b0.pk[0])
		pk1.DeserializeHexStr(b0.pk[1])
		pk0.Add(&pk1)
		return pk0.SerializeToHexStr(), nil
	}
	return b0.pk[0], nil
}

func (b0 *BLS0ChainScheme) GetPublicKeyWithIdx(i int) (string, error) {
	if i >= numSplitKeys {
		return "", fmt.Errorf("Invalid Key Index %d", i)
	}
	if len(b0.sk) <= i {
		return "", fmt.Errorf("Key Not Found")
	}
	return b0.pk[i], nil
}

func (b0 *BLS0ChainScheme) GetSecretKeyWithIdx(i int) (string, error) {
	if i >= numSplitKeys {
		return "", fmt.Errorf("Invalid Key Index %d", i)
	}
	if len(b0.sk) <= i {
		return "", fmt.Errorf("Key Not Found")
	}
	return b0.sk[i], nil
}

func (b0 *BLS0ChainScheme) GetMnemonic() (string, error) {
	if len(b0.mnemonic) == 0 {
		return "", fmt.Errorf("Mnemonic found only if key is generated")
	}
	return b0.mnemonic, nil
}

//SetPrivateKey - implement interface
func (b0 *BLS0ChainScheme) SetPrivateKey(privateKey string) error {
	if len(b0.pk) > 0 {
		return errors.New("Cannot set private key when there is a public key")
	}
	if len(b0.sk) > 0 {
		return errors.New("Private key already exists")
	}
	b0.sk = make([]string, 1)
	b0.sk[0] = privateKey
	return nil
}

//SetPublicKey - implement interface
func (b0 *BLS0ChainScheme) SetPublicKey(publicKey string) error {
	if len(b0.sk) > 0 {
		return errors.New("cannot set public key when there is a private key")
	}
	if len(b0.pk) > 0 {
		return errors.New("Public key already exists")
	}
	b0.pk = make([]string, 1)
	b0.pk[0] = publicKey
	return nil
}

func (b0 *BLS0ChainScheme) rawSign(hash string) (*bls.Sign, error) {
	var sk bls.SecretKey
	sk.SetByCSPRNG()
	if len(b0.sk[0]) == 0 {
		return &bls.Sign{}, errors.New("private key does not exists for signing")
	}
	sk.DeserializeHexStr(b0.sk[0])
	rawHash := encryption.RawHash(hash)
	if rawHash == nil {
		return &bls.Sign{}, fmt.Errorf("Failed hash while signing")
	}
	sig := sk.Sign(hex.EncodeToString(rawHash))
	return sig, nil
}

//Sign - implement interface
func (b0 *BLS0ChainScheme) Sign(hash string) (string, error) {
	sig, err := b0.rawSign(hash)
	if err != nil {
		return "", err
	}
	return sig.SerializeToHexStr(), nil
}

//Verify - implement interface
func (b0 *BLS0ChainScheme) Verify(signature, msg string) (bool, error) {
	if len(b0.pk[0]) == 0 {
		return false, errors.New("public key does not exists for verification")
	}
	var sig bls.Sign
	var pk bls.PublicKey
	err := sig.DeserializeHexStr(signature)
	if err != nil {
		return false, err
	}
	pk.DeserializeHexStr(b0.pk[0])
	return sig.Verify(&pk, encryption.Hash(msg)), nil
}

func (b0 *BLS0ChainScheme) Add(signature, msg string) (string, error) {
	var sign bls.Sign
	err := sign.DeserializeHexStr(signature)
	if err != nil {
		return "", err
	}
	signature1, err := b0.rawSign(msg)
	if err != nil {
		return "", fmt.Errorf("BLS signing failed - %s", err.Error())
	}
	sign.Add(signature1)
	return sign.SerializeToHexStr(), nil
}
