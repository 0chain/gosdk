//go:build js && wasm
// +build js,wasm

package zcncrypto

import (
	"encoding/hex"

	"github.com/0chain/errors"
)

var (
	Sign func(hash string) (string, error)
)

// WasmScheme - a signature scheme for BLS0Chain Signature
type WasmScheme struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Mnemonic   string `json:"mnemonic"`

	id  ID
	Ids string `json:"threshold_scheme_id"`
}

// NewWasmScheme - create a BLS0ChainScheme object
func NewWasmScheme() *WasmScheme {
	return &WasmScheme{}
}

func (b0 *WasmScheme) GenerateKeysWithEth(mnemonic, password string) (*Wallet, error) {
	return nil, errors.New("wasm_not_support", "please generate keys by bls_wasm in js")
}

// GenerateKeys - implement interface
func (b0 *WasmScheme) GenerateKeys() (*Wallet, error) {
	return nil, errors.New("wasm_not_support", "please generate keys by bls_wasm in js")
}

func (b0 *WasmScheme) RecoverKeys(mnemonic string) (*Wallet, error) {
	return nil, errors.New("wasm_not_support", "please recover keys by bls_wasm in js")
}

func (b0 *WasmScheme) GetMnemonic() string {
	return ""
}

// SetPrivateKey - implement interface
func (b0 *WasmScheme) SetPrivateKey(privateKey string) error {
	return errors.New("wasm_not_support", "please set keys by bls_wasm in js")
}

// SetPublicKey - implement interface
func (b0 *WasmScheme) SetPublicKey(publicKey string) error {
	return errors.New("wasm_not_support", "please set keys by bls_wasm in js")
}

// GetPublicKey - implement interface
func (b0 *WasmScheme) GetPublicKey() string {
	return "please get key in js"
}

func (b0 *WasmScheme) GetPrivateKey() string {
	return "please get key in js"
}

// Sign - implement interface
func (b0 *WasmScheme) Sign(hash string) (string, error) {
	rawHash, err := hex.DecodeString(hash)
	if err != nil {
		return "", err
	}

	if Sign != nil {
		return Sign(string(rawHash))
	}

	return "", errors.New("wasm_not_initialized", "please init wasm sdk first")
}

// Verify - implement interface
func (b0 *WasmScheme) Verify(signature, msg string) (bool, error) {
	return false, errors.New("wasm_not_support", "please verify signature by bls_wasm in js")
}

func (b0 *WasmScheme) Add(signature, msg string) (string, error) {

	return "", errors.New("wasm_not_support", "aggregate signature is not supported on wasm sdk")
}

// SetID sets ID in HexString format
func (b0 *WasmScheme) SetID(id string) error {
	return errors.New("wasm_not_support", "setid is not supported on wasm sdk")
}

// GetID gets ID in hex string format
func (b0 *WasmScheme) GetID() string {
	return ""
}

// GetPrivateKeyAsByteArray - converts private key into byte array
func (b0 *WasmScheme) GetPrivateKeyAsByteArray() ([]byte, error) {
	return nil, errors.New("wasm_not_support", "please get keys by bls_wasm in js")

}

func (b0 *WasmScheme) SplitKeys(numSplits int) (*Wallet, error) {
	return nil, errors.New("wasm_not_support", "splitkeys is not supported on wasm sdk")
}
