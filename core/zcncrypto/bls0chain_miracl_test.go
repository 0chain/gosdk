package zcncrypto

import (
	"testing"

	"github.com/0chain/errors"
	"github.com/stretchr/testify/require"

	"github.com/0chain/gosdk/core/encryption"
)

func TestMiraclSignatureScheme(t *testing.T) {
	sigScheme := &MiraclScheme{}

	w, err := sigScheme.GenerateKeys()
	if err != nil {
		t.Fatalf("Generate Key failed %s", errors.Top(err))
	}
	if w.ClientID == "" || w.ClientKey == "" || len(w.Keys) != 1 || w.Mnemonic == "" {
		t.Fatalf("Invalid keys generated")
	}
	blsWallet = w
}

func TestMiraclSSSignAndVerify(t *testing.T) {
	signScheme := &MiraclScheme{}
	err := signScheme.SetPrivateKey(signPrivatekey)

	require.NoError(t, err)

	hash := Sha3Sum256(data)
	signature, err := signScheme.Sign(hash)
	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	verifyScheme := &MiraclScheme{}
	err = verifyScheme.SetPublicKey(verifyPublickey)
	require.NoError(t, err)
	if ok, err := verifyScheme.Verify(signature, hash); err != nil || !ok {
		t.Fatalf("Verification failed\n")
	}
}

func BenchmarkMircalBLSSign(b *testing.B) {
	sigScheme := &MiraclScheme{}
	err := sigScheme.SetPrivateKey(signPrivatekey)
	require.NoError(b, err)
	for i := 0; i < b.N; i++ {
		_, err := sigScheme.Sign(encryption.Hash(data))
		if err != nil {
			b.Fatalf("BLS signing failed")
		}
	}
}

func TestMiraclRecoveryKeys(t *testing.T) {

	sigScheme := &MiraclScheme{}

	w, err := sigScheme.RecoverKeys(blsWallet.Mnemonic)
	if err != nil {
		t.Fatalf("set Recover Keys failed")
	}
	if w.ClientID != blsWallet.ClientID || w.ClientKey != blsWallet.ClientKey {
		t.Fatalf("Recover key didn't match with generated keys")
	}

}

func TestMiraclCombinedSignAndVerify(t *testing.T) {
	sk0 := `c36f2f92b673cf057a32e8bd0ca88888e7ace40337b737e9c7459fdc4c521918`
	sk1 := `704b6f489583bf1118432fcfb38e63fc2d4b61e524fb196cbd95413f8eb91c12`
	primaryKey := `f72fd53ee85e84157d3106053754594f697e0bfca1f73f91a41f7bb0797d901acefd80fcc2da98aae690af0ee9c795d6590c1808f26490306433b4e9c42f7b1f`

	hash := Sha3Sum256(data)
	// Create signatue for 1
	sig0 := &MiraclScheme{}
	err := sig0.SetPrivateKey(sk0)
	if err != nil {
		t.Fatalf("Set private key failed - %s", errors.Top(err))
	}
	signature, err := sig0.Sign(hash)
	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	// Create signature for second
	sig1 := &MiraclScheme{}
	err = sig1.SetPrivateKey(sk1)
	if err != nil {
		t.Fatalf("Set private key failed - %s", errors.Top(err))
	}
	addSig, err := sig1.Add(signature, hash)

	require.NoError(t, err)

	verifyScheme := &MiraclScheme{}
	err = verifyScheme.SetPublicKey(primaryKey)
	if err != nil {
		t.Fatalf("Set public key failed")
	}
	if ok, err := verifyScheme.Verify(addSig, hash); err != nil || !ok {
		t.Fatalf("Verification failed\n")
	}
}

func TestMiraclSplitKey(t *testing.T) {
	primaryKeyStr := `c36f2f92b673cf057a32e8bd0ca88888e7ace40337b737e9c7459fdc4c521918`
	sig0 := &MiraclScheme{}
	err := sig0.SetPrivateKey(primaryKeyStr)
	if err != nil {
		t.Fatalf("Set private key failed - %s", errors.Top(err))
	}
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
		sigAggScheme[i] = &MiraclScheme{}
		err = sigAggScheme[i].SetPrivateKey(w.Keys[i].PrivateKey)

		require.NoError(t, err)
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
