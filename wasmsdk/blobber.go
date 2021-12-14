//go:build js && wasm
// +build js,wasm

package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

// Delete delete file from blobbers
func Delete(allocationID, remotePath string, commit bool) (*transaction.Transaction, error) {

	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return nil, RequiredArg("remotePath")
	}

	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		return nil, err
	}

	fileMeta, isFile, err := getFileMeta(allocationObj, remotePath, commit)
	if err != nil {
		return nil, err
	}

	err = allocationObj.DeleteFile(remotePath)
	if err != nil {
		return nil, err
	}

	fmt.Println(remotePath + " deleted")

	txn, err := commitTxn(allocationObj, remotePath, "", "", "", "Delete", fileMeta, commit, isFile)
	if err != nil {
		return nil, err
	}

	return txn, nil
}

// Rename rename a file existing already on dStorage. Only the allocation's owner can rename a file.
func Rename(allocationID, remotePath, destName string, commit bool) (*transaction.Transaction, error) {
	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return nil, RequiredArg("remotePath")
	}

	if len(destName) == 0 {
		return nil, RequiredArg("destName")
	}

	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}

	fileMeta, isFile, err := getFileMeta(allocationObj, remotePath, commit)
	if err != nil {
		return nil, err
	}

	err = allocationObj.RenameObject(remotePath, destName)
	if err != nil {
		PrintError(err.Error())
		return nil, err
	}
	fmt.Println(remotePath + " renamed")

	txn, err := commitTxn(allocationObj, remotePath, destName, "", "", "Rename", fileMeta, commit, isFile)
	if err != nil {
		return nil, err
	}

	return txn, nil
}

// Copy copy file to another folder path on blobbers
func Copy(allocationID, remotePath, destPath string, commit bool) (*transaction.Transaction, error) {

	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return nil, RequiredArg("remotePath")
	}

	if len(destPath) == 0 {
		return nil, RequiredArg("destPath")
	}

	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}

	fileMeta, isFile, err := getFileMeta(allocationObj, remotePath, commit)
	if err != nil {
		return nil, err
	}

	err = allocationObj.CopyObject(remotePath, destPath)
	if err != nil {
		PrintError(err.Error())
		return nil, err
	}

	fmt.Println(remotePath + " copied")

	txn, err := commitTxn(allocationObj, remotePath, destPath, "", "", "Copy", fileMeta, commit, isFile)
	if err != nil {
		return nil, err
	}

	return txn, nil
}

// Move move file to another remote folder path on dStorage. Only the owner of the allocation can copy an object.
func Move(allocationID, remotePath, destPath string, commit bool) (*transaction.Transaction, error) {
	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return nil, RequiredArg("remotePath")
	}

	if len(destPath) == 0 {
		return nil, RequiredArg("destPath")
	}

	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}

	fileMeta, isFile, err := getFileMeta(allocationObj, remotePath, commit)
	if err != nil {
		return nil, err
	}

	err = allocationObj.MoveObject(remotePath, destPath)
	if err != nil {
		PrintError(err.Error())
		return nil, err
	}

	fmt.Println(remotePath + " moved")

	txn, err := commitTxn(allocationObj, remotePath, destPath, "", "", "Move", fileMeta, commit, isFile)
	if err != nil {
		return nil, err
	}

	return txn, nil
}

// Share  generate an authtoken that provides authorization to the holder to the specified file on the remotepath.
func Share(allocationID, remotePath, clientID, encryptionPublicKey string, expiration int, revoke bool, availableAfter int64) (string, error) {

	if len(allocationID) == 0 {
		return "", RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return "", RequiredArg("remotePath")
	}

	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return "", err
	}

	refType := fileref.FILE

	statsMap, err := allocationObj.GetFileStats(remotePath)
	if err != nil {
		PrintError("Error in getting information about the object." + err.Error())
		return "", err
	}
	isFile := false
	for _, v := range statsMap {
		if v != nil {
			isFile = true
			break
		}
	}
	if !isFile {
		refType = fileref.DIRECTORY
	}

	var fileName string
	_, fileName = filepath.Split(remotePath)

	if revoke {
		err := allocationObj.RevokeShare(remotePath, clientID)
		if err != nil {
			PrintError(err.Error())
			return "", err
		}
		fmt.Println("Share revoked for client " + clientID)
		return "", nil
	}

	ref, err := allocationObj.GetAuthTicket(remotePath, fileName, refType, clientID, encryptionPublicKey, int64(expiration), availableAfter)
	if err != nil {
		PrintError(err.Error())
		return "", err
	}
	fmt.Println("Auth token :" + ref)

	return ref, nil

}

// Download download file
func Download(allocationID, remotePath, authTicket, lookupHash string, downloadThumbnailOnly, commit, rxPay, live, delay bool, blocksPerMarker, startBlock, endBlock int) (*DownloadResponse, error) {

	if len(remotePath) == 0 && len(authTicket) == 0 {
		return nil, RequiredArg("remotePath/authTicket")
	}

	if live {

		// m3u8, err := createM3u8Downloader(localpath, remotepath, authticket, allocationID, lookuphash, rxPay, delay)

		// if err != nil {
		// 	PrintError("Error: download files and build playlist: ", err)
		// 	os.Exit(1)
		// }

		// err = m3u8.Start()

		// if err != nil {
		// 	PrintError("Error: download files and build playlist: ", err)
		// 	os.Exit(1)
		// }

		return nil, errors.New("download live is not supported yet")

	}

	if blocksPerMarker == 0 {
		blocksPerMarker = 10
	}

	sdk.SetNumBlockDownloads(blocksPerMarker)
	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg}
	wg.Add(1)
	var errE, err error
	var allocationObj *sdk.Allocation
	fileName := filepath.Base(remotePath)
	localPath := filepath.Join(allocationID, fileName)

	if err != nil {
		PrintError(err)
		return nil, err
	}
	defer sdk.FS.Remove(localPath) //nolint

	if len(authTicket) > 0 {
		at, err := sdk.InitAuthTicket(authTicket).Unmarshall()

		if err != nil {
			PrintError(err)
			return nil, err
		}

		allocationObj, err = sdk.GetAllocationFromAuthTicket(authTicket)
		if err != nil {
			PrintError("Error fetching the allocation", err)
			return nil, err
		}

		var filename string

		if at.RefType == fileref.FILE {
			filename = at.FileName
			lookupHash = at.FilePathHash
		} else if len(lookupHash) > 0 {
			fileMeta, err := allocationObj.GetFileMetaFromAuthTicket(authTicket, lookupHash)
			if err != nil {
				PrintError("Either remotepath or lookuphash is required when using authticket of directory type")
				return nil, err
			}
			filename = fileMeta.Name
		} else if len(remotePath) > 0 {
			lookupHash = fileref.GetReferenceLookup(allocationObj.Tx, remotePath)

			pathnames := strings.Split(remotePath, "/")
			filename = pathnames[len(pathnames)-1]
		} else {
			PrintError("Either remotepath or lookuphash is required when using authticket of directory type")
			return nil, errors.New("Either remotepath or lookuphash is required when using authticket of directory type")
		}

		if downloadThumbnailOnly {
			errE = allocationObj.DownloadThumbnailFromAuthTicket(localPath,
				authTicket, lookupHash, filename, rxPay, statusBar)
		} else {
			if startBlock != 0 || endBlock != 0 {
				errE = allocationObj.DownloadFromAuthTicketByBlocks(
					localPath, authTicket, int64(startBlock), int64(endBlock), blocksPerMarker,
					lookupHash, filename, rxPay, statusBar)
			} else {
				errE = allocationObj.DownloadFromAuthTicket(localPath,
					authTicket, lookupHash, filename, rxPay, statusBar)
			}
		}
	} else if len(remotePath) > 0 {
		if len(allocationID) == 0 {
			return nil, RequiredArg("allocationID")
		}
		allocationObj, err = sdk.GetAllocation(allocationID)

		if err != nil {
			PrintError("Error fetching the allocation", err)
			return nil, err
		}
		if downloadThumbnailOnly {
			errE = allocationObj.DownloadThumbnail(localPath, remotePath, statusBar)
		} else {
			if startBlock != 0 || endBlock != 0 {
				errE = allocationObj.DownloadFileByBlock(localPath, remotePath, int64(startBlock), int64(endBlock), blocksPerMarker, statusBar)
			} else {
				errE = allocationObj.DownloadFile(localPath, remotePath, statusBar)
			}
		}
	}

	if errE == nil {
		wg.Wait()
	} else {
		PrintError("Download failed.", errE.Error())
		return nil, errE
	}
	if !statusBar.success {
		return nil, errors.New("Download failed: unknown error")
	}

	result := &DownloadResponse{
		FileName: fileName,
	}

	fs, _ := sdk.FS.Open(localPath)

	mf, _ := fs.(*common.MemFile)

	result.Url = CreateObjectURL(mf.Buffer.Bytes(), "application/octet-stream")

	if commit {

		txn, err := commitTxn(allocationObj, remotePath, "", authTicket, lookupHash, "Download", nil, true, true)
		if err != nil {
			result.Error = err.Error()

		} else {
			result.Txn = txn
		}

	}

	return result, nil

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
