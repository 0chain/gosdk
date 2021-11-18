// +build !js,!wasm

package zcncrypto

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/0chain/errors"
	"github.com/herumi/bls-go-binary/bls"
)

// NewSignatureScheme creates an instance for using signature functions
func NewSignatureScheme(sigScheme string) SignatureScheme {
	switch sigScheme {
	case "ed25519":
		return NewED255190chainScheme()
	case "bls0chain":
		//return NewBLS0ChainScheme()
		return NewMiraclScheme()
	default:
		panic(fmt.Sprintf("unknown signature scheme: %v", sigScheme))
	}
}

// UnmarshalThresholdSignatureSchemes unmarshal ThresholdSignatureScheme from json string
func UnmarshalThresholdSignatureSchemes(sigScheme string, obj interface{}) ([]ThresholdSignatureScheme, error) {
	switch sigScheme {

	case "bls0chain":

		if obj == nil {
			return nil, nil
		}

		buf, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}

		var list []*MiraclThresholdScheme

		if err := json.Unmarshal(buf, &list); err != nil {
			return nil, err
		}

		ss := make([]ThresholdSignatureScheme, len(list))

		for i, v := range list {
			// bls.ID from json
			v.SetID(v.Ids)
			ss[i] = v
		}

		return ss, nil

	default:
		panic(fmt.Sprintf("unknown signature scheme: %v", sigScheme))
	}
}

//GenerateThresholdKeyShares given a signature scheme will generate threshold sig keys
func GenerateThresholdKeyShares(t, n int, originalKey SignatureScheme) ([]ThresholdSignatureScheme, error) {

	b0ss, ok := originalKey.(*MiraclScheme)
	if !ok {
		return nil, errors.New("bls0_generate_threshold_key_shares", "Invalid encryption scheme")
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

	var shares []ThresholdSignatureScheme
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

		share := &MiraclThresholdScheme{}
		share.PrivateKey = hex.EncodeToString(sk.GetLittleEndian())
		share.PublicKey = sk.GetPublicKey().SerializeToHexStr()

		share.id = id
		share.Ids = id.GetHexString()

		shares = append(shares, share)
	}

	return shares, nil
}
