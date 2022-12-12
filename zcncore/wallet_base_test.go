package zcncore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptAndDecrypt(t *testing.T) {
	key := "passphrase1111111111111111111111"
	mnemonics := "glare mistake gun joke bid spare across diagram wrap cube swear cactus cave repeat you brave few best wild lion pitch pole original wasp"

	encryptedMnemonics, err := Encrypt(key, mnemonics)

	require.Nil(t, err)
	require.NotEmpty(t, encryptedMnemonics)

	decryptedMnemonics, err := Decrypt(key, encryptedMnemonics)
	require.Nil(t, err)
	require.Equal(t, mnemonics, decryptedMnemonics)
}
