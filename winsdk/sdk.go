package main

/*
#include <stdlib.h>
*/
import (
	"C"
)

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/0chain/gosdk/zboxapi"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	"github.com/0chain/gosdk/zcncore"

	"github.com/0chain/gosdk/core/client"
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
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()

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

	zboxapi.GetLogger().SetLevel(logger.DEBUG)
	zboxapi.GetLogger().SetLogFile(f, true)

	log.SetLogFile(f, true)
	log.SetLevel(logger.DEBUG)

	return WithJSON(true, nil)
}

// InitSDKs - init zcncore sdk and zboxapi client from config
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
//export InitSDKs
func InitSDKs(configJson *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()

	l.Logger.Info("Start InitSDKs")

	configJs := C.GoString(configJson)

	configObj := &conf.Config{}

	l.Logger.Info("cfg: ", configJs)
	err := json.Unmarshal([]byte(configJs), configObj)
	if err != nil {
		l.Logger.Error(err)
		return WithJSON(false, err)
	}

	err = client.Init(context.Background(), *configObj)

	if err != nil {
		l.Logger.Error(err, configJs)
		return WithJSON(false, err)
	}

	l.Logger.Info("InitZCNSDK success")
	l.Logger.Info(configObj.BlockWorker)
	l.Logger.Info(configObj.ChainID)
	l.Logger.Info(configObj.SignatureScheme)
	l.Logger.Info(configObj.PreferredBlobbers)

	if zboxApiClient == nil {
		zboxApiClient = zboxapi.NewClient()
	}

	zboxApiClient.SetRequest(configObj.ZboxHost, configObj.ZboxAppType)
	l.Logger.Info("Init ZBoxAPI Client success")

	return WithJSON(true, nil)
}

// InitWallet - init wallet for storage sdk and zboxapi client
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
//
//export InitWallet
func InitWallet(clientJson *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	l.Logger.Info("Start InitWallet")

	clientJs := C.GoString(clientJson)
	
	var w zcncrypto.Wallet
	err := json.Unmarshal([]byte(clientJs), &w)
	if err != nil {
		l.Logger.Error(err)
		return WithJSON(false, err)
	}
	err = client.SetWallet(w)
	if err != nil {
		l.Logger.Error(err)
		return WithJSON(false, err)
	}

	l.Logger.Info("InitWallet success")
	zboxApiClient.SetWallet(client.Wallet().ClientID, client.Wallet().Keys[0].PrivateKey, client.Wallet().ClientKey)
	l.Logger.Info("InitZboxApiClient success")
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
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
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
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
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
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
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
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
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
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
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
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	hash := getLookupHash(C.GoString(allocationID), C.GoString(path))
	return WithJSON(hash, nil)
}

// SetFFmpeg set the full file name of ffmpeg.exe
// ## Inputs:
//   - fullFileName
//     return
//     {
//     "error":"",
//     "result":true,
//     }
//
//export SetFFmpeg
func SetFFmpeg(fullFileName *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	f := C.GoString(fullFileName)

	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		return WithJSON(false, err)
	}
	sdk.CmdFFmpeg = C.GoString(fullFileName)
	return WithJSON(true, nil)
}

// GetFileContentType get content/MIME type of file
// ## Inputs:
//   - fullFileName
//     return
//     {
//     "error":"",
//     "result":true,
//     }
//
//export GetFileContentType
func GetFileContentType(file *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	f, err := os.Open(C.GoString(file))
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}
	defer f.Close()

	mime, err := zboxutil.GetFileContentType(f)
	if err != nil {
		return WithJSON("", err)
	}

	return WithJSON(mime, nil)
}
