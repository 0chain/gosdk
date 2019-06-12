package zcncrypto

import (
	"fmt"
	"testing"
)

var verifyPublickey = `e8a6cfa7b3076ae7e04764ffdfe341632a136b52953dfafa6926361dd9a466196faecca6f696774bbd64b938ff765dbc837e8766a5e2d8996745b2b94e1beb9e`
var signPrivatekey = `5e1fc9c03d53a8b9a63030acc2864f0c33dffddb3c276bf2b3c8d739269cc018`
var data = `TEST`
var expectedHash = `f4f08e9367e133dc42a4b9c9c665a9efbd4bf15db14d49c6ec51d0dc4c437ffb`

func TestSignatureScheme(t *testing.T) {
	sigScheme := NewSignatureScheme("bls0chain")
	switch sigScheme.(type) {
	case SignatureScheme:
		// pass
	default:
		t.Fatalf("Signature scheme invalid")
	}
	err := sigScheme.GenerateKeys()
	if err != nil {
		t.Fatalf("Generate Key failed %s", err.Error())
	}
	pk, err := sigScheme.GetPublicKeyWithIdx(0)
	if err != nil || pk == "" {
		t.Fatalf("Get public key0 failed")
	}
	pk, err = sigScheme.GetPublicKeyWithIdx(1)
	if err != nil || pk == "" {
		t.Fatalf("Get public key1 failed")
	}
	_, err = sigScheme.GetPublicKeyWithIdx(2)
	if err == nil {
		t.Fatalf("Get public key2 failed")
	}
	pk, err = sigScheme.GetPublicKey()
	if err != nil || pk == "" {
		t.Fatalf("Get public key1 failed")
	}
	fmt.Printf("Aggr publickey:%s\n", pk)
	sk, err := sigScheme.GetSecretKeyWithIdx(0)
	if err != nil || sk == "" {
		t.Fatalf("Get secret key0 failed")

	}
	sk, err = sigScheme.GetSecretKeyWithIdx(1)
	if err != nil || sk == "" {
		t.Fatalf("Get secret key1 failed")

	}
	_, err = sigScheme.GetPublicKeyWithIdx(2)
	if err == nil {
		t.Fatalf("Get secret key2 failed")
	}
	mk, err := sigScheme.GetMnemonic()
	if err != nil || mk == "" {
		t.Fatalf("Get Mnemonic failed")
	}
}

func TestSSSignAndVerify(t *testing.T) {
	signScheme := NewSignatureScheme("bls0chain")
	signScheme.SetPrivateKey(signPrivatekey)
	signature, err := signScheme.Sign(data)
	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	verifyScheme := NewSignatureScheme("bls0chain")
	verifyScheme.SetPublicKey(verifyPublickey)
	if ok, err := verifyScheme.Verify(signature, data); err != nil || !ok {
		t.Fatalf("Verification failed\n")
	}
}

func BenchmarkBLSSign(b *testing.B) {
	sigScheme := NewSignatureScheme("bls0chain")
	sigScheme.SetPrivateKey(signPrivatekey)
	for i := 0; i < b.N; i++ {
		_, err := sigScheme.Sign(data)
		if err != nil {
			b.Fatalf("BLS signing failed")
		}
	}
}

func TestRecoveryKeys(t *testing.T) {
	mnemonic := `depend disease that come until quality benefit shrimp garbage normal curious artist author have april awkward weather hamster credit room gym health interest cupboard`
	sk0 := `2f7b126a03c0be1d69e2ec9d6688ea68ff420498c55bcf11805ed46dc6a2be00`
	sk1 := `fd5eaef81085acb13474e018d4ab92d3d369ceddbf7b30f81ed7172c6b748017`
	pk0 := `d0f529507c4270353e78593d253f865a37e7a74a90729c4dbcf43bd1e0947a0c7e6f3bea9355cbc53e19d5095bcf5af016e814fcea8cf95af189d54445f103a4`
	pk1 := `2c5bacdc83d737fa3d6617e31a1e89979544a66e3cf0e03bca71435ec1b8e2207928f1b6683a930483d288af61e39f53a6cbc21e598045eb6847bdc6c84e711c`
	sigScheme := NewSignatureScheme("bls0chain")
	err := sigScheme.RecoverKeys(mnemonic)
	if err != nil {
		t.Fatalf("set Recover Keys failed")
	}
	pkey, err := sigScheme.GetPublicKeyWithIdx(0)
	pkey2, err := sigScheme.GetPublicKeyWithIdx(1)
	if err != nil || pkey != pk0 || pkey2 != pk1 {
		t.Fatalf("Generate public key from mnemonic failed")
	}
	skey, err := sigScheme.GetSecretKeyWithIdx(0)
	skey2, err := sigScheme.GetSecretKeyWithIdx(1)
	if err != nil || skey != sk0 || skey2 != sk1 {
		t.Fatalf("Generate secret key from mnemonic failed")
	}

}

func TestCombinedSignAndVerify(t *testing.T) {
	// mnemonic := `grunt glad happy source inherit sing merge shop lesson oyster frost indoor symptom laugh output rail average mean pill dose buzz rhythm hill adult`
	sk0 := `c36f2f92b673cf057a32e8bd0ca88888e7ace40337b737e9c7459fdc4c521918`
	sk1 := `704b6f489583bf1118432fcfb38e63fc2d4b61e524fb196cbd95413f8eb91c12`
	// pk0 := `574ad8275d17b5f8f1b1557214342c8607cd5a267f3a5879133c694e2475a9129765a96e1564cef1c11328883b36222d57ecc8113d4d6d71096684d13de12d91`
	// pk1 := `8be97d7c4b59717f1d74861f6ee89a9ef0e6e45af825bf405a364d3ab58075085212063690d0d61c49f8e956f1e61b1a80d62508d13d9b91f0c889f9dae35a94`
	primaryKey := `f72fd53ee85e84157d3106053754594f697e0bfca1f73f91a41f7bb0797d901acefd80fcc2da98aae690af0ee9c795d6590c1808f26490306433b4e9c42f7b1f`

	// Create signatue for 1
	sig0 := NewSignatureScheme("bls0chain")
	err := sig0.SetPrivateKey(sk0)
	if err != nil {
		t.Fatalf("Set private key failed - %s", err.Error())
	}
	signature, err := sig0.Sign(data)
	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	// Create signature for second
	sig1 := NewSignatureScheme("bls0chain")
	err = sig1.SetPrivateKey(sk1)
	if err != nil {
		t.Fatalf("Set private key failed - %s", err.Error())
	}
	addSig, err := sig1.Add(signature, data)

	verifyScheme := NewSignatureScheme("bls0chain")
	err = verifyScheme.SetPublicKey(primaryKey)
	if err != nil {
		t.Fatalf("Set public key failed")
	}
	if ok, err := verifyScheme.Verify(addSig, data); err != nil || !ok {
		t.Fatalf("Verification failed\n")
	}
}
