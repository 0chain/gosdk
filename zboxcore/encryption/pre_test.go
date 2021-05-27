package encryption

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"testing"
)

func TestMnemonic(t *testing.T) {
	mnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"

	encscheme := NewEncryptionScheme()

	err := encscheme.Initialize(mnemonic)
	require.NoError(t, err)


	encscheme.InitForEncryption("filetype:audio")
	pvk, _ := encscheme.GetPrivateKey()
	expectedPvk := "XsQLPaRBOFS+3KfXq2/uyAPE+/qq3VW0OkW0T9q93wQ="
	require.Equal(t, expectedPvk, pvk)
	pubk, _ := encscheme.GetPublicKey()
	expectedPubk := "PwpVIXgXbnt8NJmy+R4aSwG8HwJbsbT2JVQqa0bayZQ="
	require.Equal(t, expectedPubk, pubk)

}

func TestKyberPointMarshal(t *testing.T) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	fmt.Println(suite.Point())
}