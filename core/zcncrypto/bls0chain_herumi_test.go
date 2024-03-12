package zcncrypto

import (
	"fmt"
	"testing"

	"github.com/0chain/errors"
	"github.com/stretchr/testify/require"

	"github.com/0chain/gosdk/core/encryption"
)

var verifyPublickey = `e8a6cfa7b3076ae7e04764ffdfe341632a136b52953dfafa6926361dd9a466196faecca6f696774bbd64b938ff765dbc837e8766a5e2d8996745b2b94e1beb9e`
var signPrivatekey = `5e1fc9c03d53a8b9a63030acc2864f0c33dffddb3c276bf2b3c8d739269cc018`
var data = `TEST`
var blsWallet *Wallet

func TestSignatureScheme(t *testing.T) {
	sigScheme := &HerumiScheme{}

	w, err := sigScheme.GenerateKeys()
	if err != nil {
		t.Fatalf("Generate Key failed %s", errors.Top(err))
	}
	if w.ClientID == "" || w.ClientKey == "" || len(w.Keys) != 1 || w.Mnemonic == "" {
		t.Fatalf("Invalid keys generated")
	}
	blsWallet = w

}

func TestSSSignAndVerify(t *testing.T) {
	signScheme := NewSignatureScheme("bls0chain")
	err := signScheme.SetPrivateKey(signPrivatekey)

	require.NoError(t, err)

	hash := Sha3Sum256(data)
	signature, err := signScheme.Sign(hash)
	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	verifyScheme := NewSignatureScheme("bls0chain")
	err = verifyScheme.SetPublicKey(verifyPublickey)
	require.NoError(t, err)
	if ok, err := verifyScheme.Verify(signature, hash); err != nil || !ok {
		t.Fatalf("Verification failed\n")
	}
}

func TestSSA(t *testing.T) {
	signScheme := NewSignatureScheme("bls0chain")
	err := signScheme.SetPrivateKey("f482aa19d3a3f6cebcd4f8a99de292bcf4bf07e937be1350634086f4aa02e704")

	require.NoError(t, err)

	hash := Sha3Sum256("hello")
	signature, err := signScheme.Sign(hash)
	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	fmt.Println(signature)
	// verifyScheme := NewSignatureScheme("bls0chain")
	// err = verifyScheme.SetPublicKey(verifyPublickey)
	// require.NoError(t, err)
	// if ok, err := verifyScheme.Verify(signature, hash); err != nil || !ok {
	// 	t.Fatalf("Verification failed\n")
	// }
}

func TestVerify(t *testing.T) {
	sk := "a931522f9949ff26b22db98b26e59cc92258457965f13ce6113cc2b5d2165513"
	hash := "eb82aa875b3298ae7e625d8d8f13475004a4942bd8fcd7285e8ab9ad20651872"
	sm := NewSignatureScheme("bls0chain")
	if err2 := sm.SetPrivateKey(sk); err2 != nil {
		t.Error(err2)
	}
	sig, _ := sm.Sign(hash)
	fmt.Println("now sig:", sig)

	pk := "47e94b6c5399f8c0005c6f3202dec43e37d171b0eff24d75cdcf14861f088106cf88df15b3335dcd0806365db4d1b3e70579a8bd82eb665c881ef2273d6bdd03"
	verifyScheme := NewSignatureScheme("bls0chain")
	if err := verifyScheme.SetPublicKey(pk); err != nil {
		t.Error(err)
	}
	ok, err := verifyScheme.Verify(sig, hash)
	require.NoError(t, err)
	fmt.Println("verify result:", ok)
}

func BenchmarkBLSSign(b *testing.B) {
	sigScheme := NewSignatureScheme("bls0chain")
	err := sigScheme.SetPrivateKey(signPrivatekey)
	require.NoError(b, err)
	for i := 0; i < b.N; i++ {
		_, err := sigScheme.Sign(encryption.Hash(data))
		if err != nil {
			b.Fatalf("BLS signing failed")
		}
	}
}

func TestRecoveryKeys(t *testing.T) {

	sigScheme := &HerumiScheme{}

	w, err := sigScheme.RecoverKeys(testMnemonic)
	if err != nil {
		t.Fatalf("set Recover Keys failed")
	}

	require.Equal(t, testHerumiPrivateKey, w.Keys[0].PrivateKey, "Recover key didn't match with private key")
	require.Equal(t, testHerumiPublicKey, w.Keys[0].PublicKey, "Recover key didn't match with public key")
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
		t.Fatalf("Set private key failed - %s", errors.Top(err))
	}
	signature, err := sig0.Sign(hash)
	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	// Create signature for second
	sig1 := NewSignatureScheme("bls0chain")
	err = sig1.SetPrivateKey(sk1)
	if err != nil {
		t.Fatalf("Set private key failed - %s", errors.Top(err))
	}
	addSig, err := sig1.Add(signature, hash)

	require.NoError(t, err)

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
	primaryKeyStr := `872eac6370c72093535fa395ad41a08ee90c9d0d46df9461eb2515451f389d1b`
	// primaryKeyStr := `c36f2f92b673cf057a32e8bd0ca88888e7ace40337b737e9c7459fdc4c521918`
	sig0 := NewSignatureScheme("bls0chain")
	err := sig0.SetPrivateKey(primaryKeyStr)
	if err != nil {
		t.Fatalf("Set private key failed - %s", errors.Top(err))
	}
	data = "823bb3dc0b80a6c86922a884e63908cb9e963ef488688b41e32cbf4d84471a1f"
	hash := Sha3Sum256(data)
	signature, err := sig0.Sign(hash)
	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	numSplitKeys := int(2)
	w, err := sig0.SplitKeys(numSplitKeys)
	if err != nil {
		t.Fatalf("Splitkeys key failed - %s", errors.Top(err))
	}
	sigAggScheme := make([]SignatureScheme, numSplitKeys)
	for i := 0; i < numSplitKeys; i++ {
		sigAggScheme[i] = NewSignatureScheme("bls0chain")
		err = sigAggScheme[i].SetPrivateKey(w.Keys[i].PrivateKey)
		fmt.Println("seckey:", sigAggScheme[i].GetPrivateKey())

		require.NoError(t, err)
	}
	var aggrSig string
	for i := 1; i < numSplitKeys; i++ {
		tmpSig, _ := sigAggScheme[i].Sign(hash)
		fmt.Println("tmpSig:", tmpSig)
		aggrSig, _ = sigAggScheme[0].Add(tmpSig, hash)
	}
	if aggrSig != signature {
		t.Fatalf("split key signature failed")
	}
	fmt.Println("aggrSig:", aggrSig)
}
