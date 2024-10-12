//go:build js && wasm
// +build js,wasm

package main

import (
	"errors"
	"syscall/js"

	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

// PrintError is to print to stderr
func PrintError(v ...interface{}) {
	sdkLogger.Error(v...)
}

// PrintInfo is to print to stdout
func PrintInfo(v ...interface{}) {
	sdkLogger.Info(v...)
}

func getFileMeta(allocationObj *sdk.Allocation, remotePath string, commit bool) (*sdk.ConsolidatedFileMeta, bool, error) {
	var fileMeta *sdk.ConsolidatedFileMeta
	isFile := false
	if commit {

		statsMap, err := allocationObj.GetFileStats(remotePath)
		if err != nil {
			return nil, false, err
		}

		for _, v := range statsMap {
			if v != nil {
				isFile = true
				break
			}
		}

		fileMeta, err = allocationObj.GetFileMeta(remotePath)
		if err != nil {
			return nil, false, err
		}
	}

	return fileMeta, isFile, nil
}

type hasher struct {
	md5HashFuncName string
}

func newFileHasher(md5HashFuncName string) sdk.Hasher {
	return &hasher{
		md5HashFuncName: md5HashFuncName,
	}
}

func (h *hasher) GetFileHash() (string, error) {
	md5Callback := js.Global().Get(h.md5HashFuncName)
	result, err := jsbridge.Await(md5Callback.Invoke())
	if len(err) > 0 && !err[0].IsNull() {
		return "", errors.New("file_hash: " + err[0].String())
	}
	return result[0].String(), nil
}

func (h *hasher) WriteToFile(_ []byte) error {
	return nil
}

func (h *hasher) GetFixedMerkleRoot() (string, error) {
	return "", nil
}

func (h *hasher) WriteToFixedMT(_ []byte) error {
	return nil
}

func (h *hasher) GetValidationRoot() (string, error) {
	return "", nil
}

func (h *hasher) WriteToValidationMT(_ []byte) error {
	return nil
}

func (h *hasher) Finalize() error {
	return nil
}

func (h *hasher) GetBlockHash() (string, error) {
	return "", nil
}

func (h *hasher) WriteToBlockHasher(buf []byte) error {
	return nil
}
