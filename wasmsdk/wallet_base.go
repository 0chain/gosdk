package main

import (
	"encoding/base64"

	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/0chain/gosdk/zcncore"
)

// Split keys from the primary master key

// splitKeys splits the primary master key into n number of keys
//   - privateKey is the primary master key
//   - numSplits is the number of keys to split into
//
// nolint: unused
func splitKeys(privateKey string, numSplits int) (string, error) {
	wStr, err := zcncore.SplitKeys(privateKey, numSplits)
	return wStr, err
}

// scryptEncrypt encrypts plaintext with a password (scrypt key derivation +
// ChaCha20-Poly1305 AEAD). Returns base64(salt|nonce|ciphertext). Used to
// protect the client half of a split key before storing it in 0box.
//
// nolint: unused
func scryptEncrypt(password, plaintext string) (string, error) {
	ciphertext, err := zboxutil.ScryptEncrypt([]byte(password), []byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// scryptDecrypt decrypts a base64 ciphertext produced by scryptEncrypt.
//
// nolint: unused
func scryptDecrypt(password, ciphertext string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	plaintext, err := zboxutil.ScryptDecrypt([]byte(password), raw)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// setWalletInfo should be set before any transaction or client specific APIs.
// splitKeyWallet parameter is valid only if SignatureScheme is "BLS0Chain"
//   - jsonWallet: json format of wallet
//   - splitKeyWallet: if wallet keys is split
//
// nolint: unused
func setWalletInfo(jsonWallet string, splitKeyWallet bool) bool {
	err := zcncore.SetWalletInfo(jsonWallet, "bls0chain", splitKeyWallet)
	if err == nil {
		return true
	} else {
		return false
	}
}

// setAuthUrl will be called by app to set zauth URL to SDK.
//   - url: the url of zAuth server
//
// nolint: unused
func setAuthUrl(url string) bool {
	err := zcncore.SetAuthUrl(url)
	if err == nil {
		return true
	} else {
		return false
	}
}
