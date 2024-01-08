package zcncore

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/0chain/gosdk/core/transaction"
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

func TestDecodeJS(t *testing.T) {
	auth := []byte(`{"hash":"7caa2cdc4fa08c6191ac397d98333459c9c41cd499fc2623bcbda4ba4eb68708","version":"1.0","client_id":"586759fdf935cc9212ed29d36737ebd8173bf74632048a26ed12e0708edc2957","public_key":"2d1a9d9b109ceebcdae7c2c8664a4789efd2674f68401a3b9e73fa7e184c6f21cccaa583eea17e63ab7677b548a7d96352ae261512ab4868e72ff24299040f22","to_client_id":"2bff5112812e079f37f278f7436eceb08c7e3e3eae68076b6b6f396c7e5d641f","chain_id":"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe","transaction_data":"{\"note\":\"\"}","transaction_value":100000000000,"signature":"47b720e6f40e555355b696be5468f628ece680b7c23c845bebb3dfc7a2b957a1","creation_date":1704676947,"transaction_type":0,"transaction_fee":100000,"transaction_nonce":1,"txn_output_hash":"","transaction_status":0}`)
	var tx transaction.Transaction
	err := json.Unmarshal(auth, &tx)
	log.Println(err)
}
