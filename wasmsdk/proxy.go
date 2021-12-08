//go:build js && wasm
// +build js,wasm

package main

import (
	"errors"
	"fmt"

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

		jsProxy := zcn.Get("jsProxy")
		if !(jsProxy.IsNull() || jsProxy.IsUndefined()) {
			sign := jsProxy.Get("sign")

			if !(sign.IsNull() || sign.IsUndefined()) {
				signer := func(hash string) (string, error) {
					result, err := jsbridge.Await(sign.Invoke(hash))

					if len(err) > 0 && !err[0].IsNull() {
						return "", errors.New("sign: " + err[0].String())
					}
					return result[0].String(), nil
				}

				//update sign with js sign
				zcncrypto.Sign = signer
				client.Sign = signer
			} else {
				PrintError("__zcn_wasm__.jsProxy.sign is not installed yet")
			}

		} else {
			PrintError("__zcn_wasm__.jsProxy is not installed yet")
		}

		// tiny wasm sdk with new methods
		sdk := zcn.Get("sdk")
		if !(sdk.IsNull() || sdk.IsUndefined()) {
			jsbridge.BindAsyncFuncs(sdk, map[string]interface{}{
				//sdk
				"init":                  Init,
				"setWallet":             SetWallet,
				"getEncryptedPublicKey": GetEncryptedPublicKey,

				//blobber
				"delete":   Delete,
				"rename":   Rename,
				"copy":     Copy,
				"move":     Move,
				"share":    Share,
				"download": Download,
			})

			fmt.Println("__wasm_initialized__ = true;")
			zcn.Set("__wasm_initialized__", true)
		} else {
			PrintError("__zcn_wasm__.sdk is not installed yet")
		}

	}

	<-make(chan bool)

	jsbridge.Close()
}
