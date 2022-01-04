//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

// Delete delete file from blobbers
func Delete(allocationID, remotePath string, autoCommit bool) (*FileCommandResponse, error) {

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

	fileMeta, isFile, err := getFileMeta(allocationObj, remotePath, autoCommit)
	if err != nil {
		return nil, err
	}

	err = allocationObj.DeleteFile(remotePath)
	if err != nil {
		return nil, err
	}

	fmt.Println(remotePath + " deleted")

	resp := &FileCommandResponse{
		CommandSuccess: true,
	}

	if autoCommit {
		txn, err := commitTxn(allocationObj, remotePath, "", "", "", "Delete", fileMeta, isFile)
		if err != nil {
			resp.Error = err.Error()

			return resp, nil
		}

		resp.CommitSuccess = true
		resp.CommitTxn = txn
	}

	return resp, nil
}

// Rename rename a file existing already on dStorage. Only the allocation's owner can rename a file.
func Rename(allocationID, remotePath, destName string, autoCommit bool) (*FileCommandResponse, error) {
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

	fileMeta, isFile, err := getFileMeta(allocationObj, remotePath, autoCommit)
	if err != nil {
		return nil, err
	}

	err = allocationObj.RenameObject(remotePath, destName)
	if err != nil {
		PrintError(err.Error())
		return nil, err
	}
	fmt.Println(remotePath + " renamed")

	resp := &FileCommandResponse{
		CommandSuccess: true,
	}

	if autoCommit {
		txn, err := commitTxn(allocationObj, remotePath, destName, "", "", "Rename", fileMeta, isFile)
		if err != nil {
			resp.Error = err.Error()

			return resp, nil
		}

		resp.CommitSuccess = true
		resp.CommitTxn = txn
	}

	return resp, nil
}

// Copy copy file to another folder path on blobbers
func Copy(allocationID, remotePath, destPath string, autoCommit bool) (*FileCommandResponse, error) {

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

	fileMeta, isFile, err := getFileMeta(allocationObj, remotePath, autoCommit)
	if err != nil {
		return nil, err
	}

	err = allocationObj.CopyObject(remotePath, destPath)
	if err != nil {
		PrintError(err.Error())
		return nil, err
	}

	fmt.Println(remotePath + " copied")
	resp := &FileCommandResponse{
		CommandSuccess: true,
	}

	if autoCommit {

		txn, err := commitTxn(allocationObj, remotePath, destPath, "", "", "Copy", fileMeta, isFile)
		if err != nil {
			resp.Error = err.Error()

			return resp, nil
		}

		resp.CommitSuccess = true
		resp.CommitTxn = txn
	}

	return resp, nil
}

// Move move file to another remote folder path on dStorage. Only the owner of the allocation can copy an object.
func Move(allocationID, remotePath, destPath string, autoCommit bool) (*FileCommandResponse, error) {
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

	fileMeta, isFile, err := getFileMeta(allocationObj, remotePath, autoCommit)
	if err != nil {
		return nil, err
	}

	err = allocationObj.MoveObject(remotePath, destPath)
	if err != nil {
		PrintError(err.Error())
		return nil, err
	}

	fmt.Println(remotePath + " moved")

	resp := &FileCommandResponse{
		CommandSuccess: true,
	}

	if autoCommit {
		txn, err := commitTxn(allocationObj, remotePath, destPath, "", "", "Move", fileMeta, isFile)
		if err != nil {
			resp.Error = err.Error()

			return resp, nil
		}

		resp.CommitSuccess = true
		resp.CommitTxn = txn
	}

	return resp, nil
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

	ref, err := allocationObj.GetAuthTicket(remotePath, fileName, refType, clientID, encryptionPublicKey, int64(expiration))
	if err != nil {
		PrintError(err.Error())
		return "", err
	}
	fmt.Println("Auth token :" + ref)

	return ref, nil

}

func downloadFile(allocationObj *sdk.Allocation, authTicket string, authTicketObj *marker.AuthTicket, localPath, remotePath, lookupHash string, downloadThumbnailOnly, rxPay bool) (string, error) {
	var blocksPerMarker, startBlock, endBlock int

	if len(remotePath) == 0 && len(authTicket) == 0 {
		return "", RequiredArg("remotePath/authTicket")
	}

	if blocksPerMarker == 0 {
		blocksPerMarker = 10
	}

	sdk.SetNumBlockDownloads(blocksPerMarker)
	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg}
	wg.Add(1)
	var errE, err error

	fileName := filepath.Base(remotePath)

	if err != nil {
		PrintError(err)
		return "", err
	}
	defer sdk.FS.Remove(localPath) //nolint

	if len(authTicket) > 0 {

		if authTicketObj.RefType == fileref.FILE {
			fileName = authTicketObj.FileName
			lookupHash = authTicketObj.FilePathHash
		} else if len(lookupHash) > 0 {
			fileMeta, err := allocationObj.GetFileMetaFromAuthTicket(authTicket, lookupHash)
			if err != nil {
				PrintError("Either remotepath or lookuphash is required when using authticket of directory type")
				return "", err
			}
			fileName = fileMeta.Name
		} else if len(remotePath) > 0 {
			lookupHash = fileref.GetReferenceLookup(allocationObj.Tx, remotePath)

			pathnames := strings.Split(remotePath, "/")
			fileName = pathnames[len(pathnames)-1]
		} else {
			PrintError("Either remotepath or lookuphash is required when using authticket of directory type")
			return "", errors.New("Either remotepath or lookuphash is required when using authticket of directory type")
		}

		if downloadThumbnailOnly {
			errE = allocationObj.DownloadThumbnailFromAuthTicket(localPath,
				authTicket, lookupHash, fileName, rxPay, statusBar)
		} else {
			if startBlock != 0 || endBlock != 0 {
				errE = allocationObj.DownloadFromAuthTicketByBlocks(
					localPath, authTicket, int64(startBlock), int64(endBlock), blocksPerMarker,
					lookupHash, fileName, rxPay, statusBar)
			} else {
				errE = allocationObj.DownloadFromAuthTicket(localPath,
					authTicket, lookupHash, fileName, rxPay, statusBar)
			}
		}
	} else if len(remotePath) > 0 {

		if err != nil {
			PrintError("Error fetching the allocation", err)
			return "", err
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
		return "", errE
	}
	if !statusBar.success {
		return "", errors.New("Download failed: unknown error")
	}

	return fileName, nil
}

// Download download file
func Download(allocationID, remotePath, authTicket, lookupHash string, downloadThumbnailOnly, rxPay, autoCommit bool) (*DownloadCommandResponse, error) {

	if len(remotePath) == 0 && len(authTicket) == 0 {
		return nil, RequiredArg("remotePath/authTicket")
	}

	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg}
	wg.Add(1)

	fileName := strings.Replace(path.Base(remotePath), "/", "-", -1)
	localPath := filepath.Join(allocationID, fileName)

	downloader, err := sdk.CreateDownloader(allocationID, localPath, remotePath,
		sdk.WithAuthticket(authTicket, lookupHash),
		sdk.WithOnlyThumbnail(downloadThumbnailOnly),
		sdk.WithRxPay(rxPay))

	if err != nil {
		PrintError(err.Error())
		return nil, err
	}

	defer sdk.FS.Remove(localPath) //nolint

	err = downloader.Start(statusBar)

	if err == nil {
		wg.Wait()
	} else {
		PrintError("Download failed.", err.Error())
		return nil, err
	}
	if !statusBar.success {
		return nil, errors.New("Download failed: unknown error")
	}

	resp := &DownloadCommandResponse{
		CommandSuccess: true,
		FileName:       fileName,
	}

	fs, _ := sdk.FS.Open(localPath)

	mf, _ := fs.(*common.MemFile)

	resp.Url = CreateObjectURL(mf.Buffer.Bytes(), "application/octet-stream")

	if autoCommit {

		txn, err := commitTxn(downloader.GetAllocation(), remotePath, "", authTicket, lookupHash, "Download", nil, true)
		if err != nil {
			resp.Error = err.Error()

			return resp, nil
		}

		resp.CommitSuccess = true
		resp.CommitTxn = txn
	}

	return resp, nil

}

// Upload upload file
func Upload(allocationID, remotePath string, fileBytes, thumbnailBytes []byte, encrypt, autoCommit bool, attrWhoPaysForReads string, isLiveUpload, isSyncUpload bool, chunkSize int, isUpdate, isRepair bool) (*FileCommandResponse, error) {
	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	if len(remotePath) == 0 {
		return nil, RequiredArg("remotePath")
	}

	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}

	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg}
	wg.Add(1)
	if strings.HasPrefix(remotePath, "/Encrypted") {
		encrypt = true
	}

	var attrs fileref.Attributes
	if len(attrWhoPaysForReads) > 0 {
		var (
			wp common.WhoPays
		)

		if err := wp.Parse(attrWhoPaysForReads); err != nil {
			PrintError(err)
			return nil, err
		}
		attrs.WhoPaysForReads = wp // set given value
	}

	if isLiveUpload {
		return nil, errors.New("live upload is not supported yet")
	} else if isSyncUpload {
		return nil, errors.New("sync upload is not supported yet")
	}

	fileReader := bytes.NewReader(fileBytes)

	mimeType, err := zboxutil.GetFileContentType(fileReader)
	if err != nil {
		return nil, err
	}

	localPath := remotePath

	remotePath = zboxutil.RemoteClean(remotePath)
	isabs := zboxutil.IsRemoteAbs(remotePath)
	if !isabs {
		err = errors.New("invalid_path: Path should be valid and absolute")
		return nil, err
	}
	remotePath = zboxutil.GetFullRemotePath(localPath, remotePath)

	_, fileName := filepath.Split(remotePath)

	fileMeta := sdk.FileMeta{
		Path:       localPath,
		ActualSize: int64(len(fileBytes)),
		MimeType:   mimeType,
		RemoteName: fileName,
		RemotePath: remotePath,
		Attributes: attrs,
	}

	ChunkedUpload, err := sdk.CreateChunkedUpload("/", allocationObj, fileMeta, fileReader, isUpdate, isRepair,
		sdk.WithThumbnail(thumbnailBytes),
		sdk.WithChunkSize(int64(chunkSize)),
		sdk.WithEncrypt(encrypt),
		sdk.WithStatusCallback(statusBar),
		sdk.WithProgressStorer(&chunkedUploadProgressStorer{list: make(map[string]*sdk.UploadProgress)}),
		sdk.WithCreateWriteMarkerLocker(func(file string) sdk.WriteMarkerLocker {
			return &writeMarkerLocker{}
		}))
	if err != nil {
		return nil, err
	}

	err = ChunkedUpload.Start()

	if err != nil {
		PrintError("Upload failed.", err)
		return nil, err
	}
	wg.Wait()
	if !statusBar.success {
		return nil, errors.New("upload failed: unknown")
	}

	resp := &FileCommandResponse{
		CommandSuccess: true,
	}

	if autoCommit {
		txn, err := commitTxn(allocationObj, remotePath, "", "", "", "Upload", nil, true)

		if err != nil {
			resp.Error = err.Error()

			return resp, nil
		}

		resp.CommitSuccess = true
		resp.CommitTxn = txn
	}

	return resp, nil
}

// CommitFileMetaTxn commit file changes to blockchain, and update to blobbers
func CommitFileMetaTxn(allocationID, commandName, remotePath, authTicket, lookupHash string) (*transaction.Transaction, error) {
	fmt.Println("Commiting changes to blockchain ...")

	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}

	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg}
	wg.Add(1)

	err = allocationObj.CommitMetaTransaction(remotePath, commandName, authTicket, lookupHash, nil, statusBar)
	if err != nil {
		PrintError("Commit failed.", err)
		return nil, err
	}

	wg.Wait()

	txn, err := getLastMetadataCommitTxn()

	if err != nil {
		return nil, err
	}

	fmt.Println("Commit Metadata successful")
	return txn, nil
}

// CommitFolderMetaTxn commit folder changes to blockchain
func CommitFolderMetaTxn(allocationID, commandName, preValue, currValue string) (*transaction.Transaction, error) {
	fmt.Println("Commiting changes to blockchain ...")

	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}

	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}

	resp, err := allocationObj.CommitFolderChange(commandName, preValue, currValue)
	if err != nil {
		PrintError("Commit failed.", err)
		return nil, err
	}

	fmt.Println("Commit Metadata successful, Response :", resp)

	txn, err := getLastMetadataCommitTxn()

	if err != nil {
		return nil, err
	}

	return txn, nil

}
