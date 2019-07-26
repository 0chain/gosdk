package zcncrypto

import (
	"encoding/hex"
	"testing"
)

var edverifyPublickey = `b987071c14695caf340ea11560f5a3cb76ad1e709803a8b339826ab3964e470a`
var edsignPrivatekey = `62fc118369fb9dd1fa6065d4f8f765c52ac68ad5aced17a1e5c4f8b4301a9469b987071c14695caf340ea11560f5a3cb76ad1e709803a8b339826ab3964e470a`
var eddata = `TEST`
var edexpectedHash = `f4f08e9367e133dc42a4b9c9c665a9efbd4bf15db14d49c6ec51d0dc4c437ffb`
var edWallet *Wallet

func TestEd25519GenerateKeys(t *testing.T) {
	sigScheme := NewSignatureScheme("ed25519")
	switch sigScheme.(type) {
	case SignatureScheme:
		// pass
	default:
		t.Fatalf("Signature scheme invalid")
	}
	w, err := sigScheme.GenerateKeys()
	if err != nil {
		t.Fatalf("Generate keys failed %s", err.Error())
	}
	if w.ClientID == "" || w.ClientKey == "" || len(w.Keys) != 1 || w.Mnemonic == "" {
		t.Fatalf("Invalid keys generated")
	}
	edWallet = w
}

func TestEd25519SignAndVerify(t *testing.T) {
	signScheme := NewSignatureScheme("ed25519")
	// Check failure without private key
	signature, err := signScheme.Sign(eddata)
	if err == nil {
		t.Fatalf("Sign passed without private key")
	}
	// Sign with valid private key
	signScheme.SetPrivateKey(edsignPrivatekey)
	signature, err = signScheme.Sign(hex.EncodeToString([]byte(eddata)))
	if err != nil {
		t.Fatalf("ed25519 signing failed")
	}
	verifyScheme := NewSignatureScheme("ed25519")
	verifyScheme.SetPublicKey(edverifyPublickey)
	if ok, err := verifyScheme.Verify(signature, hex.EncodeToString([]byte(eddata))); err != nil || !ok {
		t.Fatalf("Verification failed\n")
	}
}

func TestEd25519RecoveryKeys(t *testing.T) {
	sigScheme := NewSignatureScheme("ed25519")
	w, err := sigScheme.RecoverKeys(edWallet.Mnemonic)
	if err != nil {
		t.Fatalf("set Recover Keys failed")
	}
	if w.ClientID != edWallet.ClientID || w.ClientKey != edWallet.ClientKey {
		t.Fatalf("Recover key didn't match with generated keys")
	}
}

func BenchmarkE25519Signandverify(b *testing.B) {
	sigScheme := NewSignatureScheme("ed25519")
	sigScheme.SetPrivateKey(edsignPrivatekey)
	for i := 0; i < b.N; i++ {
		signature, err := sigScheme.Sign(eddata)
		if err != nil {
			b.Fatalf("BLS signing failed")
		}
		verifyScheme := NewSignatureScheme("ed25519")
		verifyScheme.SetPublicKey(edverifyPublickey)
		if ok, err := verifyScheme.Verify(signature, eddata); err != nil || !ok {
			b.Fatalf("Verification failed\n")
		}
	}
}
