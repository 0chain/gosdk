//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"errors"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

func listObjects(allocationId string, remotePath string) (*sdk.ListResult, error) {
	alloc, err := sdk.GetAllocation(allocationId)
	if err != nil {
		return nil, err
	}

	return alloc.ListDir(remotePath)

}

func createDir(allocationID, remotePath string) error {
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

	return allocationObj.CreateDir(remotePath)
}

// lisBlobbersForFile returns details about
func getFileStats(allocationID, remotePath string) ([]*sdk.FileStats, error) {
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

	fileStats, err := allocationObj.GetFileStats(remotePath)
	if err != nil {
		return nil, err
	}

	var output []*sdk.FileStats
	for _, stats := range fileStats {
		output = append(output, stats)
	}

	return output, nil
}

// Delete delete file from blobbers
func Delete(allocationID, remotePath string) (*FileCommandResponse, error) {

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

	err = allocationObj.DeleteFile(remotePath)
	if err != nil {
		return nil, err
	}

	sdkLogger.Info(remotePath + " deleted")

	resp := &FileCommandResponse{
		CommandSuccess: true,
	}

	return resp, nil
}

// Rename rename a file existing already on dStorage. Only the allocation's owner can rename a file.
func Rename(allocationID, remotePath, destName string) (*FileCommandResponse, error) {
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

	err = allocationObj.RenameObject(remotePath, destName)
	if err != nil {
		PrintError(err.Error())
		return nil, err
	}
	sdkLogger.Info(remotePath + " renamed")

	resp := &FileCommandResponse{
		CommandSuccess: true,
	}

	return resp, nil
}

// Copy copy file to another folder path on blobbers
func Copy(allocationID, remotePath, destPath string) (*FileCommandResponse, error) {

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

	err = allocationObj.CopyObject(remotePath, destPath)
	if err != nil {
		PrintError(err.Error())
		return nil, err
	}

	sdkLogger.Info(remotePath + " copied")

	resp := &FileCommandResponse{
		CommandSuccess: true,
	}

	return resp, nil
}

// Move move file to another remote folder path on dStorage. Only the owner of the allocation can copy an object.
func Move(allocationID, remotePath, destPath string) (*FileCommandResponse, error) {
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

	err = allocationObj.MoveObject(remotePath, destPath)
	if err != nil {
		PrintError(err.Error())
		return nil, err
	}

	sdkLogger.Info(remotePath + " moved")

	resp := &FileCommandResponse{
		CommandSuccess: true,
	}

	return resp, nil
}

// Share  generate an authtoken that provides authorization to the holder to the specified file on the remotepath.
func Share(allocationID, remotePath, clientID, encryptionPublicKey string, expiration int, revoke bool, availableAfter string) (string, error) {

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
		sdkLogger.Info("Share revoked for client " + clientID)
		return "", nil
	}

	availableAt := time.Now()

	if len(availableAfter) > 0 {
		aa, err := common.ParseTime(availableAt, availableAfter)
		if err != nil {
			PrintError(err.Error())
			return "", err
		}
		availableAt = *aa
	}

	ref, err := allocationObj.GetAuthTicket(remotePath, fileName, refType, clientID, encryptionPublicKey, int64(expiration), &availableAt)
	if err != nil {
		PrintError(err.Error())
		return "", err
	}
	sdkLogger.Info("Auth token :" + ref)

	return ref, nil

}

// download download file
func download(allocationID, remotePath, authTicket, lookupHash string, downloadThumbnailOnly bool, numBlocks int) (*DownloadCommandResponse, error) {

	if len(remotePath) == 0 && len(authTicket) == 0 {
		return nil, RequiredArg("remotePath/authTicket")
	}

	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg}
	wg.Add(1)

	fileName := strings.Replace(path.Base(remotePath), "/", "-", -1)
	localPath := allocationID + "_" + fileName

	downloader, err := sdk.CreateDownloader(allocationID, localPath, remotePath,
		sdk.WithAuthticket(authTicket, lookupHash),
		sdk.WithOnlyThumbnail(downloadThumbnailOnly),
		sdk.WithBlocks(0, 0, numBlocks))

	if err != nil {
		PrintError(err.Error())
		return nil, err
	}

	defer sys.Files.Remove(localPath) //nolint

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

	fs, _ := sys.Files.Open(localPath)

	mf, _ := fs.(*sys.MemFile)

	resp.Url = CreateObjectURL(mf.Buffer.Bytes(), "application/octet-stream")

	return resp, nil

}

// upload upload file
func upload(allocationID, remotePath string, fileBytes, thumbnailBytes []byte, encrypt, isUpdate, isRepair bool, numBlocks int) (*FileCommandResponse, error) {
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
	}

	if numBlocks < 1 {
		numBlocks = 100
	}

	ChunkedUpload, err := sdk.CreateChunkedUpload("/", allocationObj, fileMeta, fileReader, isUpdate, isRepair,
		sdk.WithThumbnail(thumbnailBytes),
		sdk.WithEncrypt(encrypt),
		sdk.WithStatusCallback(statusBar),
		sdk.WithProgressStorer(&chunkedUploadProgressStorer{list: make(map[string]*sdk.UploadProgress)}),
		sdk.WithChunkNumber(numBlocks))
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

	return resp, nil
}

// download download file blocks
func downloadBlocks(allocationID, remotePath, authTicket, lookupHash string, numBlocks int, startBlockNumber, endBlockNumber int64) (*DownloadCommandResponse, error) {

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
		sdk.WithBlocks(startBlockNumber, endBlockNumber, numBlocks))

	if err != nil {
		PrintError(err.Error())
		return nil, err
	}

	defer sys.Files.Remove(localPath) //nolint

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

	fs, _ := sys.Files.Open(localPath)

	mf, _ := fs.(*sys.MemFile)

	resp.Url = CreateObjectURL(mf.Buffer.Bytes(), "application/octet-stream")

	return resp, nil

}
