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
	"path/filepath"

	"github.com/0chain/gosdk/zboxcore/client"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"

	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/zcncrypto"
)

func main() {
	sdk.SetLogFile(filepath.Join(getHomeDir(), ".zcn", "zbox.log"), true)
	zcncore.SetLogFile(filepath.Join(getHomeDir(), ".zcn", "zcn.log"), true)

	sdk.GetLogger().Info("0Chain Windows SDK is ready")
}

// Init - init zbox/zcn sdk from config
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
//     }
//
//export Init
func Init(configJson *C.char) error {

	l.Logger.Info("Start InitStorageSDK")
	configObj := &conf.Config{}
	js := C.GoString(configJson)
	err := json.Unmarshal([]byte(js), configObj)
	if err != nil {
		l.Logger.Error(err)
		return err
	}
	err = zcncore.InitZCNSDK(configObj.BlockWorker, configObj.SignatureScheme)
	if err != nil {
		l.Logger.Error(err)
		return err
	}
	l.Logger.Info("InitZCNSDK success")
	l.Logger.Info(configObj.BlockWorker)
	l.Logger.Info(configObj.ChainID)
	l.Logger.Info(configObj.SignatureScheme)
	l.Logger.Info(configObj.PreferredBlobbers)
	err = sdk.InitStorageSDK(js, configObj.BlockWorker, configObj.ChainID, configObj.SignatureScheme, configObj.PreferredBlobbers, 0)
	if err != nil {
		l.Logger.Error(err)
		return err
	}
	l.Logger.Info("InitStorageSDK success")
	l.Logger.Info("Init successful")
	return nil
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

	return WithJSON(client.SignHash(hash, scheme, []sys.KeyPair{{
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
