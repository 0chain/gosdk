//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"

	"syscall/js"
)

//-----------------------------------------------------------------------------

var (
	signMutex sync.Mutex
	signCache = make(map[string]string)
)

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
			jsSign := jsProxy.Get("sign")

			if !(jsSign.IsNull() || jsSign.IsUndefined()) {
				signFunc := func(hash string) (string, error) {
					c := client.GetClient()
					pk := c.Keys[0].PrivateKey
					result, err := jsbridge.Await(jsSign.Invoke(hash, pk))

					if len(err) > 0 && !err[0].IsNull() {
						return "", errors.New("sign: " + err[0].String())
					}
					return result[0].String(), nil
				}

				//update sign with js sign
				zcncrypto.Sign = signFunc
				zcncore.SignFn = signFunc
				sys.Sign = func(hash, signatureScheme string, keys []sys.KeyPair) (string, error) {
					// js already has signatureScheme and keys
					return signFunc(hash)
				}

				sys.SignWithAuth = func(hash, signatureScheme string, keys []sys.KeyPair) (string, error) {
					sig, err := sys.Sign(hash, signatureScheme, keys)
					if err != nil {
						return "", fmt.Errorf("failed to sign with split key: %v", err)
					}

					data, err := json.Marshal(struct {
						Hash      string `json:"hash"`
						Signature string `json:"signature"`
						ClientID  string `json:"client_id"`
					}{
						Hash:      hash,
						Signature: sig,
						ClientID:  client.GetClient().ClientID,
					})
					if err != nil {
						return "", err
					}

					if sys.AuthCommon == nil {
						return "", errors.New("authCommon is not set")
					}

					rsp, err := sys.AuthCommon(string(data))
					if err != nil {
						return "", err
					}

					var sigpk struct {
						Sig string `json:"sig"`
					}

					err = json.Unmarshal([]byte(rsp), &sigpk)
					if err != nil {
						return "", err
					}

					return sigpk.Sig, nil
				}
			} else {
				PrintError("__zcn_wasm__.jsProxy.sign is not installed yet")
			}

			jsVerify := jsProxy.Get("verify")

			if !(jsVerify.IsNull() || jsVerify.IsUndefined()) {
				verifyFunc := func(signature, hash string) (bool, error) {
					result, err := jsbridge.Await(jsVerify.Invoke(signature, hash))

					if len(err) > 0 && !err[0].IsNull() {
						return false, errors.New("verify: " + err[0].String())
					}
					return result[0].Bool(), nil
				}

				//update Verify with js sign
				sys.Verify = verifyFunc
			} else {
				PrintError("__zcn_wasm__.jsProxy.verify is not installed yet")
			}

			jsVerifyWith := jsProxy.Get("verifyWith")

			if !(jsVerifyWith.IsNull() || jsVerifyWith.IsUndefined()) {
				verifyFuncWith := func(pk, signature, hash string) (bool, error) {
					result, err := jsbridge.Await(jsVerifyWith.Invoke(pk, signature, hash))

					if len(err) > 0 && !err[0].IsNull() {
						return false, errors.New("verify: " + err[0].String())
					}
					return result[0].Bool(), nil
				}

				//update Verify with js sign
				sys.VerifyWith = verifyFuncWith
			} else {
				PrintError("__zcn_wasm__.jsProxy.verifyWith is not installed yet")
			}

			jsAddSignature := jsProxy.Get("addSignature")
			if !(jsAddSignature.IsNull() || jsAddSignature.IsUndefined()) {
				zcncore.AddSignature = func(privateKey, signature, hash string) (string, error) {
					result, err := jsbridge.Await(jsAddSignature.Invoke(privateKey, signature, hash))
					if len(err) > 0 && !err[0].IsNull() {
						return "", errors.New("add signature: " + err[0].String())
					}

					return result[0].String(), nil
				}
			} else {
				PrintError("__zcn_wasm__.jsProxy.addSignature is not installed yet")
			}

			jsCreateObjectURL := jsProxy.Get("createObjectURL")
			if !(jsCreateObjectURL.IsNull() || jsCreateObjectURL.IsUndefined()) {

				CreateObjectURL = func(buf []byte, mimeType string) string {

					arrayBuffer := js.Global().Get("ArrayBuffer").New(len(buf))

					uint8Array := js.Global().Get("Uint8Array").New(arrayBuffer)

					js.CopyBytesToJS(uint8Array, buf)

					result, err := jsbridge.Await(jsCreateObjectURL.Invoke(uint8Array, mimeType))

					if len(err) > 0 && !err[0].IsNull() {
						PrintError(err[0].String())
						return ""
					}

					return result[0].String()
				}
			} else {
				PrintError("__zcn_wasm__.jsProxy.createObjectURL is not installed yet")
			}

			sys.Sleep = func(d time.Duration) {
				<-time.After(d)
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
				"init":                   initSDKs,
				"setWallet":              setWallet,
				"getPublicEncryptionKey": zcncore.GetPublicEncryptionKey,
				"hideLogs":               hideLogs,
				"showLogs":               showLogs,
				"getUSDRate":             getUSDRate,
				"isWalletID":             isWalletID,
				"getVersion":             getVersion,
				"getLookupHash":          getLookupHash,
				"createThumbnail":        createThumbnail,
				"makeSCRestAPICall":      makeSCRestAPICall,

				//blobber
				"delete":                    Delete,
				"share":                     Share,
				"multiDownload":             multiDownload,
				"upload":                    upload,
				"setUploadMode":             setUploadMode,
				"multiUpload":               multiUpload,
				"multiOperation":            MultiOperation,
				"listObjects":               listObjects,
				"listObjectsFromAuthTicket": listObjectsFromAuthTicket,
				"createDir":                 createDir,
				"downloadBlocks":            downloadBlocks,
				"getFileStats":              getFileStats,
				"updateBlobberSettings":     updateBlobberSettings,
				"getRemoteFileMap":          getRemoteFileMap,
				"getBlobbers":               getBlobbers,
				"getcontainers":             GetContainers,
				"updatecontainer":           UpdateContainer,
				"searchcontainer":           SearchContainer,
				"updateForbidAllocation":    UpdateForbidAllocation,
				"send":                      send,
				"cancelUpload":              cancelUpload,
				"pauseUpload":               pauseUpload,

				// player
				"play":           play,
				"stop":           stop,
				"getNextSegment": getNextSegment,

				//allocation
				"createAllocation":           createAllocation,
				"getAllocationBlobbers":      getAllocationBlobbers,
				"getBlobberIds":              getBlobberIds,
				"listAllocations":            listAllocations,
				"getAllocation":              getAllocation,
				"reloadAllocation":           reloadAllocation,
				"transferAllocation":         transferAllocation,
				"freezeAllocation":           freezeAllocation,
				"cancelAllocation":           cancelAllocation,
				"updateAllocation":           updateAllocation,
				"updateAllocationWithRepair": updateAllocationWithRepair,
				"getAllocationMinLock":       getAllocationMinLock,
				"getUpdateAllocationMinLock": getUpdateAllocationMinLock,
				"getAllocationWith":          getAllocationWith,
				"createfreeallocation":       createfreeallocation,

				// readpool
				"getReadPoolInfo": getReadPoolInfo,
				"lockReadPool":    lockReadPool,
				"unLockReadPool":  unLockReadPool,
				"createReadPool":  createReadPool,

				// claim rewards
				"collectRewards": collectRewards,

				// stakepool
				"getSkatePoolInfo": getSkatePoolInfo,
				"lockStakePool":    lockStakePool,
				"unlockStakePool":  unlockStakePool,

				// writepool
				"lockWritePool": lockWritePool,

				"decodeAuthTicket": decodeAuthTicket,
				"allocationRepair": allocationRepair,

				//smartcontract
				"executeSmartContract": executeSmartContract,
				"faucet":               faucet,

				// bridge
				"initBridge":                    initBridge,
				"burnZCN":                       burnZCN,
				"mintZCN":                       mintZCN,
				"getMintWZCNPayload":            getMintWZCNPayload,
				"getNotProcessedWZCNBurnEvents": getNotProcessedWZCNBurnEvents,
				"getNotProcessedZCNBurnTickets": getNotProcessedZCNBurnTickets,
				"estimateBurnWZCNGasAmount":     estimateBurnWZCNGasAmount,
				"estimateMintWZCNGasAmount":     estimateMintWZCNGasAmount,
				"estimateGasPrice":              estimateGasPrice,

				//zcn
				"getWalletBalance": getWalletBalance,

				//0box api
				"getCsrfToken":     getCsrfToken,
				"createJwtSession": createJwtSession,
				"createJwtToken":   createJwtToken,
				"refreshJwtToken":  refreshJwtToken,

				//split key
				"splitKeys":     splitKeys,
				"setWalletInfo": setWalletInfo,
				"setAuthUrl":    setAuthUrl,

				"registerAuthorizer": js.FuncOf(registerAuthorizer),
				"registerAuthCommon": js.FuncOf(registerAuthCommon),
				"callAuth":           js.FuncOf(callAuth),
				"authResponse":       authResponse,

				// zauth
				"registerZauthServer": registerZauthServer,
				// zvault
				"zvaultNewWallet": zvaultNewWallet,
				"zvaultNewSplit":  zvaultNewSplit,
				"zvaultStoreKey":  zvaultStoreKey,
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
