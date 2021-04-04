package zcncrypto

import (
	"testing"

	"github.com/0chain/gosdk/core/encryption"
)

var verifyPublickey = `e8a6cfa7b3076ae7e04764ffdfe341632a136b52953dfafa6926361dd9a466196faecca6f696774bbd64b938ff765dbc837e8766a5e2d8996745b2b94e1beb9e`
var signPrivatekey = `5e1fc9c03d53a8b9a63030acc2864f0c33dffddb3c276bf2b3c8d739269cc018`
var data = `TEST`
var blsWallet *Wallet

func TestSignatureScheme(t *testing.T) {
	sigScheme := NewSignatureScheme("bls0chain")
	switch sigScheme.(type) {
	case SignatureScheme:
		// pass
	default:
		t.Fatalf("Signature scheme invalid")
	}
	w, err := sigScheme.GenerateKeys()
	if err != nil {
		t.Fatalf("Generate Key failed %s", err.Error())
	}
	if w.ClientID == "" || w.ClientKey == "" || len(w.Keys) != 1 || w.Mnemonic == "" {
		t.Fatalf("Invalid keys generated")
	}
	blsWallet = w
}

func TestSSSignAndVerify(t *testing.T) {
	signScheme := NewSignatureScheme("bls0chain")
	signScheme.SetPrivateKey(signPrivatekey)
	hash := Sha3Sum256(data)
	signature, err := signScheme.Sign(hash)
	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	verifyScheme := NewSignatureScheme("bls0chain")
	verifyScheme.SetPublicKey(verifyPublickey)
	if ok, err := verifyScheme.Verify(signature, hash); err != nil || !ok {
		t.Fatalf("Verification failed\n")
	}
}

func BenchmarkBLSSign(b *testing.B) {
	sigScheme := NewSignatureScheme("bls0chain")
	sigScheme.SetPrivateKey(signPrivatekey)
	for i := 0; i < b.N; i++ {
		_, err := sigScheme.Sign(encryption.Hash(data))
		if err != nil {
			b.Fatalf("BLS signing failed")
		}
	}
}

func TestRecoveryKeys(t *testing.T) {

	sigScheme := NewSignatureScheme("bls0chain")
	TestSignatureScheme(t)
	w, err := sigScheme.RecoverKeys(blsWallet.Mnemonic)
	if err != nil {
		t.Fatalf("set Recover Keys failed")
	}
	if w.ClientID != blsWallet.ClientID || w.ClientKey != blsWallet.ClientKey {
		t.Fatalf("Recover key didn't match with generated keys")
	}

}

func TestCombinedSignAndVerify(t *testing.T) {
	sk0 := `c36f2f92b673cf057a32e8bd0ca88888e7ace40337b737e9c7459fdc4c521918`
	sk1 := `704b6f489583bf1118432fcfb38e63fc2d4b61e524fb196cbd95413f8eb91c12`
	primaryKey := `f72fd53ee85e84157d3106053754594f697e0bfca1f73f91a41f7bb0797d901acefd80fcc2da98aae690af0ee9c795d6590c1808f26490306433b4e9c42f7b1f`

	hash := Sha3Sum256(data)
	// Create signatue for 1
	sig0 := NewSignatureScheme("bls0chain")
	err := sig0.SetPrivateKey(sk0)
	if err != nil {
		t.Fatalf("Set private key failed - %s", err.Error())
	}
	signature, err := sig0.Sign(hash)
	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	// Create signature for second
	sig1 := NewSignatureScheme("bls0chain")
	err = sig1.SetPrivateKey(sk1)
	if err != nil {
		t.Fatalf("Set private key failed - %s", err.Error())
	}
	addSig, err := sig1.Add(signature, hash)

	verifyScheme := NewSignatureScheme("bls0chain")
	err = verifyScheme.SetPublicKey(primaryKey)
	if err != nil {
		t.Fatalf("Set public key failed")
	}
	if ok, err := verifyScheme.Verify(addSig, hash); err != nil || !ok {
		t.Fatalf("Verification failed\n")
	}
}

func TestSplitKey(t *testing.T) {
	primaryKeyStr := `c36f2f92b673cf057a32e8bd0ca88888e7ace40337b737e9c7459fdc4c521918`
	sig0 := NewBLS0ChainScheme()
	err := sig0.SetPrivateKey(primaryKeyStr)
	if err != nil {
		t.Fatalf("Set private key failed - %s", err.Error())
	}
	hash := Sha3Sum256(data)
	signature, err := sig0.Sign(hash)
	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	numSplitKeys := int(2)
	w, err := sig0.SplitKeys(numSplitKeys)
	if err != nil {
		t.Fatalf("Splitkeys key failed - %s", err.Error())
	}
	sigAggScheme := make([]BLS0ChainScheme, numSplitKeys)
	for i := 0; i < numSplitKeys; i++ {
		sigAggScheme[i].SetPrivateKey(w.Keys[i].PrivateKey)
	}
	var aggrSig string
	for i := 1; i < numSplitKeys; i++ {
		tmpSig, _ := sigAggScheme[i].Sign(hash)
		aggrSig, _ = sigAggScheme[0].Add(tmpSig, hash)
	}
	if aggrSig != signature {
		t.Fatalf("split key signature failed")
	}
}
