//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/imageutil"
	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/core/screstapi"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"

	"io"
	"os"
	"syscall/js"
)


var CreateObjectURL func(buf []byte, mimeType string) string

// initSDKs init sdk with its parameters
//   - chainID is the chain id
//   - blockWorker is the block worker url, which is the DNS server used to locate the network nodes
//   - signatureScheme is the signature scheme used for signing transactions
//   - minConfirmation is the minimum number of confirmations required for a transaction to be considered final
//   - minSubmit is the minimum number of times a transaction must be submitted to the network
//   - confirmationChainLength is the number of blocks to wait for a transaction to be confirmed
//   - zboxHost is the url of the 0box service
//   - zboxAppType is the application type of the 0box service
//   - sharderconsensous is the number of sharders to reach consensus
//   - sharderconsensous is the number of sharders to reach consensus
func initSDKs(chainID, blockWorker, signatureScheme string,
	minConfirmation, minSubmit, confirmationChainLength int,
	zboxHost, zboxAppType string, sharderConsensous int, refresh bool) error {

	// Print the parameters beautified
	fmt.Printf("{ chainID: %s, blockWorker: %s, signatureScheme: %s, minConfirmation: %d, minSubmit: %d, confirmationChainLength: %d, zboxHost: %s, zboxAppType: %s, sharderConsensous: %d, refresh: %v }\n", chainID, blockWorker, signatureScheme, minConfirmation, minSubmit, confirmationChainLength, zboxHost, zboxAppType, sharderConsensous, refresh)

	zboxApiClient.SetRequest(zboxHost, zboxAppType)

	var miners, sharders []string
	if !refresh {
		miners, sharders = loadNetworkFromCache()
	}

	params := client.InitSdkOptions{
		WalletJSON:              "{}",
		BlockWorker:             blockWorker,
		ChainID:                 chainID,
		SignatureScheme:         signatureScheme,
		Nonce:                   int64(0),
		AddWallet:               false,
		MinConfirmation:         &minConfirmation,
		MinSubmit:               &minSubmit,
		SharderConsensous:       &sharderConsensous,
		ConfirmationChainLength: &confirmationChainLength,
		ZboxHost:                zboxHost,
		ZboxAppType:             zboxAppType,
		Miners:                  miners,
		Sharders:                sharders,
	}

	err := client.InitSDKWithWebApp(params)
	if err != nil {
		fmt.Println("wasm: InitStorageSDK ", err)
		return err
	}

	if refresh || len(miners) == 0 {
		saveNetworkToCache()
	}

	sdk.SetWasm()
	return nil
}

func loadNetworkFromCache() ([]string, []string) {
	storage := js.Global().Get("localStorage")
	if storage.IsNull() || storage.IsUndefined() {
		return nil, nil
	}
	val := storage.Call("getItem", "zcn_network_info")
	if val.IsNull() || val.IsUndefined() {
		return nil, nil
	}
	jsonStr := val.String()
	type networkInfo struct {
		Miners   []string `json:"miners"`
		Sharders []string `json:"sharders"`
	}
	var info networkInfo
	if err := json.Unmarshal([]byte(jsonStr), &info); err != nil {
		return nil, nil
	}
	return info.Miners, info.Sharders
}

func saveNetworkToCache() {
	node, err := client.GetNode()
	if err != nil {
		return
	}
	network := node.Network()
	if network == nil {
		return
	}
	type networkInfo struct {
		Miners   []string `json:"miners"`
		Sharders []string `json:"sharders"`
	}
	info := networkInfo{
		Miners:   network.Miners,
		Sharders: network.Sharders,
	}
	data, err := json.Marshal(info)
	if err != nil {
		return
	}
	storage := js.Global().Get("localStorage")
	if !storage.IsNull() && !storage.IsUndefined() {
		storage.Call("setItem", "zcn_network_info", string(data))
	}
}

// getVersion retrieve the sdk version
func getVersion() string {
	return sdk.GetVersion()
}

func getWasmType() string {
	return "normal"
}

var sdkLogger *logger.Logger
var zcnLogger *logger.Logger
var logEnabled = false

// showLogs enable logging
func showLogs() {
	zcnLogger.SetLevel(logger.DEBUG)
	sdkLogger.SetLevel(logger.DEBUG)

	zcnLogger.SetLogFile(os.Stdout, true)
	sdkLogger.SetLogFile(os.Stdout, true)

	logEnabled = true
}

// hideLogs disable logging
func hideLogs() {
	zcnLogger.SetLevel(logger.ERROR)
	sdkLogger.SetLevel(logger.ERROR)

	zcnLogger.SetLogFile(io.Discard, false)
	sdkLogger.SetLogFile(io.Discard, false)

	logEnabled = false
}

// isWalletID check if the client id is a valid wallet hash
//   - clientID is the client id to check
func isWalletID(clientID string) bool {
	if clientID == "" {
		return false
	}

	if !isHash(clientID) {
		return false
	}

	return true

}

const HASH_LENGTH = 32

func isHash(str string) bool {
	bytes, err := hex.DecodeString(str)
	return err == nil && len(bytes) == HASH_LENGTH
}

// getLookupHash retrieve lookup hash with allocation id and path
// Lookup hash is generated by hashing the allocation id and path
//   - allocationID is the allocation id
//   - path is the path
func getLookupHash(allocationID string, path string) string {
	return encryption.Hash(allocationID + ":" + path)
}

// createThumbnail create thumbnail of an image buffer. It supports
//   - png

//   - jpeg
//   - gif
//   - bmp
//   - ccitt
//   - riff
//   - tiff
//   - vector
//   - vp8
//   - vp8l
//   - webp
//     Paramters:
//   - buf is the image buffer which carry the image in bytes
//   - width is the width of the thumbnail
//   - height is the height of the thumbnail
func createThumbnail(buf []byte, width, height int) ([]byte, error) {
	return imageutil.CreateThumbnail(buf, width, height)
}

// makeSCRestAPICall issue a request to the public API of one of the smart contracts
//   - scAddress is the smart contract address
//   - relativePath is the relative path of the endpoint
//   - paramsJson is the parameters in JSON format. It's a key-value map, and added as query parameters to the request.
func makeSCRestAPICall(scAddress, relativePath, paramsJson string) (string, error) {
	var params map[string]string
	err := json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		sdkLogger.Error(fmt.Sprintf("Error parsing JSON: %v", err))
	}
	b, err := screstapi.MakeSCRestAPICall(scAddress, relativePath, params)
	return string(b), err
}

// send Send tokens to a client
//   - toClientID is the client id to send tokens to
//   - tokens is the number of tokens to send
//   - fee is the transaction fee
//   - desc is the description of the transaction
func send(toClientID string, tokens uint64, fee uint64, desc string) (string, error) {
	_, _, _, txn, err := zcncore.Send(toClientID, tokens, desc)
	if err != nil {
		return "", err
	}
	return txn.TransactionOutput, nil
}
