package main

import (
	"encoding/json"
	"strings"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/zcncrypto"
)

var _config localConfig
// var logging logger.Logger

const (
	StatusSuccess      int = 0
	StatusNetworkError int = 1
	// TODO: Change to specific error
	StatusError            int = 2
	StatusRejectedByUser   int = 3
	StatusInvalidSignature int = 4
	StatusAuthError        int = 5
	StatusAuthVerifyFailed int = 6
	StatusAuthTimeout      int = 7
	StatusUnknown          int = -1
)


// Split keys from the primary master key
func splitKeys(privateKey string, numSplits int) (string, error) {
	if _config.chain.SignatureScheme != "bls0chain" {
		return "", errors.New("", "signature key doesn't support split key")
	}
	sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
	err := sigScheme.SetPrivateKey(privateKey)
	if err != nil {
		return "", errors.Wrap(err, "set private key failed")
	}
	w, err := sigScheme.SplitKeys(numSplits)
	if err != nil {
		return "", errors.Wrap(err, "split key failed.")
	}
	wStr, err := w.Marshal()
	if err != nil {
		return "", errors.Wrap(err, "wallet encoding failed.")
	}
	return wStr, nil
}


// SetWalletInfo should be set before any transaction or client specific APIs
// splitKeyWallet parameter is valid only if SignatureScheme is "BLS0Chain"
//
//	# Inputs
//  - jsonWallet: json format of wallet
//  - splitKeyWallet: if wallet keys is split
func setWalletInfo(jsonWallet string, splitKeyWallet bool) error {
	err := json.Unmarshal([]byte(jsonWallet), &_config.wallet)
	if err == nil {
		if _config.chain.SignatureScheme == "bls0chain" {
			_config.isSplitWallet = splitKeyWallet
		}
		_config.isValidWallet = true
	}
	return err
}

// SetAuthUrl will be called by app to set zauth URL to SDK.
// # Inputs
// - url: the url of zAuth server
func setAuthUrl(url string) error {
	if !_config.isSplitWallet {
		return errors.New("", "wallet type is not split key")
	}
	if url == "" {
		return errors.New("", "invalid auth url")
	}
	_config.authUrl = strings.TrimRight(url, "/")
	return nil
}
