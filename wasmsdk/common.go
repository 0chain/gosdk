//go:build js && wasm
// +build js,wasm

package main

import (
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
