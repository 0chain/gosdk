package zcncrypto

import (
	"fmt"
)

//SignatureScheme - an encryption scheme for signing and verifying messages
type SignatureScheme interface {
	GenerateKeys() error
	GetPublicKey() (string, error)
	GetPublicKeyWithIdx(int) (string, error)
	GetSecretKeyWithIdx(int) (string, error)
	GetMnemonic() (string, error)

	SetPublicKey(publicKey string) error
	SetPrivateKey(privateKey string) error
	RecoverKeys(mnemonic string) error

	Sign(signature string) (string, error)
	Verify(signature string, msg string) (bool, error)
	Add(signature, msg string) (string, error)
}

func NewSignatureScheme(sigScheme string) SignatureScheme {
	switch sigScheme {
	case "ed25519":
		return NewED255190chainScheme()
	case "bls0chain":
		return NewBLS0ChainScheme()
	default:
		panic(fmt.Sprintf("unknown signature scheme: %v", sigScheme))
	}
	return nil
}
