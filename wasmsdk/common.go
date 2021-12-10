package main

import (
	"fmt"
	"sync"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

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

func commitTxn(allocationObj *sdk.Allocation, remotePath, newFolderPath, authTicket, lookupHash string, commandName string, fileMeta *sdk.ConsolidatedFileMeta, commit, isFile bool) (*transaction.Transaction, error) {
	if commit {
		if isFile {

			fmt.Println("Commiting changes to blockchain ...")

			wg := &sync.WaitGroup{}
			statusBar := &StatusBar{wg: wg}
			wg.Add(1)

			err := allocationObj.CommitMetaTransaction(remotePath, commandName, authTicket, lookupHash, fileMeta, statusBar)
			if err != nil {
				PrintError("Commit failed.", err)
				return nil, err
			}

			wg.Wait()

			fmt.Println("Commit Metadata successful")
		} else {
			fmt.Println("Commiting changes to blockchain ...")
			resp, err := allocationObj.CommitFolderChange(commandName, remotePath, newFolderPath)
			if err != nil {
				PrintError("Commit failed.", err)
				return nil, err
			}

			fmt.Println("Commit Metadata successful, Response :", resp)
		}

		txn, err := getLastMetadataCommitTxn()

		if err != nil {
			return nil, err
		}

		return txn, nil
	}

	return nil, nil
}
