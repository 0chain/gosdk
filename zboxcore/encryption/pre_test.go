package encryption

import (
	"github.com/stretchr/testify/assert"
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
	reenc := ReEncryptedMessage {
		D1: suite.Point(),
		D2: []byte("d2"),
		D3: []byte("d3"),
		D4: suite.Point(),
		D5: suite.Point(),
	}
	marshalled, err := reenc.MarshalJSON()
	assert.Nil(t, err)
	expected := "{\"d1Bytes\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"d2Bytes\":\"ZDI=\",\"d3Bytes\":\"ZDM=\",\"d4Bytes\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"d5Bytes\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"}"
	assert.Equal(t, expected, string(marshalled))
	newmsg := &ReEncryptedMessage{
		D1: suite.Point(),
		D4: suite.Point(),
		D5: suite.Point(),
	}
	err = newmsg.UnmarshalJSON(marshalled)
	assert.Equal(t, newmsg.D2, reenc.D2)
	assert.Equal(t, newmsg.D3, reenc.D3)
	assert.Equal(t, newmsg.D1.String(), reenc.D1.String())
	assert.Equal(t, newmsg.D4.String(), reenc.D4.String())
	assert.Equal(t, newmsg.D5.String(), reenc.D5.String())
}