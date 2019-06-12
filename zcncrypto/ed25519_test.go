package zcncrypto

import (
	"testing"
)

var edverifyPublickey = `b987071c14695caf340ea11560f5a3cb76ad1e709803a8b339826ab3964e470a`
var edsignPrivatekey = `62fc118369fb9dd1fa6065d4f8f765c52ac68ad5aced17a1e5c4f8b4301a9469b987071c14695caf340ea11560f5a3cb76ad1e709803a8b339826ab3964e470a`
var eddata = `TEST`
var edexpectedHash = `f4f08e9367e133dc42a4b9c9c665a9efbd4bf15db14d49c6ec51d0dc4c437ffb`

func TestEd25519GenerateKeys(t *testing.T) {
	sigScheme := NewSignatureScheme("ed25519")
	switch sigScheme.(type) {
	case SignatureScheme:
		// pass
	default:
		t.Fatalf("Signature scheme invalid")
	}
	err := sigScheme.GenerateKeys()
	if err != nil {
		t.Fatalf("Generate keys failed %s", err.Error())
	}
	public, err := sigScheme.GetPublicKey()
	if err != nil {
		t.Fatalf("Get public key failed - %s", err.Error())
	}
	private, err := sigScheme.GetSecretKeyWithIdx(0)
	if err != nil {
		t.Fatalf("Get secret key failed - %s", err.Error())
	}
	//Get mnemonic and verify recover keys works
	mn, err := sigScheme.GetMnemonic()
	if err != nil {
		t.Fatalf("Get Mnemonic failed - %s", err.Error())
	}
	err = sigScheme.RecoverKeys(mn)
	if err == nil {
		t.Fatalf("Recover Keys failed - %s", err.Error())
	}
	rec := NewSignatureScheme("ed25519")
	err = rec.RecoverKeys(mn)
	if err != nil {
		t.Fatalf("Recover keys failed - %s", err.Error())
	}
	mnpublic, err := rec.GetPublicKey()
	if err != nil {
		t.Fatalf("Get public key failed - %s", err.Error())
	}
	mnprivate, err := rec.GetSecretKeyWithIdx(0)
	if err != nil {
		t.Fatalf("Get secret key failed - %s", err.Error())
	}
	if public != mnpublic || private != mnprivate {
		t.Fatalf("Recovered keys does not match")
	}
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
	signature, err = signScheme.Sign(eddata)
	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	verifyScheme := NewSignatureScheme("ed25519")
	verifyScheme.SetPublicKey(edverifyPublickey)
	if ok, err := verifyScheme.Verify(signature, eddata); err != nil || !ok {
		t.Fatalf("Verification failed\n")
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
