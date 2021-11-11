// +build js,wasm

package zcncrypto

import (
	"encoding/json"
	"fmt"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/bls"
)

// NewSignatureScheme creates an instance for using signature functions
func NewSignatureScheme(sigScheme string) SignatureScheme {
	switch sigScheme {
	case "ed25519":
		return NewED255190chainScheme()
	case "bls0chain":
		return NewBLS0ChainScheme()
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

		var list []*BLS0ChainThresholdScheme

		if err := json.Unmarshal(buf, &list); err != nil {
			return nil, err
		}

		ss := make([]ThresholdSignatureScheme, len(list))

		for i, v := range list {
			ss[i] = v
		}

		return ss, nil

	default:
		panic(fmt.Sprintf("unknown signature scheme: %v", sigScheme))
	}
}

//GenerateThresholdKeyShares given a signature scheme will generate threshold sig keys
func GenerateThresholdKeyShares(t, n int, originalKey SignatureScheme) ([]ThresholdSignatureScheme, error) {

	b0ss, ok := originalKey.(*BLS0ChainScheme)
	if !ok {
		return nil, errors.New("bls0_generate_threshold_key_shares", "Invalid encryption scheme")
	}

	b0PrivateKeyBytes, err := b0ss.GetPrivateKeyAsByteArray()
	if err != nil {
		return nil, err
	}

	b0original := bls.SecretKey_fromBytes(b0PrivateKeyBytes)
	polynomial := b0original.GetMasterSecretKey(t)

	var shares []ThresholdSignatureScheme
	for i := 1; i <= n; i++ {
		var id bls.ID
		err = id.SetHexString(fmt.Sprintf("%x", i))
		if err != nil {
			return nil, err
		}

		var sk bls.SecretKey
		err = sk.Set(polynomial, &id)
		if err != nil {
			return nil, err
		}

		share := &BLS0ChainThresholdScheme{}
		share.PrivateKey = sk.SerializeToHexStr()
		share.PublicKey = sk.GetPublicKey().SerializeToHexStr()

		share.id = id
		share.Ids = id.GetHexString()

		shares = append(shares, share)
	}

	return shares, nil
}
