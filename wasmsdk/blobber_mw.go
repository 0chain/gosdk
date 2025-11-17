//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"
	"syscall/js"
	"time"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/core/pathutil"
	"github.com/0chain/gosdk/core/sys"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
)

// NOTE: This file provides *MW suffixed* wrappers for several wasm blobber functions
// that accept an additional `key string` parameter (the MultiWalletSupportKey).
// Where the underlying SDK exposes a way to use a per-operation key, the wrapper
// forwards it. Where the SDK does not currently accept a key for that operation,
// the wrapper will either call the existing function (and ignore the key) or
// return an error indicating the key cannot be applied.

// listObjectsMW lists allocation objects but signs the list request using `key` when provided.
func listObjectsMW(allocationID string, remotePath string, offset, pageLimit int, key string) (*sdk.ListResult, error) {
	alloc, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}
	if key != "" {
		return alloc.ListDir(remotePath, sdk.WithListRequestOffset(offset), sdk.WithListRequestPageLimit(pageLimit), sdk.WithListRequestPubKey(key))
	}
	return alloc.ListDir(remotePath, sdk.WithListRequestOffset(offset), sdk.WithListRequestPageLimit(pageLimit))
}

// listObjectsFromAuthTicketMW lists allocation objects using an auth ticket and an optional signing key.
func listObjectsFromAuthTicketMW(allocationID, authTicket, lookupHash string, offset, pageLimit int, key string) (*sdk.ListResult, error) {
	alloc, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}
	if key != "" {
		return alloc.ListDirFromAuthTicket(authTicket, lookupHash, sdk.WithListRequestOffset(offset), sdk.WithListRequestPageLimit(pageLimit), sdk.WithListRequestPubKey(key))
	}
	return alloc.ListDirFromAuthTicket(authTicket, lookupHash, sdk.WithListRequestOffset(offset), sdk.WithListRequestPageLimit(pageLimit))
}

// createDirMW creates a directory and uses the provided key for signing the multi-operation.
func createDirMW(allocationID, remotePath, key string) error {
	if allocationID == "" {
		return errors.New("allocationID required")
	}
	if remotePath == "" {
		return errors.New("remotePath required")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return err
	}
	// Use DoMultiOperation and set the MultiWalletSupportKey on the MultiOperation
	return allocationObj.DoMultiOperation([]sdk.OperationRequest{{
		OperationType: constants.FileOperationCreateDir,
		RemotePath:    remotePath,
	}}, func(mo *sdk.MultiOperation) { mo.MultiWalletSupportKey = key })
}

// DeleteMW deletes a file using the provided MultiWalletSupportKey.
func DeleteMW(allocationID, remotePath, key string) (*FileCommandResponse, error) {
	if allocationID == "" {
		return nil, RequiredArg("allocationID")
	}
	if remotePath == "" {
		return nil, RequiredArg("remotePath")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}

	err = allocationObj.DoMultiOperation([]sdk.OperationRequest{{
		OperationType: constants.FileOperationDelete,
		RemotePath:    remotePath,
	}}, func(mo *sdk.MultiOperation) { mo.MultiWalletSupportKey = key })
	if err != nil {
		return nil, err
	}
	resp := &FileCommandResponse{CommandSuccess: true}
	return resp, nil
}

// RenameMW renames a file using the provided key for signing.
func RenameMW(allocationID, remotePath, destName, key string) (*FileCommandResponse, error) {
	if allocationID == "" {
		return nil, RequiredArg("allocationID")
	}
	if remotePath == "" {
		return nil, RequiredArg("remotePath")
	}
	if destName == "" {
		return nil, RequiredArg("destName")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}

	err = allocationObj.DoMultiOperation([]sdk.OperationRequest{{
		OperationType: constants.FileOperationRename,
		RemotePath:    remotePath,
		DestName:      destName,
	}}, func(mo *sdk.MultiOperation) { mo.MultiWalletSupportKey = key })
	if err != nil {
		return nil, err
	}
	resp := &FileCommandResponse{CommandSuccess: true}
	return resp, nil
}

// CopyMW copies a file and signs using the provided key.
func CopyMW(allocationID, remotePath, destPath, key string) (*FileCommandResponse, error) {
	if allocationID == "" {
		return nil, RequiredArg("allocationID")
	}
	if remotePath == "" {
		return nil, RequiredArg("remotePath")
	}
	if destPath == "" {
		return nil, RequiredArg("destPath")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}

	err = allocationObj.DoMultiOperation([]sdk.OperationRequest{{
		OperationType: constants.FileOperationCopy,
		RemotePath:    remotePath,
		DestPath:      destPath,
	}}, func(mo *sdk.MultiOperation) { mo.MultiWalletSupportKey = key })
	if err != nil {
		return nil, err
	}
	resp := &FileCommandResponse{CommandSuccess: true}
	return resp, nil
}

// MoveMW moves a file and signs using the provided key.
func MoveMW(allocationID, remotePath, destPath, key string) (*FileCommandResponse, error) {
	if allocationID == "" {
		return nil, RequiredArg("allocationID")
	}
	if remotePath == "" {
		return nil, RequiredArg("remotePath")
	}
	if destPath == "" {
		return nil, RequiredArg("destPath")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}

	err = allocationObj.DoMultiOperation([]sdk.OperationRequest{{
		OperationType: constants.FileOperationMove,
		RemotePath:    remotePath,
		DestPath:      destPath,
	}}, func(mo *sdk.MultiOperation) { mo.MultiWalletSupportKey = key })
	if err != nil {
		return nil, err
	}
	resp := &FileCommandResponse{CommandSuccess: true}
	return resp, nil
}

// MultiOperationMW performs multiple operations and signs the whole multi-operation with `key`.
func MultiOperationMW(allocationID string, jsonMultiUploadOptions string, key string) error {
	if allocationID == "" {
		return errors.New("AllocationID is required")
	}
	if jsonMultiUploadOptions == "" {
		return errors.New("operations are empty")
	}
	var options []MultiOperationOption
	if err := json.Unmarshal([]byte(jsonMultiUploadOptions), &options); err != nil {
		return err
	}

	totalOp := len(options)
	operations := make([]sdk.OperationRequest, totalOp)
	for idx, op := range options {
		operations[idx] = sdk.OperationRequest{
			OperationType: op.OperationType,
			RemotePath:    op.RemotePath,
			DestName:      op.DestName,
			DestPath:      op.DestPath,
		}
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return err
	}
	return allocationObj.DoMultiOperation(operations, func(mo *sdk.MultiOperation) { mo.MultiWalletSupportKey = key })
}

// multiDownloadMW - start multi-download operation with per-op key support.
func multiDownloadMW(allocationID, jsonMultiDownloadOptions, authTicket, callbackFuncName, key string) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			PrintError("Recovered in multiDownload Error", r)
		}
	}()
	wg := &sync.WaitGroup{}
	useCallback := false
	if callbackFuncName != "" {
		useCallback = true
	}
	var options []*MultiDownloadOption
	err := json.Unmarshal([]byte(jsonMultiDownloadOptions), &options)
	if err != nil {
		return "", err
	}
	var alloc *sdk.Allocation
	if authTicket == "" {
		alloc, err = getAllocation(allocationID, key)
	} else {
		alloc, err = sdk.GetAllocationFromAuthTicket(authTicket, key)
	}
	if err != nil {
		return "", err
	}
	allStatusBar := make([]*StatusBar, len(options))
	wg.Add(len(options))
	for ind, option := range options {
		fileName := strings.Replace(path.Base(option.RemotePath), "/", "-", -1)
		localPath := allocationID + "_" + fileName
		option.LocalPath = localPath
		statusBar := &StatusBar{wg: wg, totalBytesMap: make(map[string]int)}
		allStatusBar[ind] = statusBar
		if useCallback {
			callback := js.Global().Get(callbackFuncName)
			statusBar.callback = func(totalBytes, completedBytes int, filename, objURL, err string) {
				callback.Invoke(totalBytes, completedBytes, filename, objURL, err)
			}
		}
		var mf sys.File
		if option.DownloadToDisk {
			if option.SuggestedName != "" {
				fileName = option.SuggestedName
			}
			mf, err = jsbridge.NewFileWriter(fileName)
			if err != nil {
				PrintError(err.Error())
				return "", err
			}
		} else {
			statusBar.localPath = localPath
			fs, _ := sys.Files.Open(localPath)
			mf, _ = fs.(*sys.MemFile)
		}

		var downloader sdk.Downloader
		if option.DownloadOp == 1 {
			downloader, err = sdk.CreateDownloader(allocationID, localPath, option.RemotePath,
				sdk.WithAllocation(alloc),
				sdk.WithAuthticket(authTicket, option.RemoteLookupHash),
				sdk.WithOnlyThumbnail(false),
				sdk.WithBlocks(0, 0, option.NumBlocks),
				sdk.WithFileHandler(mf),
				sdk.WithDownloadPubKey(key),
			)
		} else {
			downloader, err = sdk.CreateDownloader(allocationID, localPath, option.RemotePath,
				sdk.WithAllocation(alloc),
				sdk.WithAuthticket(authTicket, option.RemoteLookupHash),
				sdk.WithOnlyThumbnail(true),
				sdk.WithBlocks(0, 0, option.NumBlocks),
				sdk.WithFileHandler(mf),
				sdk.WithDownloadPubKey(key),
			)
		}
		if err != nil {
			PrintError(err.Error())
			return "", err
		}
		defer sys.Files.Remove(option.LocalPath) //nolint
		downloader.Start(statusBar, ind == len(options)-1)
	}
	wg.Wait()
	resp := make([]DownloadCommandResponse, len(options))
	for ind, statusBar := range allStatusBar {
		statusResponse := DownloadCommandResponse{}
		if !statusBar.success {
			statusResponse.CommandSuccess = false
			statusResponse.Error = "Download failed: " + statusBar.err.Error()
		} else {
			statusResponse.CommandSuccess = true
			statusResponse.FileName = options[ind].RemoteFileName
			statusResponse.Url = statusBar.objURL
		}
		resp[ind] = statusResponse
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(respBytes), nil
}

// uploadMW uploads a single file using the provided per-op key when fetching the allocation.
func uploadMW(allocationID, remotePath string, fileBytes, thumbnailBytes []byte, webStreaming, encrypt, isUpdate, isRepair bool, numBlocks int, key string) (*FileCommandResponse, error) {
	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}
	if len(remotePath) == 0 {
		return nil, RequiredArg("remotePath")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}
	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg, totalBytesMap: make(map[string]int)}
	wg.Add(1)
	fileReader := bytes.NewReader(fileBytes)
	localPath := remotePath
	remotePath = zboxutil.RemoteClean(remotePath)
	isabs := zboxutil.IsRemoteAbs(remotePath)
	if !isabs {
		err = errors.New("invalid_path: Path should be valid and absolute")
		return nil, err
	}
	remotePath = zboxutil.GetFullRemotePath(localPath, remotePath)
	_, fileName := pathutil.Split(remotePath)
	mimeType, err := zboxutil.GetFileContentType(path.Ext(fileName), fileReader)
	if err != nil {
		return nil, err
	}
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
	if allocationObj.DataShards > 7 {
		numBlocks = 50
	}
	ChunkedUpload, err := sdk.CreateChunkedUpload(context.TODO(), "/", allocationObj, fileMeta, fileReader, isUpdate, isRepair, webStreaming,
		zboxutil.NewConnectionId(),
		sdk.WithThumbnail(thumbnailBytes),
		sdk.WithEncrypt(encrypt),
		sdk.WithStatusCallback(statusBar),
		sdk.WithChunkNumber(numBlocks),
		sdk.WithUploadPubKey(key),
	)
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
	resp := &FileCommandResponse{CommandSuccess: true}
	return resp, nil
}


// multiUploadMW uploads multiple files in parallel using per-op key when selecting the allocation.
func multiUploadMW(jsonBulkUploadOptions string, key string) ([]BulkUploadResult, error) {
	var options []BulkUploadOption
	err := json.Unmarshal([]byte(jsonBulkUploadOptions), &options)
	if err != nil {
		return nil, err
	}
	n := len(options)
	if n == 0 {
		return nil, errors.New("No files to upload")
	}
	allocationID := options[0].AllocationID
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}
	err = addWebWorkers(allocationObj)
	if err != nil {
		return nil, err
	}
	wait := make(chan BulkUploadResult, 1)
	for _, option := range options {
		go func(o BulkUploadOption) {
			result := BulkUploadResult{RemotePath: o.RemotePath}
			defer func() { wait <- result }()

			ok, err := uploadWithJsFuncsWithKey(o.AllocationID, o.RemotePath, o.ReadChunkFuncName, o.FileSize, o.ThumbnailBytes.Buffer, o.IsWebstreaming, o.Encrypt, o.IsUpdate, o.IsRepair, o.NumBlocks, o.CallbackFuncName, key)

			result.Success = ok
			if err != nil {
				result.Error = err.Error()
				result.Success = false
			}
		}(option)
	}
	results := make([]BulkUploadResult, 0, n)
	for i := 0; i < n; i++ {
		result := <-wait
		results = append(results, result)
	}
	return results, nil
}

// uploadWithJsFuncsWithKey mirrors uploadWithJsFuncs but allows passing per-op key when fetching allocation.
func uploadWithJsFuncsWithKey(allocationID, remotePath string, readChunkFuncName string, fileSize int64, thumbnailBytes []byte, webStreaming, encrypt, isUpdate, isRepair bool, numBlocks int, callbackFuncName string, key string) (bool, error) {
	if len(allocationID) == 0 {
		return false, RequiredArg("allocationID")
	}
	if len(remotePath) == 0 {
		return false, RequiredArg("remotePath")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return false, err
	}
	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg, totalBytesMap: make(map[string]int)}
	if callbackFuncName != "" {
		callback := js.Global().Get(callbackFuncName)
		statusBar.callback = func(totalBytes, completedBytes int, filename, objURL, err string) {
			callback.Invoke(totalBytes, completedBytes, filename, objURL, err)
		}
	}
	wg.Add(1)

	fileReader, err := jsbridge.NewFileReader(readChunkFuncName, fileSize, allocationObj.GetChunkReadSize(encrypt))
	if err != nil {
		return false, err
	}

	localPath := remotePath
	remotePath = zboxutil.RemoteClean(remotePath)
	isabs := zboxutil.IsRemoteAbs(remotePath)
	if !isabs {
		err = errors.New("invalid_path: Path should be valid and absolute")
		return false, err
	}
	remotePath = zboxutil.GetFullRemotePath(localPath, remotePath)

	_, fileName := pathutil.Split(remotePath)

	mimeType, err := zboxutil.GetFileContentType(path.Ext(fileName), fileReader)
	if err != nil {
		return false, err
	}

	fileMeta := sdk.FileMeta{
		Path:       localPath,
		ActualSize: fileSize,
		MimeType:   mimeType,
		RemoteName: fileName,
		RemotePath: remotePath,
	}

	if numBlocks < 1 {
		numBlocks = 100
	}
	if allocationObj.DataShards > 7 {
		numBlocks = 50
	}

	ChunkedUpload, err := sdk.CreateChunkedUpload(context.TODO(), "/", allocationObj, fileMeta, fileReader, isUpdate, isRepair, webStreaming, zboxutil.NewConnectionId(),
		sdk.WithThumbnail(thumbnailBytes),
		sdk.WithEncrypt(encrypt),
		sdk.WithStatusCallback(statusBar),
		sdk.WithChunkNumber(numBlocks),
		sdk.WithUploadPubKey(key),
	)
	if err != nil {
		return false, err
	}

	err = ChunkedUpload.Start()
	if err != nil {
		PrintError("Upload failed.", err)
		return false, err
	}

	wg.Wait()
	if !statusBar.success {
		return false, errors.New("upload failed: unknown")
	}

	return true, nil
}

// cancelUploadMW cancels an upload using the provided key when fetching the allocation.
func cancelUploadMW(allocationID, remotePath, key string) error {
	if allocationID == "" {
		return errors.New("allocationID required")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return err
	}
	return allocationObj.CancelUpload(remotePath)
}

// pauseUploadMW pauses an upload using the provided key when fetching the allocation.
func pauseUploadMW(allocationID, remotePath, key string) error {
	if allocationID == "" {
		return errors.New("allocationID required")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return err
	}
	return allocationObj.PauseUpload(remotePath)
}

// getFileStatsMW returns file stats and allows selecting the wallet via key.
func getFileStatsMW(allocationID, remotePath, key string) ([]*sdk.FileStats, error) {
	if allocationID == "" {
		return nil, RequiredArg("allocationID")
	}
	if remotePath == "" {
		return nil, RequiredArg("remotePath")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}
	m, err := allocationObj.GetFileStats(remotePath)
	if err != nil {
		return nil, err
	}
	var stats []*sdk.FileStats
	for _, v := range m {
		stats = append(stats, v)
	}
	return stats, nil
}

// updateBlobberSettingsMW updates blobber settings. The underlying SDK call does not accept a per-op key
// so the key parameter is ignored here (kept for API symmetry).
func updateBlobberSettingsMW(blobberSettingsJson string, key string) (*transaction.Transaction, error) {
	// Forward to existing implementation (no per-op key supported in SDK wrapper)
	return updateBlobberSettings(blobberSettingsJson)
}

// ShareMW generates an auth ticket and uses the provided key when acquiring the allocation.
func ShareMW(allocationID, remotePath, clientID, encryptionPublicKey string, expiration int, revoke bool, availableAfter string, key string) (string, error) {
	if allocationID == "" {
		return "", RequiredArg("allocationID")
	}
	if remotePath == "" {
		return "", RequiredArg("remotePath")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return "", err
	}
	if revoke {
		if err := allocationObj.RevokeShare(remotePath, clientID); err != nil {
			return "", err
		}
		return "", nil
	}
	availableAt := time.Now()
	ref, err := allocationObj.GetAuthTicket(remotePath, path.Base(remotePath), fileref.DIRECTORY, clientID, encryptionPublicKey, int64(expiration), &availableAt)
	if err != nil {
		return "", err
	}
	return ref, nil
}

// getFileMetaByNameMW wraps getFileMetaByName and supports selecting the allocation via key.
func getFileMetaByNameMW(allocationID, fileNameQuery, key string) ([]*sdk.ConsolidatedFileMetaByName, error) {
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}
	return allocationObj.GetFileMetaByName(fileNameQuery)
}

// getFileMetaByAuthTicketMW wraps getFileMetaByAuthTicket and supports selecting the allocation via key.
func getFileMetaByAuthTicketMW(allocationID, authTicket, lookupHash, key string) (*sdk.ConsolidatedFileMeta, error) {
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}
	return allocationObj.GetFileMetaFromAuthTicket(authTicket, lookupHash)
}

// downloadBlocksMW downloads blocks and supports selecting allocation via key.
func downloadBlocksMW(allocId, remotePath, authTicket, lookupHash, writeChunkFuncName string, startBlock, endBlock int64, key string) ([]byte, error) {
	return downloadBlocksWithKey(allocId, remotePath, authTicket, lookupHash, writeChunkFuncName, startBlock, endBlock, key)
}

// repairAllocationMW repairs allocation using provided key for allocation selection.
func repairAllocationMW(allocationID, callbackFuncName, key string) error {
	return repairAllocationWithKey(allocationID, callbackFuncName, key)
}

// checkAllocStatusMW checks allocation status and supports key for allocation selection.
func checkAllocStatusMW(allocationID, key string) (string, error) {
	alloc, err := getAllocation(allocationID, key)
	if err != nil {
		return "", err
	}
	status, _, err := alloc.CheckAllocStatus(key)
	var statusStr string
	switch status {
	case sdk.Repair:
		statusStr = "repair"
	case sdk.Broken:
		statusStr = "broken"
	default:
		statusStr = "ok"
	}
	return statusStr, err
}

// skipStatusCheckMW sets the check status flag on the allocation, using the provided key.
func skipStatusCheckMW(allocationID string, checkStatus bool, key string) error {
	alloc, err := getAllocation(allocationID, key)
	if err != nil {
		return err
	}
	alloc.SetCheckStatus(checkStatus)
	return nil
}

// terminateWorkersMW cancels workers for an allocation using provided key.
func terminateWorkersMW(allocationID, key string) {
	alloc, err := getAllocation(allocationID, key)
	if err != nil {
		return
	}
	terminateWorkersWithAllocation(alloc)
}

// createWorkersMW creates workers for allocation using provided key.
func createWorkersMW(allocationID, key string) error {
	alloc, err := getAllocation(allocationID, key)
	if err != nil {
		return err
	}
	return addWebWorkers(alloc)
}

// downloadDirectoryMW downloads a directory; uses provided key to select allocation.
func downloadDirectoryMW(allocationID, remotePath, authticket, callbackFuncName, key string) error {
	return downloadDirectoryWithKey(allocationID, remotePath, authticket, callbackFuncName, key)
}

// cancelDownloadDirectoryMW cancels a directory download; key ignored.
func cancelDownloadDirectoryMW(remotePath, key string) {
	cancelDownloadDirectory(remotePath)
}

// cancelDownloadBlocksMW cancels download blocks for allocation using provided key.
func cancelDownloadBlocksMW(allocationID, remotePath string, start, end int64, key string) error {
	alloc, err := getAllocation(allocationID, key)
	if err != nil {
		return err
	}
	return alloc.CancelDownloadBlocks(remotePath, start, end)
}

// setConsensusThresholdMW sets consensus threshold using provided key to find allocation.
func setConsensusThresholdMW(allocationID string, threshold int, key string) error {
	alloc, err := getAllocation(allocationID, key)
	if err != nil {
		return err
	}
	alloc.SetConsensusThreshold(threshold)
	return nil
}

// downloadBlocks downloads file blocks using an optional per-op key when fetching the allocation.
func downloadBlocksWithKey(allocId, remotePath, authTicket, lookupHash, writeChunkFuncName string, startBlock, endBlock int64, key string) ([]byte, error) {

	if len(remotePath) == 0 && len(authTicket) == 0 {
		return nil, RequiredArg("remotePath/authTicket")
	}

	alloc, err := getAllocation(allocId, key)
	if err != nil {
		PrintError("Error fetching the allocation", err)
		return nil, err
	}

	var (
		wg        = &sync.WaitGroup{}
		statusBar = &StatusBar{wg: wg, totalBytesMap: make(map[string]int)}
	)

	if lookupHash == "" {
		lookupHash = getLookupHash(allocId, remotePath)
	}

	var fh sys.File
	if writeChunkFuncName == "" {
		pathHash := encryption.FastHash(fmt.Sprintf("%s:%d:%d", lookupHash, startBlock, endBlock))
		fs, err := sys.Files.Open(pathHash)
		if err != nil {
			return nil, fmt.Errorf("could not open local file: %v", err)
		}

		mf, _ := fs.(*sys.MemFile)
		if mf == nil {
			return nil, fmt.Errorf("invalid memfile")
		}
		fh = mf
		defer sys.Files.Remove(pathHash) //nolint
	} else {
		fh = jsbridge.NewFileCallbackWriter(writeChunkFuncName, lookupHash)
		if fh == nil {
			return nil, fmt.Errorf("could not create file writer, callback function not found")
		}
	}

	wg.Add(1)
	if authTicket != "" {
		err = alloc.DownloadByBlocksToFileHandlerFromAuthTicket(fh, authTicket, lookupHash, startBlock, endBlock, 100, remotePath, false, statusBar, true, sdk.WithFileCallback(
			func() {
				fh.Close() //nolint:errcheck
			},
		))
	} else {
		err = alloc.DownloadByBlocksToFileHandler(
			fh,
			remotePath,
			startBlock,
			endBlock,
			100,
			false,
			statusBar, true, sdk.WithFileCallback(
				func() {
					fh.Close() //nolint:errcheck
				},
			))
	}
	if err != nil {
		return nil, err
	}
	wg.Wait()
	var buf []byte
	if mf, ok := fh.(*sys.MemFile); ok {
		buf = mf.Buffer
	}
	return buf, nil
}

// repairAllocation repairs the allocation using an optional per-op key for allocation lookup.
func repairAllocationWithKey(allocationID, callbackFuncName, key string) error {
	alloc, err := getAllocation(allocationID, key)
	if err != nil {
		return err
	}
	err = addWebWorkers(alloc)
	if err != nil {
		return err
	}
	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg, isRepair: true, totalBytesMap: make(map[string]int)}
	wg.Add(1)
	if callbackFuncName != "" {
		callback := js.Global().Get(callbackFuncName)
		statusBar.callback = func(totalBytes, completedBytes int, filename, objURL, err string) {
			callback.Invoke(totalBytes, completedBytes, filename, objURL, err)
		}
	}
	err = alloc.RepairAlloc(statusBar)
	if err != nil {
		return err
	}
	wg.Wait()
	if statusBar.err != nil {
		fmt.Println("Error in repair allocation: ", statusBar.err)
		return statusBar.err
	}
	status, _, err := alloc.CheckAllocStatus()
	if err != nil {
		return err
	}
	if status == sdk.Repair || status == sdk.Broken {
		fmt.Println("allocation repair failed")
		return errors.New("allocation repair failed")
	}
	return nil
}

// downloadDirectory downloads directory to local file system using fs api, will only work in browsers where fs api is available
// It uses the provided per-op key when selecting the allocation.
func downloadDirectoryWithKey(allocationID, remotePath, authticket, callbackFuncName, key string) error {
	alloc, err := getAllocation(allocationID, key)
	if err != nil {
		return err
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	statusBar := &StatusBar{wg: wg, totalBytesMap: make(map[string]int)}
	if callbackFuncName != "" {
		callback := js.Global().Get(callbackFuncName)
		statusBar.callback = func(totalBytes, completedBytes int, filename, objURL, err string) {
			callback.Invoke(totalBytes, completedBytes, filename, objURL, err)
		}
	}
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	errChan := make(chan error, 1)
	go func() {
		errChan <- alloc.DownloadDirectory(ctx, remotePath, "", authticket, statusBar)
	}()
	downloadDirLock.Lock()
	downloadDirContextMap[remotePath] = cancel
	downloadDirLock.Unlock()
	select {
	case err = <-errChan:
		if err != nil {
			PrintError("Error in download directory: ", err)
		}
		return err
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}

