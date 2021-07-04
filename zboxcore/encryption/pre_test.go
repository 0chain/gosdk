package encryption

import (
	crand "crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"math/rand"
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
	require.Nil(t, err)

	shared_client_mnemonic := "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"
	shared_client_encscheme := NewEncryptionScheme()
	shared_client_encscheme.Initialize(shared_client_mnemonic)
	shared_client_encscheme.InitForEncryption("filetype:audio")

	enc_msg, err := shared_client_encscheme.Encrypt([]byte("encrypted_data_uttam"))
	require.Nil(t, err)
	regenkey, err := shared_client_encscheme.GetReGenKey(client_enc_pub_key, "filetype:audio")
	t.Log("regen", regenkey)
	require.Nil(t, err)
	enc_msg.ReEncryptionKey = regenkey

	client_decryption_scheme := NewEncryptionScheme()
	client_decryption_scheme.Initialize(client_mnemonic)
	client_decryption_scheme.InitForDecryption("filetype:audio", enc_msg.EncryptedKey)

	t.Log("decrypting")
	result, err := client_decryption_scheme.Decrypt(enc_msg)
	require.Nil(t, err)
	require.Equal(t, string(result), "encrypted_data_uttam")
}

func TestReEncryptionAndDecryptionForMarketplaceShare(t *testing.T) {
	client_mnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"
	client_encscheme := NewEncryptionScheme()
	client_encscheme.Initialize(client_mnemonic)
	client_encscheme.InitForEncryption("filetype:audio")
	client_enc_pub_key, err := client_encscheme.GetPublicKey()
	require.Nil(t, err)

	// seller uploads and blobber encrypts the data
	blobber_mnemonic := "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"
	blobber_encscheme := NewEncryptionScheme()
	blobber_encscheme.Initialize(blobber_mnemonic)
	blobber_encscheme.InitForEncryption("filetype:audio")
	data_to_encrypt := "encrypted_data_uttaencrypted_data_uttaencrypted_data_uttaencrypted_data_uttaencrypted_data_uttaencrypted_data_uttaencrypted_data_uttaencrypted_data_uttaencrypted_data_uttammmmmmmmmencrypted_data_uttam"
	enc_msg, err := blobber_encscheme.Encrypt([]byte(data_to_encrypt))
	require.Nil(t, err)

	// buyer requests data from blobber, blobber reencrypts the data with regen key using buyer public key
	blobber_encscheme = NewEncryptionScheme()
	blobber_encscheme.Initialize(blobber_mnemonic)
	blobber_encscheme.InitForDecryption("filetype:audio", enc_msg.EncryptedKey)
	regenkey, err := blobber_encscheme.GetReGenKey(client_enc_pub_key, "filetype:audio")
	require.Nil(t, err)
	reenc_msg, err := blobber_encscheme.ReEncrypt(enc_msg, regenkey, client_enc_pub_key)
	require.Nil(t, err)
	// verify encrypted message size
	d1, _ := reenc_msg.D1.MarshalBinary()
	d4, _ := reenc_msg.D4.MarshalBinary()
	d5, _ := reenc_msg.D5.MarshalBinary()
	require.Equal(t, 44, len(base64.StdEncoding.EncodeToString(d1)))
	require.Equal(t, 88, len(base64.StdEncoding.EncodeToString(reenc_msg.D3)))
	require.Equal(t, 44, len(base64.StdEncoding.EncodeToString(d4)))
	require.Equal(t, 44, len(base64.StdEncoding.EncodeToString(d5)))

	client_decryption_scheme := NewEncryptionScheme()
	client_decryption_scheme.Initialize(client_mnemonic)
	client_decryption_scheme.InitForDecryption("filetype:audio", enc_msg.EncryptedKey)

	result, err := client_decryption_scheme.ReDecrypt(reenc_msg)
	require.Nil(t, err)
	require.Equal(t, string(result), data_to_encrypt)
}

func TestKyberPointMarshal(t *testing.T) {
	reenc := ReEncryptedMessage {
		D1: curve.NewEdwardsPoint(),
		D2: []byte("d2"),
		D3: []byte("d3"),
		D4: curve.NewEdwardsPoint(),
		D5: curve.NewEdwardsPoint(),
	}
	marshalled, err := reenc.Marshal()
	require.Nil(t, err)
	newmsg := &ReEncryptedMessage{
		D1: curve.NewEdwardsPoint(),
		D4: curve.NewEdwardsPoint(),
		D5: curve.NewEdwardsPoint(),
	}
	err = newmsg.Unmarshal(marshalled)
	require.Equal(t, newmsg.D2, reenc.D2)
	require.Equal(t, newmsg.D3, reenc.D3)
	require.Equal(t, newmsg.D1, reenc.D1)
	require.Equal(t, newmsg.D4, reenc.D4)
	require.Equal(t, newmsg.D5, reenc.D5)
}

func BenchmarkMarshal(t *testing.B) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	for i := 0; i < 1000; i++ {
		point := suite.Point().Pick(suite.RandomStream())
		data, err := point.MarshalBinary()
		require.Nil(t, err)
		require.Equal(t, 44, len(base64.StdEncoding.EncodeToString(data)))
	}
}

func BenchmarkEncrypt(t *testing.B) {
	mnemonic := "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"
	encscheme := NewEncryptionScheme()
	encscheme.Initialize(mnemonic)
	encscheme.InitForEncryption("filetype:audio")
	for i := 0; i < 10000; i++ {
		dataToEncrypt := make([]byte, fileref.CHUNK_SIZE)
		rand.Read(dataToEncrypt)
		_, err := encscheme.Encrypt(dataToEncrypt)
		require.Nil(t, err)
		require.Equal(t, len(dataToEncrypt), fileref.CHUNK_SIZE)
	}
}

func BenchmarkReEncryptAndReDecrypt(t *testing.B) {
	client_mnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"
	client_encscheme := NewEncryptionScheme()
	client_encscheme.Initialize(client_mnemonic)
	client_encscheme.InitForEncryption("filetype:audio")
	client_enc_pub_key, err := client_encscheme.GetPublicKey()
	require.Nil(t, err)

	// seller uploads and blobber encrypts the data
	blobber_mnemonic := "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"
	blobber_encscheme := NewEncryptionScheme()
	blobber_encscheme.Initialize(blobber_mnemonic)
	blobber_encscheme.InitForEncryption("filetype:audio")
	// buyer requests data from blobber, blobber reencrypts the data with regen key using buyer public key
	regenkey, err := blobber_encscheme.GetReGenKey(client_enc_pub_key, "filetype:audio")
	for i := 0; i < 10000; i++ {
		dataToEncrypt := make([]byte, fileref.CHUNK_SIZE)
		rand.Read(dataToEncrypt)
		enc_msg, err := blobber_encscheme.Encrypt(dataToEncrypt)
		require.Nil(t, err)
		reenc_msg, err := blobber_encscheme.ReEncrypt(enc_msg, regenkey, client_enc_pub_key)
		require.Nil(t, err)

		client_decryption_scheme := NewEncryptionScheme()
		client_decryption_scheme.Initialize(client_mnemonic)
		client_decryption_scheme.InitForDecryption("filetype:audio", enc_msg.EncryptedKey)

		_, err = client_decryption_scheme.ReDecrypt(reenc_msg)
		require.Nil(t, err)
	}
}

func TestNic(t *testing.T) {
	s, err := scalar.New().SetRandom(crand.Reader)
	marsh, err := s.MarshalBinary()
	t.Log(err)
	t.Log(base64.StdEncoding.EncodeToString(marsh))
	t.Log(len(marsh))


	suite := edwards25519.NewBlakeSHA256Ed25519()
	mnemonic := "expose culture dignity plastic digital couple promote best pool error brush upgrade correct art become lobster nature moment obtain trial multiply arch miss toe"
	rand := suite.XOF([]byte(mnemonic))

	d := curve.EdwardsPoint{}

	// Create a public/private keypair (X,x)
	key := suite.Scalar().Pick(rand)
	t.Log(d, key)
	e := scalar.New().UnmarshalBinary([]byte("e8yBVeFVqnMrXJ4wyP3S/NcPLumwaszdmUOjmfoTz3A="))
	fmt.Println(e)
}