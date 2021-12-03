//go:build js && wasm
// +build js,wasm

package main

import (
	"errors"
	"fmt"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zboxcore/client"

	"syscall/js"
)

//-----------------------------------------------------------------------------

func main() {
	fmt.Printf("0CHAIN - GOSDK (version=%v)\n", version.VERSIONSTR)

	window := js.Global()

	zcn := window.Get("__zcn_wasm__")
	if !(zcn.IsNull() || zcn.IsUndefined()) {
		sdk := zcn.Get("sdk")
		jsClient := zcn.Get("js")

		sign := jsClient.Get("sign")

		signer := func(hash string) (string, error) {
			result, err := jsbridge.Await(sign.Invoke(hash))

			if len(err) > 0 && !err[0].IsNull() {
				return "", errors.New("sign: " + err[0].String())
			}
			return result[0].String(), nil
		}

		fire := jsClient.Get("fireTransactionAdd")

		fireTransactionAdd = func(txn *transaction.Transaction) {
			jsbridge.Await(fire.Invoke(jsbridge.NewObject(txn), client.GetClientID()))
		}

		//update sign with js sign
		zcncrypto.Sign = signer
		client.Sign = signer

		// tiny wasm sdk with new methods

		jsbridge.BindAsyncFuncs(sdk, map[string]interface{}{
			//sdk
			"init":                  Init,
			"setWallet":             SetWallet,
			"getEncryptedPublicKey": GetEncryptedPublicKey,

			//blobber
			"delete": Delete,
			"rename": Rename,
			"copy":   Copy,
			"move":   Move,
			"share":  Share,
		})

		fmt.Println("__wasm_initialized__ = true;")
		zcn.Set("__wasm_initialized__", true)
	}

	<-make(chan bool)

	jsbridge.Close()
}
