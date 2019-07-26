package zcncrypto

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/herumi/bls/ffi/go/bls"
	"github.com/tyler-smith/go-bip39"
)

func init() {
	err := bls.Init(bls.CurveFp254BNb)
	if err != nil {
		panic(err)
	}
}

//BLS0ChainScheme - a signature scheme for BLS0Chain Signature
type BLS0ChainScheme struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Mnemonic   string `json:"mnemonic"`
}

//NewBLS0ChainScheme - create a BLS0ChainScheme object
func NewBLS0ChainScheme() *BLS0ChainScheme {
	return &BLS0ChainScheme{}
}

//GenerateKeys - implement interface
func (b0 *BLS0ChainScheme) GenerateKeys() (*Wallet, error) {
	// Check for recovery
	if len(b0.Mnemonic) == 0 {
		entropy, err := bip39.NewEntropy(256)
		if err != nil {
			return nil, fmt.Errorf("Generating entropy failed")
		}
		b0.Mnemonic, err = bip39.NewMnemonic(entropy)
		if err != nil {
			return nil, fmt.Errorf("Generating mnemonic failed")
		}
	}

	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(b0.Mnemonic, "0chain-client-split-key")
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
	w.DateCreated = time.Now().String()

	// Revert the Random function to default
	bls.SetRandFunc(nil)
	return w, nil
}

func (b0 *BLS0ChainScheme) RecoverKeys(mnemonic string) (*Wallet, error) {
	if mnemonic == "" {
		return nil, fmt.Errorf("Set mnemonic key failed")
	}
	if b0.PublicKey != "" || b0.PrivateKey != "" {
		return nil, errors.New("Cannot recover when there are keys")
	}
	b0.Mnemonic = mnemonic
	return b0.GenerateKeys()
}

//SetPrivateKey - implement interface
func (b0 *BLS0ChainScheme) SetPrivateKey(privateKey string) error {
	if b0.PublicKey != "" {
		return errors.New("cannot set private key when there is a public key")
	}
	if b0.PrivateKey != "" {
		return errors.New("private key already exists")
	}
	b0.PrivateKey = privateKey
	//ToDo: b0.publicKey should be set here?
	return nil
}

//SetPublicKey - implement interface
func (b0 *BLS0ChainScheme) SetPublicKey(publicKey string) error {
	if b0.PrivateKey != "" {
		return errors.New("cannot set public key when there is a private key")
	}
	if b0.PublicKey != "" {
		return errors.New("public key already exists")
	}
	b0.PublicKey = publicKey
	return nil
}

//GetPublicKey - implement interface
func (b0 *BLS0ChainScheme) GetPublicKey() string {
	return b0.PublicKey
}

func (b0 *BLS0ChainScheme) GetPrivateKey() string {
	return b0.PrivateKey
}

func (b0 *BLS0ChainScheme) rawSign(hash string) (*bls.Sign, error) {
	var sk bls.SecretKey
	if b0.PrivateKey == "" {
		return nil, errors.New("private key does not exists for signing")
	}
	rawHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}
	if rawHash == nil {
		return nil, errors.New("failed hash while signing")
	}
	sk.SetByCSPRNG()
	sk.DeserializeHexStr(b0.PrivateKey)
	sig := sk.Sign(string(rawHash))
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
	if b0.PublicKey == "" {
		return false, errors.New("public key does not exists for verification")
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
		return false, errors.New("failed hash while signing")
	}
	pk.DeserializeHexStr(b0.PublicKey)
	return sig.Verify(&pk, string(rawHash)), nil
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

type ThresholdSignatureScheme interface {
	SignatureScheme

	SetID(id string) error
	GetID() string
}

//BLS0ChainThresholdScheme - a scheme that can create threshold signature shares for BLS0Chain signature scheme
type BLS0ChainThresholdScheme struct {
	BLS0ChainScheme
	id  bls.ID
	Ids string `json:"threshold_scheme_id"`
}

//NewBLS0ChainThresholdScheme - create a new instance
func NewBLS0ChainThresholdScheme() *BLS0ChainThresholdScheme {
	return &BLS0ChainThresholdScheme{}
}

//SetID sets ID in HexString format
func (tss *BLS0ChainThresholdScheme) SetID(id string) error {
	tss.Ids = id
	return tss.id.SetHexString(id)
}

//GetID gets ID in hex string format
func (tss *BLS0ChainThresholdScheme) GetID() string {
	return tss.id.GetHexString()
}

// GetPrivateKeyAsByteArray - converts private key into byte array
func (b0 *BLS0ChainScheme) GetPrivateKeyAsByteArray() ([]byte, error) {
	if len(b0.PrivateKey) == 0 {
		return nil, errors.New("cannot convert empty private key to byte array")
	}
	privateKeyBytes, err := hex.DecodeString(b0.PrivateKey)
	if err != nil {
		return nil, err
	}
	return privateKeyBytes, nil

}

//BLS0GenerateThresholdKeyShares given a signature scheme will generate threshold sig keys
func BLS0GenerateThresholdKeyShares(t, n int, originalKey SignatureScheme) ([]BLS0ChainThresholdScheme, error) {

	b0ss, ok := originalKey.(*BLS0ChainScheme)
	if !ok {
		return nil, errors.New("Invalid encryption scheme")
	}

	var b0original bls.SecretKey
	b0PrivateKeyBytes, err := b0ss.GetPrivateKeyAsByteArray()
	if err != nil {
		return nil, err
	}

	err = b0original.SetLittleEndian(b0PrivateKeyBytes)
	if err != nil {
		return nil, err
	}

	polynomial := b0original.GetMasterSecretKey(t)

	var shares []BLS0ChainThresholdScheme
	for i := 1; i <= n; i++ {
		var id bls.ID
		err = id.SetDecString(fmt.Sprint(i))
		if err != nil {
			return nil, err
		}

		var sk bls.SecretKey
		err = sk.Set(polynomial, &id)
		if err != nil {
			return nil, err
		}

		share := BLS0ChainThresholdScheme{}
		share.PrivateKey = hex.EncodeToString(sk.GetLittleEndian())
		share.PublicKey = sk.GetPublicKey().SerializeToHexStr()

		share.id = id
		share.Ids = share.GetID()

		shares = append(shares, share)
	}

	return shares, nil
}

func (b0 *BLS0ChainScheme) SplitKeys(numSplits int) (*Wallet, error) {
	if b0.PrivateKey == "" {
		return nil, errors.New("primary private key not found")
	}
	var primaryFr bls.Fr
	var primarySk bls.SecretKey
	primarySk.DeserializeHexStr(b0.PrivateKey)
	primaryFr.SetLittleEndian(primarySk.GetLittleEndian())

	// New Wallet
	w := &Wallet{}
	w.Keys = make([]KeyPair, numSplits)
	var sk bls.SecretKey
	for i := 0; i < numSplits-1; i++ {
		var tmpSk bls.SecretKey
		tmpSk.SetByCSPRNG()
		w.Keys[i].PrivateKey = tmpSk.SerializeToHexStr()
		pub := tmpSk.GetPublicKey()
		w.Keys[i].PublicKey = pub.SerializeToHexStr()
		sk.Add(&tmpSk)
	}
	var aggregateSk bls.Fr
	aggregateSk.SetLittleEndian(sk.GetLittleEndian())

	//Subtract the aggregated private key from the primary private key to derive the last split private key
	var lastSk bls.Fr
	bls.FrSub(&lastSk, &primaryFr, &aggregateSk)

	// Last key
	var lastSecretKey bls.SecretKey
	lastSecretKey.SetLittleEndian(lastSk.Serialize())
	w.Keys[numSplits-1].PrivateKey = lastSecretKey.SerializeToHexStr()
	w.Keys[numSplits-1].PublicKey = lastSecretKey.GetPublicKey().SerializeToHexStr()

	// Generate client ID and public
	w.ClientKey = primarySk.GetPublicKey().SerializeToHexStr()
	w.ClientID = encryption.Hash(primarySk.GetPublicKey().Serialize())
	w.Mnemonic = b0.Mnemonic
	w.Version = CryptoVersion
	w.DateCreated = time.Now().String()

	return w, nil
}
