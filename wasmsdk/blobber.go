//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"sync"

	"github.com/0chain/gosdk/zboxcore/sdk"
)

// Delete delete file from blobbers
func Delete(allocationID, remotePath string, commit bool) error {

	if len(allocationID) == 0 {
		return RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return RequiredArg("remotePath")
	}

	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		return err
	}

	var fileMeta *sdk.ConsolidatedFileMeta
	isFile := false
	if commit {

		statsMap, err := allocationObj.GetFileStats(remotePath)
		if err != nil {
			return err
		}

		for _, v := range statsMap {
			if v != nil {
				isFile = true
				break
			}
		}

		fileMeta, err = allocationObj.GetFileMeta(remotePath)
		if err != nil {
			return err
		}
	}

	err = allocationObj.DeleteFile(remotePath)
	if err != nil {
		return err
	}

	if commit {
		if isFile {

			fmt.Println("Commiting changes to blockchain ...")

			wg := &sync.WaitGroup{}
			statusBar := &StatusBar{wg: wg}
			wg.Add(1)

			err = allocationObj.CommitMetaTransaction(remotePath, "Delete", "", "", fileMeta, statusBar)
			if err != nil {
				PrintError("Commit failed.", err)
				return err
			}

			wg.Wait()

			fmt.Println("Commit Metadata successful")
		} else {
			fmt.Println("Commiting changes to blockchain ...")
			resp, err := allocationObj.CommitFolderChange("Delete", remotePath, "")
			if err != nil {
				PrintError("Commit failed.", err)
				return err
			}

			fmt.Println("Commit Metadata successful, Response :", resp)
		}
	}

	return nil
}
