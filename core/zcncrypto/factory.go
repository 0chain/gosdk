package zcncrypto

import "fmt"

// NewSignatureScheme creates an instance for using signature functions
func NewSignatureScheme(sigScheme string) SignatureScheme {
	switch sigScheme {
	case "ed25519":
		return NewED255190chainScheme()
	case "bls0chain":
		//return NewBLS0ChainScheme()
		return NewMircalScheme()
	default:
		panic(fmt.Sprintf("unknown signature scheme: %v", sigScheme))
	}
}
