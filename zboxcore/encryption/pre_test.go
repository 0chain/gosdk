package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/group/edwards25519"
)

func TestMnemonic(t *testing.T) {
	mnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"

	encscheme := NewEncryptionScheme()

	_, err := encscheme.Initialize(mnemonic)
	require.NoError(t, err)

	encscheme.InitForEncryption("filetype:audio")
	pvk, _ := encscheme.GetPrivateKey()
	expectedPvk := "XsQLPaRBOFS+3KfXq2/uyAPE+/qq3VW0OkW0T9q93wQ="
	require.Equal(t, expectedPvk, pvk)
	pubk, _ := encscheme.GetPublicKey()
	expectedPubk := "PwpVIXgXbnt8NJmy+R4aSwG8HwJbsbT2JVQqa0bayZQ="
	require.Equal(t, expectedPubk, pubk)

}

func TestEncryptDecrypt(t *testing.T) {
	mnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"
	dataToEncrypt := "encrypted_data_uttam"

	encscheme := NewEncryptionScheme()
	_, err := encscheme.Initialize(mnemonic)
	require.NoError(t, err)
	encscheme.InitForEncryption("filetype:audio")

	encMessage, err := encscheme.Encrypt([]byte(dataToEncrypt))
	require.Nil(t, err)

	decrypted, err := encscheme.Decrypt(encMessage)
	require.Nil(t, err)

	require.Equal(t, string(decrypted), dataToEncrypt)
}

func TestReEncryptionAndDecryptionForShareData(t *testing.T) {
	client_mnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"
	client_encscheme := NewEncryptionScheme()
	_, err := client_encscheme.Initialize(client_mnemonic)
	require.Nil(t, err)
	client_encscheme.InitForEncryption("filetype:audio")
	client_enc_pub_key, err := client_encscheme.GetPublicKey()
	require.Nil(t, err)

	shared_client_mnemonic := "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"
	shared_client_encscheme := NewEncryptionScheme()
	_, err = shared_client_encscheme.Initialize(shared_client_mnemonic)
	require.Nil(t, err)
	shared_client_encscheme.InitForEncryption("filetype:audio")

	enc_msg, err := shared_client_encscheme.Encrypt([]byte("encrypted_data_uttam"))
	require.Nil(t, err)
	regenkey, err := shared_client_encscheme.GetReGenKey(client_enc_pub_key, "filetype:audio")
	require.Nil(t, err)
	enc_msg.ReEncryptionKey = regenkey

	client_decryption_scheme := NewEncryptionScheme()
	_, err = client_decryption_scheme.Initialize(client_mnemonic)
	require.Nil(t, err)
	err = client_decryption_scheme.InitForDecryption("filetype:audio", enc_msg.EncryptedKey)
	require.Nil(t, err)

	result, err := client_decryption_scheme.Decrypt(enc_msg)
	require.Nil(t, err)
	require.Equal(t, string(result), "encrypted_data_uttam")
}

func TestReEncryptionAndDecryptionForMarketplaceShare(t *testing.T) {
	client_mnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"
	client_encscheme := NewEncryptionScheme()
	_, err := client_encscheme.Initialize(client_mnemonic)
	require.Nil(t, err)
	client_encscheme.InitForEncryption("filetype:audio")
	client_enc_pub_key, err := client_encscheme.GetPublicKey()
	require.Nil(t, err)

	// seller uploads and blobber encrypts the data
	blobber_mnemonic := "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"
	blobber_encscheme := NewEncryptionScheme()
	_, err = blobber_encscheme.Initialize(blobber_mnemonic)
	require.Nil(t, err)
	blobber_encscheme.InitForEncryption("filetype:audio")
	data_to_encrypt := "encrypted_data_uttaencrypted_data_uttaencrypted_data_uttaencrypted_data_uttaencrypted_data_uttaencrypted_data_uttaencrypted_data_uttaencrypted_data_uttaencrypted_data_uttammmmmmmmmencrypted_data_uttam"
	enc_msg, err := blobber_encscheme.Encrypt([]byte(data_to_encrypt))
	require.Nil(t, err)

	// buyer requests data from blobber, blobber reencrypts the data with regen key using buyer public key
	blobber_encscheme = NewEncryptionScheme()
	_, err = blobber_encscheme.Initialize(blobber_mnemonic)
	require.Nil(t, err)
	err = blobber_encscheme.InitForDecryption("filetype:audio", enc_msg.EncryptedKey)
	require.Nil(t, err)
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
	_, err = client_decryption_scheme.Initialize(client_mnemonic)
	require.Nil(t, err)
	err = client_decryption_scheme.InitForDecryption("filetype:audio", enc_msg.EncryptedKey)
	require.Nil(t, err)

	result, err := client_decryption_scheme.ReDecrypt(reenc_msg)
	require.Nil(t, err)
	require.Equal(t, string(result), data_to_encrypt)
}

func TestKyberPointMarshal(t *testing.T) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	reenc := ReEncryptedMessage{
		D1: suite.Point(),
		D2: []byte("d2"),
		D3: []byte("d3"),
		D4: suite.Point(),
		D5: suite.Point(),
	}
	marshalled, err := reenc.Marshal()
	require.Nil(t, err)
	newmsg := &ReEncryptedMessage{
		D1: suite.Point(),
		D4: suite.Point(),
		D5: suite.Point(),
	}
	err = newmsg.Unmarshal(marshalled)
	require.Nil(t, err)
	require.Equal(t, newmsg.D2, reenc.D2)
	require.Equal(t, newmsg.D3, reenc.D3)
	require.Equal(t, newmsg.D1.String(), reenc.D1.String())
	require.Equal(t, newmsg.D4.String(), reenc.D4.String())
	require.Equal(t, newmsg.D5.String(), reenc.D5.String())
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

// helper: create an initialized encryption scheme from a mnemonic
func newScheme(t *testing.T, mnemonic string) EncryptionScheme {
	t.Helper()
	s := NewEncryptionScheme()
	_, err := s.Initialize(mnemonic)
	require.NoError(t, err)
	return s
}

// helper: simulate blobber re-encryption + recipient decryption
func blobberReEncryptAndDecrypt(t *testing.T, encMsg *EncryptedMessage, fileC1 string, reKey string, recipientPubKey string, recipientScheme EncryptionScheme) []byte {
	t.Helper()
	// Blobber side: init with the file's C1, re-encrypt
	blobber := NewEncryptionScheme()
	blobber.Initialize("") // blobber has no owner mnemonic
	err := blobber.InitForDecryption("filetype:audio", fileC1)
	require.NoError(t, err)
	reEnc, err := blobber.ReEncrypt(encMsg, reKey, recipientPubKey)
	require.NoError(t, err)
	// Recipient side: decrypt
	dec, err := recipientScheme.ReDecrypt(reEnc)
	require.NoError(t, err)
	return dec
}

// TestReKeyWorksAfterFileUpdate verifies that when a file is updated (new
// random T, new C1), the existing re-encryption key still works. The blobber
// re-encrypts with the new C1 and the recipient decrypts successfully.
func TestReKeyWorksAfterFileUpdate(t *testing.T) {
	ownerMnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"
	recipientMnemonic := "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"

	// --- Setup recipient ---
	recipient := newScheme(t, recipientMnemonic)
	recipient.InitForEncryption("filetype:audio")
	recipientPubKey, _ := recipient.GetPublicKey()

	// --- Owner uploads v1 ---
	ownerV1 := newScheme(t, ownerMnemonic)
	ownerV1.InitForEncryption("filetype:audio")
	c1V1 := ownerV1.GetEncryptedKey()

	v1Data := []byte("original file content v1")
	v1Enc, err := ownerV1.Encrypt(v1Data)
	require.NoError(t, err)

	// --- Owner shares file, generating re-encryption key ---
	reKey, err := ownerV1.GetReGenKey(recipientPubKey, "filetype:audio")
	require.NoError(t, err)

	// --- Recipient downloads v1 ---
	recipientDec := newScheme(t, recipientMnemonic)
	recipientDec.InitForEncryption("filetype:audio")
	dec := blobberReEncryptAndDecrypt(t, v1Enc, c1V1, reKey, recipientPubKey, recipientDec)
	require.Equal(t, string(v1Data), string(dec))

	// --- Owner updates file to v2 (new random T, new C1) ---
	ownerV2 := newScheme(t, ownerMnemonic)
	ownerV2.InitForEncryption("filetype:audio") // new random T
	c1V2 := ownerV2.GetEncryptedKey()
	require.NotEqual(t, c1V1, c1V2, "update produces new C1")

	v2Data := []byte("updated file content v2 — completely different")
	v2Enc, err := ownerV2.Encrypt(v2Data)
	require.NoError(t, err)

	// --- Recipient downloads v2 using the SAME re-encryption key ---
	recipientDec2 := newScheme(t, recipientMnemonic)
	recipientDec2.InitForEncryption("filetype:audio")
	dec2 := blobberReEncryptAndDecrypt(t, v2Enc, c1V2, reKey, recipientPubKey, recipientDec2)
	require.Equal(t, string(v2Data), string(dec2), "same reKey decrypts updated file with new C1")
}

// TestReKeyWorksForEncryptedFolder verifies that a single re-encryption key
// works for multiple files in a folder, each with a different C1.
func TestReKeyWorksForEncryptedFolder(t *testing.T) {
	ownerMnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"
	recipientMnemonic := "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"

	// --- Setup recipient ---
	recipient := newScheme(t, recipientMnemonic)
	recipient.InitForEncryption("filetype:audio")
	recipientPubKey, _ := recipient.GetPublicKey()

	// --- Owner generates ONE re-encryption key for the folder share ---
	ownerForShare := newScheme(t, ownerMnemonic)
	ownerForShare.InitForEncryption("filetype:audio")
	reKey, err := ownerForShare.GetReGenKey(recipientPubKey, "filetype:audio")
	require.NoError(t, err)

	// --- Simulate 3 files in the folder, each with unique C1 ---
	files := []struct {
		name string
		data []byte
	}{
		{"file1.txt", []byte("first file in folder")},
		{"subfolder/file2.txt", []byte("nested file in subfolder")},
		{"file3.png", []byte("binary image data for third file")},
	}

	for _, f := range files {
		// Each file upload gets its own random T → unique C1
		ownerFile := newScheme(t, ownerMnemonic)
		ownerFile.InitForEncryption("filetype:audio")
		fileC1 := ownerFile.GetEncryptedKey()

		encMsg, err := ownerFile.Encrypt(f.data)
		require.NoError(t, err)

		// Recipient decrypts using the folder's re-encryption key
		recipientDec := newScheme(t, recipientMnemonic)
		recipientDec.InitForEncryption("filetype:audio")
		dec := blobberReEncryptAndDecrypt(t, encMsg, fileC1, reKey, recipientPubKey, recipientDec)
		require.Equal(t, string(f.data), string(dec), "folder reKey works for %s", f.name)
	}
}

// TestReKeyWorksForNewFileAddedToSharedFolder verifies that when a new file
// is added to an already-shared encrypted folder, the existing re-encryption
// key (generated at share time) works for the new file.
func TestReKeyWorksForNewFileAddedToSharedFolder(t *testing.T) {
	ownerMnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"
	recipientMnemonic := "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"

	// --- Setup recipient ---
	recipient := newScheme(t, recipientMnemonic)
	recipient.InitForEncryption("filetype:audio")
	recipientPubKey, _ := recipient.GetPublicKey()

	// --- Owner shares folder (generates re-encryption key) ---
	ownerForShare := newScheme(t, ownerMnemonic)
	ownerForShare.InitForEncryption("filetype:audio")
	reKey, err := ownerForShare.GetReGenKey(recipientPubKey, "filetype:audio")
	require.NoError(t, err)

	// --- Original file in folder ---
	ownerFile1 := newScheme(t, ownerMnemonic)
	ownerFile1.InitForEncryption("filetype:audio")
	c1File1 := ownerFile1.GetEncryptedKey()

	file1Data := []byte("original file already in folder at share time")
	enc1, err := ownerFile1.Encrypt(file1Data)
	require.NoError(t, err)

	recipientDec := newScheme(t, recipientMnemonic)
	recipientDec.InitForEncryption("filetype:audio")
	dec1 := blobberReEncryptAndDecrypt(t, enc1, c1File1, reKey, recipientPubKey, recipientDec)
	require.Equal(t, string(file1Data), string(dec1))

	// --- Later: owner adds a NEW file to the folder ---
	ownerFile2 := newScheme(t, ownerMnemonic)
	ownerFile2.InitForEncryption("filetype:audio") // fresh random T
	c1File2 := ownerFile2.GetEncryptedKey()
	require.NotEqual(t, c1File1, c1File2, "new file gets different C1")

	file2Data := []byte("brand new file added after folder was already shared")
	enc2, err := ownerFile2.Encrypt(file2Data)
	require.NoError(t, err)

	// --- Recipient downloads the new file using the SAME folder reKey ---
	recipientDec2 := newScheme(t, recipientMnemonic)
	recipientDec2.InitForEncryption("filetype:audio")
	dec2 := blobberReEncryptAndDecrypt(t, enc2, c1File2, reKey, recipientPubKey, recipientDec2)
	require.Equal(t, string(file2Data), string(dec2),
		"folder reKey works for file added after sharing — no new auth ticket needed")
}

// TestReKeyWorksForNewSubfolderAddedToSharedFolder verifies that when a new
// subfolder containing encrypted files is added to an already-shared encrypted
// folder, the existing re-encryption key works for all new files.
func TestReKeyWorksForNewSubfolderAddedToSharedFolder(t *testing.T) {
	ownerMnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"
	recipientMnemonic := "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"

	recipient := newScheme(t, recipientMnemonic)
	recipient.InitForEncryption("filetype:audio")
	recipientPubKey, _ := recipient.GetPublicKey()

	// --- Owner shares /docs/ folder (generates re-encryption key) ---
	ownerForShare := newScheme(t, ownerMnemonic)
	ownerForShare.InitForEncryption("filetype:audio")
	reKey, err := ownerForShare.GetReGenKey(recipientPubKey, "filetype:audio")
	require.NoError(t, err)

	// --- Original files at share time ---
	// /docs/file1.txt
	originalFiles := []struct {
		path string
		data []byte
	}{
		{"/docs/file1.txt", []byte("existing file at share time")},
		{"/docs/existing-sub/readme.txt", []byte("existing subfolder file")},
	}

	for _, f := range originalFiles {
		owner := newScheme(t, ownerMnemonic)
		owner.InitForEncryption("filetype:audio")
		enc, err := owner.Encrypt(f.data)
		require.NoError(t, err)

		rec := newScheme(t, recipientMnemonic)
		rec.InitForEncryption("filetype:audio")
		dec := blobberReEncryptAndDecrypt(t, enc, owner.GetEncryptedKey(), reKey, recipientPubKey, rec)
		require.Equal(t, string(f.data), string(dec), "original file %s decrypts", f.path)
	}

	// --- Later: owner adds /docs/new-project/ subfolder with multiple files ---
	newFiles := []struct {
		path string
		data []byte
	}{
		{"/docs/new-project/design.md", []byte("# Design Doc\nNew project design")},
		{"/docs/new-project/spec.pdf", []byte("PDF binary data for spec")},
		{"/docs/new-project/src/main.go", []byte("package main\nfunc main() {}")},
		{"/docs/new-project/src/utils/helper.go", []byte("package utils\nfunc Helper() {}")},
	}

	for _, f := range newFiles {
		// Each new file gets its own random T → unique C1
		owner := newScheme(t, ownerMnemonic)
		owner.InitForEncryption("filetype:audio")
		fileC1 := owner.GetEncryptedKey()

		enc, err := owner.Encrypt(f.data)
		require.NoError(t, err)

		// Recipient decrypts using the ORIGINAL folder reKey
		rec := newScheme(t, recipientMnemonic)
		rec.InitForEncryption("filetype:audio")
		dec := blobberReEncryptAndDecrypt(t, enc, fileC1, reKey, recipientPubKey, rec)
		require.Equal(t, string(f.data), string(dec),
			"folder reKey works for new subfolder file %s", f.path)
	}

	// --- Also: update an existing file ---
	ownerUpdate := newScheme(t, ownerMnemonic)
	ownerUpdate.InitForEncryption("filetype:audio") // new random T
	updatedC1 := ownerUpdate.GetEncryptedKey()
	updatedData := []byte("file1.txt updated content after subfolder was added")
	encUpdated, err := ownerUpdate.Encrypt(updatedData)
	require.NoError(t, err)

	recUpdate := newScheme(t, recipientMnemonic)
	recUpdate.InitForEncryption("filetype:audio")
	decUpdated := blobberReEncryptAndDecrypt(t, encUpdated, updatedC1, reKey, recipientPubKey, recUpdate)
	require.Equal(t, string(updatedData), string(decUpdated),
		"folder reKey works for updated file with new C1")
}

func BenchmarkEncrypt(t *testing.B) {
	mnemonic := "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"
	encscheme := NewEncryptionScheme()
	_, err := encscheme.Initialize(mnemonic)
	require.Nil(t, err)
	encscheme.InitForEncryption("filetype:audio")
	for i := 0; i < 10000; i++ {
		dataToEncrypt := make([]byte, fileref.CHUNK_SIZE)
		read, err := rand.Read(dataToEncrypt)
		require.Nil(t, err, "Error in reading random data", read)

		_, err = encscheme.Encrypt(dataToEncrypt)
		require.Nil(t, err)
		require.Equal(t, len(dataToEncrypt), fileref.CHUNK_SIZE)
	}
}

func BenchmarkReEncryptAndReDecrypt(t *testing.B) {
	client_mnemonic := "travel twenty hen negative fresh sentence hen flat swift embody increase juice eternal satisfy want vessel matter honey video begin dutch trigger romance assault"
	client_encscheme := NewEncryptionScheme()
	_, err := client_encscheme.Initialize(client_mnemonic)
	require.Nil(t, err)
	client_encscheme.InitForEncryption("filetype:audio")
	client_enc_pub_key, err := client_encscheme.GetPublicKey()
	require.Nil(t, err)

	// seller uploads and blobber encrypts the data
	blobber_mnemonic := "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"
	blobber_encscheme := NewEncryptionScheme()
	_, err = blobber_encscheme.Initialize(blobber_mnemonic)
	require.Nil(t, err)
	blobber_encscheme.InitForEncryption("filetype:audio")
	// buyer requests data from blobber, blobber reencrypts the data with regen key using buyer public key
	regenkey, err := blobber_encscheme.GetReGenKey(client_enc_pub_key, "filetype:audio")
	require.Nil(t, err)
	for i := 0; i < 10000; i++ {
		dataToEncrypt := make([]byte, fileref.CHUNK_SIZE)
		rand.Read(dataToEncrypt) //nolint
		enc_msg, err := blobber_encscheme.Encrypt(dataToEncrypt)
		require.Nil(t, err)
		reenc_msg, err := blobber_encscheme.ReEncrypt(enc_msg, regenkey, client_enc_pub_key)
		require.Nil(t, err)

		client_decryption_scheme := NewEncryptionScheme()
		_, err = client_decryption_scheme.Initialize(client_mnemonic)
		require.Nil(t, err)
		err = client_decryption_scheme.InitForDecryption("filetype:audio", enc_msg.EncryptedKey)
		require.Nil(t, err)

		_, err = client_decryption_scheme.ReDecrypt(reenc_msg)
		require.Nil(t, err)
	}
}
