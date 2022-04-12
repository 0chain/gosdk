//go:build js && wasm
// +build js,wasm

package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"

	"syscall/js"
)

//-----------------------------------------------------------------------------

func main() {
	fmt.Printf("0CHAIN - GOSDK (version=%v)\n", version.VERSIONSTR)
	sys.Files = sys.NewMemFS()
	sdkLogger = sdk.GetLogger()
	zcnLogger = zcncore.GetLogger()

	window := js.Global()

	zcn := window.Get("__zcn_wasm__")
	if !(zcn.IsNull() || zcn.IsUndefined()) {

		jsProxy := zcn.Get("jsProxy")
		// import functions from js object
		if !(jsProxy.IsNull() || jsProxy.IsUndefined()) {
			sign := jsProxy.Get("sign")

			if !(sign.IsNull() || sign.IsUndefined()) {
				signFunc := func(hash string) (string, error) {
					result, err := jsbridge.Await(sign.Invoke(hash))

					if len(err) > 0 && !err[0].IsNull() {
						return "", errors.New("sign: " + err[0].String())
					}
					return result[0].String(), nil
				}

				//update sign with js sign
				zcncrypto.Sign = signFunc
				sys.Sign = func(hash, signatureScheme string, keys []sys.KeyPair) (string, error) {
					// js already has signatureScheme and keys
					return signFunc(hash)
				}
			} else {
				PrintError("__zcn_wasm__.jsProxy.sign is not installed yet")
			}

			createObjectURL := jsProxy.Get("createObjectURL")
			if !(createObjectURL.IsNull() || createObjectURL.IsUndefined()) {

				CreateObjectURL = func(buf []byte, mimeType string) string {
					arrayBuffer := js.Global().Get("ArrayBuffer").New(len(buf))

					uint8Array := js.Global().Get("Uint8Array").New(arrayBuffer)

					js.CopyBytesToJS(uint8Array, buf)

					result, err := jsbridge.Await(createObjectURL.Invoke(uint8Array, mimeType))
					if len(err) > 0 && !err[0].IsNull() {
						PrintError(err[0].String())
						return ""
					}

					return result[0].String()
				}
			} else {
				PrintError("__zcn_wasm__.jsProxy.createObjectURL is not installed yet")
			}

			sleep := jsProxy.Get("sleep")
			if !(sleep.IsNull() || sleep.IsUndefined()) {
				sys.Sleep = func(d time.Duration) {
					ms := d.Milliseconds()
					jsbridge.Await(sleep.Invoke(ms))
				}
			} else {
				sys.Sleep = func(d time.Duration) {
					PrintInfo("sleep is not bridged to js method. it doesn't work")
				}
				PrintError("__zcn_wasm__.jsProxy.sleep is not installed yet")
			}
		} else {
			PrintError("__zcn_wasm__.jsProxy is not installed yet")
		}

		// tiny wasm sdk with new methods
		sdk := zcn.Get("sdk")
		// register go functions on wasm.sdk
		if !(sdk.IsNull() || sdk.IsUndefined()) {
			jsbridge.BindAsyncFuncs(sdk, map[string]interface{}{
				//sdk
				"init":                  Init,
				"setWallet":             SetWallet,
				"getEncryptedPublicKey": GetEncryptedPublicKey,
				"hideLogs":              hideLogs,
				"showLogs":              showLogs,

				//blobber
				"delete":   Delete,
				"rename":   Rename,
				"copy":     Copy,
				"move":     Move,
				"share":    Share,
				"download": Download,
				"upload":   Upload,

				// zcn txn
				"commitFileMetaTxn":   CommitFileMetaTxn,
				"commitFolderMetaTxn": CommitFolderMetaTxn,

				// player
				"play":           Play,
				"stop":           Stop,
				"getNextSegment": GetNextSegment,
			})

			fmt.Println("__wasm_initialized__ = true;")
			zcn.Set("__wasm_initialized__", true)
		} else {
			PrintError("__zcn_wasm__.sdk is not installed yet")
		}

	}

	hideLogs()

	<-make(chan bool)

	jsbridge.Close()
}
