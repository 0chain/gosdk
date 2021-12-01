//go:build js && wasm
// +build js,wasm

package zcncrypto

import (
	"encoding/json"
	"fmt"

	"github.com/0chain/errors"
)

// NewSignatureScheme creates an instance for using signature functions
func NewSignatureScheme(sigScheme string) SignatureScheme {
	switch sigScheme {
	case "ed25519":
		return NewED255190chainScheme()
	case "bls0chain":
		return NewWasmScheme()
	default:
		panic(fmt.Sprintf("unknown signature scheme: %v", sigScheme))
	}
}

// UnmarshalSignatureSchemes unmarshal SignatureScheme from json string
func UnmarshalSignatureSchemes(sigScheme string, obj interface{}) ([]SignatureScheme, error) {
	switch sigScheme {

	case "bls0chain":

		if obj == nil {
			return nil, nil
		}

		buf, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}

		var list []*WasmScheme

		if err := json.Unmarshal(buf, &list); err != nil {
			return nil, err
		}

		ss := make([]SignatureScheme, len(list))

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
func GenerateThresholdKeyShares(t, n int, originalKey SignatureScheme) ([]SignatureScheme, error) {

	return nil, errors.New("wasm_not_supported", "GenerateThresholdKeyShares")
}
