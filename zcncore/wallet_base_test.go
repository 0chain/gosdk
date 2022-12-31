package zcncore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCFBEncryption(t *testing.T) {
	key := "passphrase1111111111111111111111"
	mnemonics := "glare mistake gun joke bid spare across diagram wrap cube swear cactus cave repeat you brave few best wild lion pitch pole original wasp"

	encryptedMnemonics, err := Encrypt(key, mnemonics)

	require.Nil(t, err)
	require.NotEmpty(t, encryptedMnemonics)

	decryptedMnemonics, err := Decrypt(key, encryptedMnemonics)
	require.Nil(t, err)
	require.Equal(t, mnemonics, decryptedMnemonics)
}

func TestCBCEncryption(t *testing.T) {
	passphrase := "abcabc11111111111111111111111111"
	mnemonics := "dog peanut knee mule crouch mushroom easy climb good grid squeeze naive enter genuine bone file talent topic robust long siren depth suspect wide"

	enc, err := CryptoJsEncrypt(passphrase, mnemonics)
	require.Nil(t, err)

	dec, err := CryptoJsDecrypt(passphrase, string(enc))
	require.Nil(t, err)
	require.Equal(t, mnemonics, string(dec))

	encryptedMnemonics := "U2FsdGVkX1/1vVZfJJOqBRInSChGhVr3YQ3gnQFpCGNB0cXAVM2900GCiQZRLIibgJZZKNqAlopo1jT4LCe9qLRVS0CyTWhblPjk/XchhJENoGRdbGK0vtWBVtFbXaAvMV55ssinpdMLnKphgLOpH4wVHSkPIAMVGrMviNcycAtpKBhJTaSM2UtF5M9mjkLZ0hdCV7Ho+yqHi6ZXhGnJdsVM/yJBAAyG1PzdRfAHiys="
	dec, err = CryptoJsDecrypt(passphrase, encryptedMnemonics)
	require.Nil(t, err)
	require.Equal(t, mnemonics, string(dec))
}
