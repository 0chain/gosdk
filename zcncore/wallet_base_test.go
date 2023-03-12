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
	passphrase := "12345611111111111111111111111111"
	mnemonics := "cube reward february gym summer peanut marble slow puppy picnic cart waste aspect either anchor vacant horse north border wonder stamp mansion steak magic"

	enc, err := CryptoJsEncrypt(passphrase, mnemonics)
	require.Nil(t, err)

	dec, err := CryptoJsDecrypt(passphrase, string(enc))
	require.Nil(t, err)
	require.Equal(t, mnemonics, string(dec))
	encryptedMnemonics := "U2FsdGVkX1/Dz58HfdXjHGJioPZ8bnEWIfa0dZcz0JuizI/Tu1+1ncVv60f4w53VimvKG0dC5zhVFQC8dt7K7Lydutu/pquTCDfKt3AUK2iJ5mjN1n4rCvp5IMG+5fKuVyY0z+PbH5MgyJdAF1Fbsi3X+ccfd/ZB9jg6deHpneHDMxhRzuGKcKUuWA6+D/peQTGCmHCLbAPYswFUeF0Elcmgi1mx69UYeM1qgfumuFs="
	dec, err = CryptoJsDecrypt(passphrase, encryptedMnemonics)
	require.Nil(t, err)
	require.Equal(t, mnemonics, string(dec))
}
