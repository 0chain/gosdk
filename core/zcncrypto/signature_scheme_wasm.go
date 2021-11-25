package zcncrypto

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/bls"
	BN254 "github.com/0chain/gosdk/miracl"

	"github.com/tyler-smith/go-bip39"

	"github.com/0chain/gosdk/core/encryption"
)

func init() {
	err := bls.Init()
	if err != nil {
		panic(err)
	}
}

//WasmScheme - a signature scheme for BLS0Chain Signature
type WasmScheme struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Mnemonic   string `json:"mnemonic"`

	id  bls.ID
	Ids string `json:"threshold_scheme_id"`
}

//NewWasmScheme - create a BLS0ChainScheme object
func NewWasmScheme() *WasmScheme {
	return &WasmScheme{}
}

func (b0 *WasmScheme) GenerateKeysWithEth(mnemonic, password string) (*Wallet, error) {
	if len(mnemonic) == 0 {
		return nil, fmt.Errorf("Mnemonic phrase is mandatory.")
	}
	b0.Mnemonic = mnemonic

	_, err := bip39.NewSeedWithErrorChecking(b0.Mnemonic, password)
	if err != nil {
		return nil, fmt.Errorf("Wrong mnemonic phrase.")
	}

	return b0.generateKeys(password)
}

//GenerateKeys - implement interface
func (b0 *WasmScheme) GenerateKeys() (*Wallet, error) {
	return b0.generateKeys("0chain-client-split-key")
}

func (b0 *WasmScheme) generateKeys(password string) (*Wallet, error) {
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

func (b0 *WasmScheme) RecoverKeys(mnemonic string) (*Wallet, error) {
	if mnemonic == "" {
		return nil, errors.New("recover_keys", "Set mnemonic key failed")
	}
	if b0.PublicKey != "" || b0.PrivateKey != "" {
		return nil, errors.New("recover_keys", "Cannot recover when there are keys")
	}
	b0.Mnemonic = mnemonic
	return b0.GenerateKeys()
}

func (b0 *WasmScheme) GetMnemonic() string {
	if b0 == nil {
		return ""
	}

	return b0.Mnemonic
}

//SetPrivateKey - implement interface
func (b0 *WasmScheme) SetPrivateKey(privateKey string) error {
	if b0.PublicKey != "" {
		return errors.New("set_private_key", "cannot set private key when there is a public key")
	}
	if b0.PrivateKey != "" {
		return errors.New("set_private_key", "private key already exists")
	}
	b0.PrivateKey = privateKey

	var sk bls.SecretKey
	sk.DeserializeHexStr(b0.PrivateKey)

	b0.PrivateKey = sk.SerializeToHexStr()
	//ToDo: b0.publicKey should be set here?
	return nil
}

//SetPublicKey - implement interface
func (b0 *WasmScheme) SetPublicKey(publicKey string) error {
	if b0.PrivateKey != "" {
		return errors.New("set_public_key", "cannot set public key when there is a private key")
	}
	if b0.PublicKey != "" {
		return errors.New("set_public_key", "public key already exists")
	}
	b0.PublicKey = publicKey

	// TODO: remove this line once we are sure nothing was depending on this
	// Miracl->Herumi conversion.
	// b0.PublicKey = MiraclToHerumiPK(publicKey)

	return nil
}

// TODO:
// 1a) find whatever repo had dependency on gosdk's MiraclToHerumiPK func.
// 1b) replace their dependency with this function code in that repo maybe.
// 2) remove this code
//
// this gosdk's MiraclToHerumiPK function needs to be replaced wherever
// it is used with a local version, so that gosdk is able to compile without
// C++ dependencies.
//
// // Converts public key 'pk' to format that the herumi/bls library likes.
// // It's possible to get a MIRACL PublicKey which is of much longer format
// // (See below example), as wallets are using MIRACL library not herumi lib.
// // If 'pk' is not in MIRACL format, we just return the original 'pk' then.
// //
// // This is an example of the raw public key we expect from MIRACL
// var miraclExamplePK = `0418a02c6bd223ae0dfda1d2f9a3c81726ab436ce5e9d17c531ff0a385a13a0b491bdfed3a85690775ee35c61678957aaba7b1a1899438829f1dc94248d87ed36817f6dfafec19bfa87bf791a4d694f43fec227ae6f5a867490e30328cac05eaff039ac7dfc3364e851ebd2631ea6f1685609fc66d50223cc696cb59ff2fee47ac`
//
// //
// // This is an example of the same MIRACL public key serialized with ToString().
// // pk ([1bdfed3a85690775ee35c61678957aaba7b1a1899438829f1dc94248d87ed368,18a02c6bd223ae0dfda1d2f9a3c81726ab436ce5e9d17c531ff0a385a13a0b49],[039ac7dfc3364e851ebd2631ea6f1685609fc66d50223cc696cb59ff2fee47ac,17f6dfafec19bfa87bf791a4d694f43fec227ae6f5a867490e30328cac05eaff])
// func MiraclToHerumiPK(pk string) string {
// 	if len(pk) != len(miraclExamplePK) {
// 		return pk
// 	}
// 	fmt.Println(">> pk", pk)
// 	n1 := pk[2:66]
// 	n2 := pk[66:(66 + 64)]
// 	n3 := pk[(66 + 64):(66 + 64 + 64)]
// 	n4 := pk[(66 + 64 + 64):(66 + 64 + 64 + 64)]
// 	var p bls.PublicKey
// 	p.SetHexString("1 " + n2 + " " + n1 + " " + n4 + " " + n3)
// 	fmt.Println(">> bp1")
// 	return p.SerializeToHexStr()
// }

//GetPublicKey - implement interface
func (b0 *WasmScheme) GetPublicKey() string {
	return b0.PublicKey
}

func (b0 *WasmScheme) GetPrivateKey() string {
	return b0.PrivateKey
}

func (b0 *WasmScheme) rawSign(hash string) (*bls.Sign, error) {
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

	// My port.
	var sk bls.SecretKey
	sk.SetByCSPRNG()
	sk.DeserializeHexStr(b0.PrivateKey)
	sig := sk.Sign(rawHash)
	return sig, nil
}

//Sign - implement interface
func (b0 *WasmScheme) Sign(hash string) (string, error) {
	sig, err := b0.rawSign(hash)
	if err != nil {
		return "", err
	}
	return sig.SerializeToHexStr(), nil
}

//Verify - implement interface
func (b0 *WasmScheme) Verify(signature, msg string) (bool, error) {
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
	return sig.Verify(&pk, rawHash), nil
}

func (b0 *WasmScheme) Add(signature, msg string) (string, error) {
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

//SetID sets ID in HexString format
func (b0 *WasmScheme) SetID(id string) error {
	b0.Ids = id
	return b0.id.SetHexString(id)
}

//GetID gets ID in hex string format
func (b0 *WasmScheme) GetID() string {
	return b0.id.GetHexString()
}

// GetPrivateKeyAsByteArray - converts private key into byte array
func (b0 *WasmScheme) GetPrivateKeyAsByteArray() ([]byte, error) {
	if len(b0.PrivateKey) == 0 {
		return nil, errors.New("get_private_key_as_byte_array", "cannot convert empty private key to byte array")
	}
	privateKeyBytes, err := hex.DecodeString(b0.PrivateKey)
	if err != nil {
		return nil, err
	}
	return privateKeyBytes, nil

}

func (b0 *WasmScheme) SplitKeys(numSplits int) (*Wallet, error) {
	if b0.PrivateKey == "" {
		return nil, errors.New("split_keys", "primary private key not found")
	}

	var primarySk bls.SecretKey
	primarySk.DeserializeHexStr(b0.PrivateKey)
	limit := BN254.NewBIGcopy(primarySk.GetBIG())
	limit.Div(BN254.NewBIGint(numSplits))

	// New Wallet
	w := &Wallet{}
	w.Keys = make([]KeyPair, numSplits)
	aggregateSk := BN254.NewBIG()
	for i := 0; i < numSplits-1; i++ {
		var tmpSk bls.SecretKey
		tmpSk.SetByCSPRNG()

		// It is extremely important that aggregateSk < lastSk.
		// We can ensure this by capping every tmpSk to lastSk/n
		for BN254.Comp(limit, tmpSk.GetBIG()) < 0 {
			tmpSk.SetByCSPRNG()
		}

		w.Keys[i].PrivateKey = tmpSk.SerializeToHexStr()
		w.Keys[i].PublicKey = tmpSk.GetPublicKey().SerializeToHexStr()
		aggregateSk.Add(tmpSk.GetBIG())
	}

	// Subtract the aggregated private key from the primary private key to derive
	// the last split private key
	lastSk := primarySk.GetBIG()
	lastSk.Sub(aggregateSk)

	// Last key
	lastSecretKey := bls.SecretKey_fromBIG(lastSk)
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
