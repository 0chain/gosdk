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

func TestReEncryptionAndDecryptionForShareData(t *testing.T) {
	client_mnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"
	client_encscheme := NewEncryptionScheme()
	client_encscheme.Initialize(client_mnemonic)
	client_encscheme.InitForEncryption("filetype:audio")
	client_enc_pub_key, err := client_encscheme.GetPublicKey()
	assert.Nil(t, err)

	shared_client_mnemonic := "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"
	shared_client_encscheme := NewEncryptionScheme()
	shared_client_encscheme.Initialize(shared_client_mnemonic)
	shared_client_encscheme.InitForEncryption("filetype:audio")

	enc_msg, err := shared_client_encscheme.Encrypt([]byte("encrypted_data_uttam"))
	assert.Nil(t, err)
	regenkey, err := shared_client_encscheme.GetReGenKey(client_enc_pub_key, "filetype:audio")
	assert.Nil(t, err)
	enc_msg.ReEncryptionKey = regenkey

	client_decryption_scheme := NewEncryptionScheme()
	client_decryption_scheme.Initialize(client_mnemonic)
	client_decryption_scheme.InitForDecryption("filetype:audio", enc_msg.EncryptedKey)

	result, err := client_decryption_scheme.Decrypt(enc_msg)
	assert.Nil(t, err)
	assert.Equal(t, string(result), "encrypted_data_uttam")
}

func TestReEncryptionAndDecryptionForMarketplaceShare(t *testing.T) {
	client_mnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"
	client_encscheme := NewEncryptionScheme()
	client_encscheme.Initialize(client_mnemonic)
	client_encscheme.InitForEncryption("filetype:audio")
	client_enc_pub_key, err := client_encscheme.GetPublicKey()
	assert.Nil(t, err)

	// seller uploads and blobber encrypts the data
	blobber_mnemonic := "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"
	blobber_encscheme := NewEncryptionScheme()
	blobber_encscheme.Initialize(blobber_mnemonic)
	blobber_encscheme.InitForEncryption("filetype:audio")
	enc_msg, err := blobber_encscheme.Encrypt([]byte("encrypted_data_uttam"))
	assert.Nil(t, err)

	// buyer requests data from blobber, blobber reencrypts the data with regen key using buyer public key
	blobber_encscheme = NewEncryptionScheme()
	blobber_encscheme.Initialize(blobber_mnemonic)
	blobber_encscheme.InitForDecryption("filetype:audio", enc_msg.EncryptedKey)
	regenkey, err := blobber_encscheme.GetReGenKey(client_enc_pub_key, "filetype:audio")
	assert.Nil(t, err)
	reenc_msg, err := blobber_encscheme.ReEncrypt(enc_msg, regenkey, client_enc_pub_key)
	assert.Nil(t, err)

	client_decryption_scheme := NewEncryptionScheme()
	client_decryption_scheme.Initialize(client_mnemonic)
	client_decryption_scheme.InitForDecryption("filetype:audio", enc_msg.EncryptedKey)

	result, err := client_decryption_scheme.ReDecrypt(reenc_msg)
	assert.Nil(t, err)
	assert.Equal(t, string(result), "encrypted_data_uttam")
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