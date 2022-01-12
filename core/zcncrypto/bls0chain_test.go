//go:build !js && !wasm
// +build !js,!wasm

package zcncrypto

import (
	"testing"

	"github.com/herumi/bls-go-binary/bls"
	"github.com/stretchr/testify/require"
)

const (
	testMnemonic = "silent tape impulse glimpse state craft sheriff embody bonus clay confirm column swift kingdom door stove mad switch chalk theory pause canoe insane struggle"

	testHerumiPublicKey    = "fd2f78b5988719434d6a0782231962934fe1a6f805f98e1bff2c90399a765500ffff9a1cc8c5826feea66d738a7e74ffba7f7dd23e499b5817d8a88e68185f95"
	testHerumiPublicKeyStr = "1 55769a39902cff1b8ef905f8a6e14f9362192382076a4d43198798b5782ffd 155f18688ea8d817589b493ed27d7fbaff747e8a736da6ee6f82c5c81c9affff e72525d5dda83d7b169653d3a78bd6d6e36cee1f9974d8f30cbfac33a18efb9 19c1c219dbd76990330f778f18d472f10494a6811bb46e36d21bfdf273c03220"
	testHerumiPrivateKey   = "baa512aee00f5ff9eafcd82a16fa81d450b2a1a1e35f638cb7e4c2caf01bc407"

	testMiraclPublicKeyStr = "55769a39902cff1b8ef905f8a6e14f9362192382076a4d43198798b5782ffd155f18688ea8d817589b493ed27d7fbaff747e8a736da6ee6f82c5c81c9affffe72525d5dda83d7b169653d3a78bd6d6e36cee1f9974d8f30cbfac33a18efb919c1c219dbd76990330f778f18d472f10494a6811bb46e36d21bfdf273c03220"
)

func TestGenerateKeys(t *testing.T) {
	herumi := &HerumiScheme{}

	w1, err := herumi.RecoverKeys(testMnemonic)

	require.NoError(t, err)

	require.Equal(t, testHerumiPublicKey, w1.Keys[0].PublicKey)

	var pk1 bls.PublicKey
	err = pk1.DeserializeHexStr(w1.Keys[0].PublicKey)
	require.NoError(t, err)

	require.Equal(t, testHerumiPublicKeyStr, pk1.GetHexString())

	require.NoError(t, err)

}

func TestSignAndVerify(t *testing.T) {
	signScheme := &HerumiScheme{}

	w, err := signScheme.RecoverKeys(testMnemonic)

	var pk = w.Keys[0].PublicKey
	var pk1 bls.PublicKey

	pk1.DeserializeHexStr(pk)

	require.NoError(t, err)

	hash := Sha3Sum256(data)
	signature, err := signScheme.Sign(hash)
	if err != nil {
		t.Fatalf("BLS signing failed")
	}
	verifyScheme := &HerumiScheme{}
	err = verifyScheme.SetPublicKey(w.Keys[0].PublicKey)
	require.NoError(t, err)
	if ok, err := verifyScheme.Verify(signature, hash); err != nil || !ok {
		t.Fatalf("Verification failed\n")
	}
}
