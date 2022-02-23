package main

import (
	"sync"

	"github.com/0chain/gosdk/core/transaction"
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

func commitFileMetaTxn(allocationObj *sdk.Allocation, remotePath, authTicket, lookupHash string, commandName string, fileMeta *sdk.ConsolidatedFileMeta) (*transaction.Transaction, error) {
	sdkLogger.Info("Commiting changes to blockchain ...")

	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg}
	wg.Add(1)

	err := allocationObj.CommitMetaTransaction(remotePath, commandName, authTicket, lookupHash, fileMeta, statusBar)
	if err != nil {
		PrintError("Commit failed.", err)
		return nil, err
	}

	wg.Wait()

	sdkLogger.Info("Commit Metadata successful")

	txn, err := getLastMetadataCommitTxn()

	if err != nil {
		return nil, err
	}

	return txn, nil
}

func commitFolderMetaTxn(allocationObj *sdk.Allocation, preValue, currValue, commandName string) (*transaction.Transaction, error) {
	sdkLogger.Info("Commiting changes to blockchain ...")
	resp, err := allocationObj.CommitFolderChange(commandName, preValue, currValue)
	if err != nil {
		PrintError("Commit failed.", err)
		return nil, err
	}

	sdkLogger.Info("Commit Metadata successful, Response :", resp)

	txn, err := getLastMetadataCommitTxn()

	if err != nil {
		return nil, err
	}

	return txn, nil
}

func commitTxn(allocationObj *sdk.Allocation, remotePath, newFolderPath, authTicket, lookupHash string, commandName string, fileMeta *sdk.ConsolidatedFileMeta, isFile bool) (*transaction.Transaction, error) {

	if isFile {

		return commitFileMetaTxn(allocationObj, remotePath, authTicket, lookupHash, commandName, fileMeta)
	} else {

		return commitFolderMetaTxn(allocationObj, remotePath, newFolderPath, commandName)
	}

}
