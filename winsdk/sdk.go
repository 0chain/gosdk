package main

/*
#include <stdlib.h>
*/
import (
	"C"
)

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/0chain/gosdk/zboxapi"
	"github.com/0chain/gosdk/zboxcore/client"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/0chain/gosdk/zcncore"

	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/zcncrypto"
)

var log logger.Logger

func main() {

}

// SetLogFile - set log file
// ## Inputs
//   - file: the full path of log file
//
//export SetLogFile
func SetLogFile(file *C.char) *C.char {

	f, err := os.OpenFile(C.GoString(file), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return WithJSON(false, err)
	}

	sdk.GetLogger().SetLevel(logger.DEBUG)
	sdk.GetLogger().SetLogFile(f, true)

	zcncore.GetLogger().SetLevel(logger.DEBUG)
	zcncore.GetLogger().SetLogFile(f, true)

	zboxutil.GetLogger().SetLevel(logger.DEBUG)
	zboxutil.GetLogger().SetLogFile(f, true)

	log.SetLogFile(f, true)
	log.SetLevel(logger.DEBUG)

	return WithJSON(true, nil)
}

// InitSDK - init zbox/zcn sdk from config
//   - clientJson
//     {
//     "client_id":"8f6ce6457fc04cfb4eb67b5ce3162fe2b85f66ef81db9d1a9eaa4ffe1d2359e0",
//     "client_key":"c8c88854822a1039c5a74bdb8c025081a64b17f52edd463fbecb9d4a42d15608f93b5434e926d67a828b88e63293b6aedbaf0042c7020d0a96d2e2f17d3779a4",
//     "keys":[
//     {
//     "public_key":"c8c88854822a1039c5a74bdb8c025081a64b17f52edd463fbecb9d4a42d15608f93b5434e926d67a828b88e63293b6aedbaf0042c7020d0a96d2e2f17d3779a4",
//     "private_key":"72f480d4b1e7fb76e04327b7c2348a99a64f0ff2c5ebc3334a002aa2e66e8506"
//     }],
//     "mnemonics":"abandon mercy into make powder fashion butter ignore blade vanish plastic shock learn nephew matrix indoor surge document motor group barely offer pottery antenna",
//     "version":"1.0",
//     "date_created":"1668667145",
//     "nonce":0
//     }
//   - configJson
//     {
//     "block_worker": "https://dev.0chain.net/dns",
//     "signature_scheme": "bls0chain",
//     "min_submit": 50,
//     "min_confirmation": 50,
//     "confirmation_chain_length": 3,
//     "max_txn_query": 5,
//     "query_sleep_time": 5,
//     "preferred_blobbers": ["https://dev.0chain.net/blobber02","https://dev.0chain.net/blobber03"],
//     "chain_id":"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe",
//     "ethereum_node":"https://ropsten.infura.io/v3/xxxxxxxxxxxxxxx",
//     "zbox_host":"https://0box.dev.0chain.net",
//     "zbox_app_type":"vult",
//     "sharder_consensous": 2,
//     }
//
//export InitSDK
func InitSDK(configJson *C.char, clientJson *C.char) *C.char {

	l.Logger.Info("Start InitStorageSDK")

	configJs := C.GoString(configJson)
	clientJs := C.GoString(clientJson)

	configObj := &conf.Config{}

	l.Logger.Info("cfg: ", configJs)
	err := json.Unmarshal([]byte(configJs), configObj)
	if err != nil {
		l.Logger.Error(err)
		return WithJSON(false, err)
	}

	err = zcncore.InitZCNSDK(configObj.BlockWorker, configObj.SignatureScheme, func(cc *zcncore.ChainConfig) error {
		cc.BlockWorker = configObj.BlockWorker
		cc.ChainID = configObj.ChainID
		cc.ConfirmationChainLength = configObj.ConfirmationChainLength
		cc.MinConfirmation = configObj.MinConfirmation
		cc.EthNode = configObj.EthereumNode
		cc.MinSubmit = configObj.MinSubmit
		cc.SharderConsensous = configObj.SharderConsensous
		cc.SignatureScheme = configObj.SignatureScheme

		return nil
	})
	if err != nil {
		l.Logger.Error(err, configJs, clientJs)
		return WithJSON(false, err)
	}

	l.Logger.Info("InitZCNSDK success")
	l.Logger.Info(configObj.BlockWorker)
	l.Logger.Info(configObj.ChainID)
	l.Logger.Info(configObj.SignatureScheme)
	l.Logger.Info(configObj.PreferredBlobbers)

	err = sdk.InitStorageSDK(clientJs, configObj.BlockWorker, configObj.ChainID, configObj.SignatureScheme, configObj.PreferredBlobbers, 0)
	if err != nil {
		l.Logger.Error(err, configJs, clientJs)
		return WithJSON(false, err)
	}
	l.Logger.Info("InitStorageSDK success")

	zboxApiClient = zboxapi.NewClient()
	zboxApiClient.SetRequest(configObj.ZboxHost, configObj.ZboxAppType)
	c := client.GetClient()
	if c != nil {
		zboxApiClient.SetWallet(client.GetClientID(), client.GetClientPrivateKey(), client.GetClientPublicKey())
		l.Logger.Info("InitZboxApiClient success")
	} else {
		l.Logger.Info("InitZboxApiClient skipped")
	}

	l.Logger.Info("Init successful")
	return WithJSON(true, nil)
}

var ErrInvalidSignatureScheme = errors.New("invalid_signature_scheme")

// SignRequest sign data with private key and scheme
//
//	return
//		{
//			"error":"",
//			"result":"xxx",
//		}
//
//export SignRequest
func SignRequest(privateKey, signatureScheme, data *C.char) *C.char {
	key := C.GoString(privateKey)
	scheme := C.GoString(signatureScheme)
	d := C.GoString(data)

	hash := encryption.Hash(d)

	return WithJSON(sys.Sign(hash, scheme, []sys.KeyPair{{
		PrivateKey: key,
	}}))
}

// VerifySignature verify signature with public key, schema and data
// return
//
//	{
//		"error":"",
//		"result":true,
//	}
//
//export VerifySignature
func VerifySignature(publicKey, signatureScheme string, data string, signature string) *C.char {

	hash := encryption.Hash(data)

	signScheme := zcncrypto.NewSignatureScheme(signatureScheme)
	if signScheme != nil {
		err := signScheme.SetPublicKey(publicKey)
		if err != nil {
			return WithJSON(false, err)
		}
		return WithJSON(signScheme.Verify(signature, hash))
	}
	return WithJSON(false, ErrInvalidSignatureScheme)
}

// CryptoJsEncrypt encrypt message with AES+CCB
//
//	return
//		{
//			"error":"",
//			"result":"xxx",
//		}
//
//export CryptoJsEncrypt
func CryptoJsEncrypt(passphrase, message *C.char) *C.char {
	pass := C.GoString(passphrase)
	msg := C.GoString(message)

	return WithJSON(zcncore.CryptoJsEncrypt(pass, msg))
}

// CryptoJsDecrypt decrypt message with AES+CCB
//
//	return
//		{
//			"error":"",
//			"result":"xxx",
//		}
//
//export CryptoJsDecrypt
func CryptoJsDecrypt(passphrase, encryptedMessage *C.char) *C.char {
	pass := C.GoString(passphrase)
	msg := C.GoString(encryptedMessage)

	return WithJSON(zcncore.CryptoJsDecrypt(pass, msg))
}

// GetPublicEncryptionKey get public encryption key by mnemonic
//
//	return
//		{
//			"error":"",
//			"result":"xxxx",
//		}
//
//export GetPublicEncryptionKey
func GetPublicEncryptionKey(mnemonics *C.char) *C.char {
	m := C.GoString(mnemonics)
	return WithJSON(zcncore.GetPublicEncryptionKey(m))
}

// GetLookupHash get lookup hash with allocation id and path
// ## Inputs:
//   - allocationID
//   - path
//     return
//     {
//     "error":"",
//     "result":"xxxx",
//     }
//
//export GetLookupHash
func GetLookupHash(allocationID *C.char, path *C.char) *C.char {
	hash := getLookupHash(C.GoString(allocationID), C.GoString(path))
	return WithJSON(hash, nil)
}
