package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/zcncrypto"
)

var _config localConfig
var logging logger.Logger

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

type AuthCallback interface {
	// This call back gives the status of the Two factor authenticator(zauth) setup.
	OnSetupComplete(status int, err string)
}

func GetLogger() *logger.Logger {
	return &logging
}

// CloseLog closes log file
func CloseLog() {
	logging.Close()
}


// Split keys from the primary master key
func SplitKeys(privateKey string, numSplits int) (string, error) {
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

// SetupAuth prepare auth app with clientid, key and a set of public, private key and local publickey
// which is running on PC/Mac.
func SetupAuth(authHost, clientID, clientKey, publicKey, privateKey, localPublicKey string, cb AuthCallback) error {
	go func() {
		authHost = strings.TrimRight(authHost, "/")
		data := map[string]string{"client_id": clientID, "client_key": clientKey, "public_key": publicKey, "private_key": privateKey, "peer_public_key": localPublicKey}
		req, err := util.NewHTTPPostRequest(authHost+"/setup", data)
		if err != nil {
			logging.Error("new post request failed. ", err.Error())
			return
		}
		res, err := req.Post()
		if err != nil {
			logging.Error(authHost+"send error. ", err.Error())
		}
		if res.StatusCode != http.StatusOK {
			cb.OnSetupComplete(StatusError, res.Body)
			return
		}
		cb.OnSetupComplete(StatusSuccess, "")
	}()
	return nil
}

// SetWalletInfo should be set before any transaction or client specific APIs
// splitKeyWallet parameter is valid only if SignatureScheme is "BLS0Chain"
//
//	# Inputs
//	- jsonWallet: json format of wallet
//	{
//	"client_id":"30764bcba73216b67c36b05a17b4dd076bfdc5bb0ed84856f27622188c377269",
//	"client_key":"1f495df9605a4479a7dd6e5c7a78caf9f9d54e3a40f62a3dd68ed377115fe614d8acf0c238025f67a85163b9fbf31d10fbbb4a551d1cf00119897edf18b1841c",
//	"keys":[
//		{"public_key":"1f495df9605a4479a7dd6e5c7a78caf9f9d54e3a40f62a3dd68ed377115fe614d8acf0c238025f67a85163b9fbf31d10fbbb4a551d1cf00119897edf18b1841c","private_key":"41729ed8d82f782646d2d30b9719acfd236842b9b6e47fee12b7bdbd05b35122"}
//	],
//	"mnemonics":"glare mistake gun joke bid spare across diagram wrap cube swear cactus cave repeat you brave few best wild lion pitch pole original wasp",
//	"version":"1.0",
//	"date_created":"1662534022",
//	"nonce":0
//	}
//
// - splitKeyWallet: if wallet keys is split
func SetWalletInfo(jsonWallet string, splitKeyWallet bool) error {
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
//   - url: the url of zAuth server
func SetAuthUrl(url string) error {
	if !_config.isSplitWallet {
		return errors.New("", "wallet type is not split key")
	}
	if url == "" {
		return errors.New("", "invalid auth url")
	}
	_config.authUrl = strings.TrimRight(url, "/")
	return nil
}
