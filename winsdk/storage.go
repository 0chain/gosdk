package main

/*
#include <stdlib.h>
*/
import (
	"C"
)

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/0chain/gosdk/core/common"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

// GetFileStats get file stats of blobbers
//
//	return
//		{
//			"error":"",
//			"result":"xxx",
//		}
//
//export GetFileStats
func GetFileStats(allocationID, remotePath *C.char) *C.char {
	allocID := C.GoString(allocationID)
	path := C.GoString(remotePath)

	if len(allocID) == 0 {
		return WithJSON(nil, errors.New("allocationID is required"))
	}

	if len(path) == 0 {
		return WithJSON(nil, errors.New("remotePath is required"))
	}

	allocationObj, err := getAllocation(allocID)
	if err != nil {
		return WithJSON(nil, err)
	}

	stats, err := allocationObj.GetFileStats(path)
	if err != nil {
		return WithJSON(nil, err)
	}

	result := make([]*sdk.FileStats, 0, len(stats))

	//convert map[string]*sdk.FileStats to []*sdk.FileStats
	for _, v := range stats {
		result = append(result, v)
	}

	return WithJSON(result, nil)
}

// Delete delete path
//
//	return
//		{
//			"error":"",
//			"result":"true",
//		}
//
//export Delete
func Delete(allocationID, path *C.char) *C.char {
	allocID := C.GoString(allocationID)

	alloc, err := getAllocation(allocID)
	if err != nil {
		return WithJSON(false, err)
	}

	s := C.GoString(path)

	err = alloc.DeleteFile(s)

	if err != nil {
		return WithJSON(false, err)
	}

	log.Info("winsdk: deleted ", s)

	return WithJSON(true, nil)
}

type MultiOperationOption struct {
	OperationType string `json:"OperationType,omitempty"`
	RemotePath    string `json:"RemotePath,omitempty"`
	DestName      string `json:"DestName,omitempty"` // Required only for rename operation
	DestPath      string `json:"DestPath,omitempty"` // Required for copy and move operation`
}

// MultiOperation - do copy, move, delete and createdir operation together
// ## Inputs
//   - allocationID
//   - jsonMultiOperationOptions: Json Array of MultiOperationOption. eg: "[{"operationType":"move","remotePath":"/README.md","destPath":"/folder1/"},{"operationType":"delete","remotePath":"/t3.txt"}]"
//     return
//     {
//     "error":"",
//     "result":"true",
//     }
//
//export MultiOperation
func MultiOperation(_allocationID, _jsonMultiOperationOptions *C.char) *C.char {
	allocationID := C.GoString(_allocationID)
	jsonMultiOperationOptions := C.GoString(_jsonMultiOperationOptions)
	if allocationID == "" {
		return WithJSON(false, errors.New("AllocationID is required"))
	}

	var options []MultiOperationOption
	err := json.Unmarshal([]byte(jsonMultiOperationOptions), &options)
	if err != nil {
		return WithJSON(false, err)
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

		log.Info("multi-operation: index=", idx, " op=", op.OperationType, " remotePath=", op.RemotePath, " destName=", op.DestName, " destPath=", op.DestPath)
	}
	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		return WithJSON(false, err)
	}
	err = allocationObj.DoMultiOperation(operations)
	if err != nil {
		return WithJSON(false, err)
	}
	return WithJSON(true, nil)

}

// GetFileMeta get metadata by path
// ## Inputs
// - allocationID
// - path
// - authTicket
//
// ## Outputs
//
//	return
//		{
//			"error":"",
//			"result":"true",
//		}
//
//export GetFileMeta
func GetFileMeta(allocationID, path, authTicket *C.char) *C.char {
	allocID := C.GoString(allocationID)

	t := C.GoString(authTicket)

	var ticket *marker.AuthTicket

	var alloc *sdk.Allocation
	var err error
	isShared := len(t) > 0
	if isShared {
		alloc, ticket, err = getAllocationWith(t)
	} else {
		alloc, err = getAllocation(allocID)
	}

	if err != nil {
		return WithJSON(nil, err)
	}

	s := C.GoString(path)

	var f *sdk.ConsolidatedFileMeta
	if isShared {
		f, err = alloc.GetFileMetaFromAuthTicket(t, ticket.FilePathHash)
	} else {
		f, err = alloc.GetFileMeta(s)
	}

	if err != nil {
		return WithJSON(nil, err)
	}

	return WithJSON(f, nil)
}

// BulkUpload - upload files from local path to remote path
// ## Inputs
//   - allocationID
//   - files: Json Array of UploadFile
//     return
//     {
//     "error":"",
//     "result":"true",
//     }
//
//export BulkUpload
func BulkUpload(allocationID, files *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: ", r)
		}
	}()
	allocID := C.GoString(allocationID)
	workdir, _ := os.UserHomeDir()
	jsFiles := C.GoString(files)
	var options []UploadFile
	err := json.Unmarshal([]byte(jsFiles), &options)
	if err != nil {
		return WithJSON(nil, err)
	}
	totalUploads := len(options)
	filePaths := make([]string, totalUploads)
	fileNames := make([]string, totalUploads)
	remotePaths := make([]string, totalUploads)
	thumbnailPaths := make([]string, totalUploads)
	chunkNumbers := make([]int, totalUploads)
	encrypts := make([]bool, totalUploads)
	isUpdates := make([]bool, totalUploads)
	isWebstreaming := make([]bool, totalUploads)

	statusBar := NewStatusBar(statusUpload, "")

	for idx, option := range options {
		filePaths[idx] = option.Path
		fileNames[idx] = option.Name
		thumbnailPaths[idx] = option.ThumbnailPath
		remotePaths[idx] = option.RemotePath
		chunkNumbers[idx] = option.ChunkNumber
		isUpdates[idx] = option.IsUpdate
		isWebstreaming[idx] = option.IsWebstreaming
		encrypts[idx] = option.Encrypt
		if option.IsWebstreaming {
			originalLookupHash := getLookupHash(allocID, option.RemotePath+option.Name)
			_, transcodeRemotePath := sdk.GetTranscodeFile(option.RemotePath + option.Name)
			transcodeLookupHash := getLookupHash(allocID, transcodeRemotePath)
			transcodeFiles.Add(originalLookupHash, transcodeLookupHash)
			statusUpload.Add(transcodeLookupHash, &Status{})

		} else {
			statusUpload.Add(getLookupHash(allocID, option.RemotePath+option.Name), &Status{})
		}

	}

	a, err := getAllocation(allocID)
	if err != nil {
		return WithJSON(nil, err)
	}

	err = a.StartMultiUpload(workdir, filePaths, fileNames, thumbnailPaths, encrypts, chunkNumbers, remotePaths, isUpdates, isWebstreaming, statusBar)
	if err != nil {
		return WithJSON(nil, err)
	}
	return WithJSON(nil, nil)
}

// GetUploadStatus - get upload status
// ## Inputs
//   - lookupHash
//
// ## Outputs
//
//	{
//	"error":"",
//	"result":"{'Started':false,'CompletedBytes': 0,Error:”,'Completed':false}",
//	}
//
//export GetUploadStatus
func GetUploadStatus(lookupHash *C.char) *C.char {
	h := C.GoString(lookupHash)

	h2, ok := transcodeFiles.Get(h)

	if ok {
		h = h2
	}

	s, ok := statusUpload.Get(h)

	if !ok {
		s = &Status{}
	}

	return WithJSON(s, nil)
}

// SetNumBlockDownloads - set global the number of blocks on downloading
// ## Inputs
//   - num
//
// ## Outputs
//
//	{
//	"error":"",
//	"result":"",
//	}
//
//export SetNumBlockDownloads
func SetNumBlockDownloads(num int) {
	sdk.SetNumBlockDownloads(num)
}

// DownloadFile - downalod file
// ## Inputs
//   - allocationID
//   - localPath
//   - remotePath
//   - verifyDownload
//   - isFinal
//
// ## Outputs
//
//	{
//	"error":"",
//	"result":"true",
//	}
//
//export DownloadFile
func DownloadFile(allocationID, localPath, remotePath *C.char, verifyDownload, isFinal bool) *C.char {
	allocID := C.GoString(allocationID)

	alloc, err := getAllocation(allocID)
	if err != nil {
		return WithJSON(false, err)
	}

	statusBar := NewStatusBar(statusDownload, "")

	err = alloc.DownloadFile(C.GoString(localPath), C.GoString(remotePath), verifyDownload, statusBar, isFinal)
	if err != nil {
		return WithJSON(false, err)
	}

	return WithJSON(true, nil)
}

// DownloadThumbnail - downalod thumbnial
// ## Inputs
//   - allocationID
//   - localPath
//   - remotePath
//   - verifyDownload
//   - isFinal
//
// ## Outputs
//
//	{
//	"error":"",
//	"result":"true",
//	}
//
//export DownloadThumbnail
func DownloadThumbnail(allocationID, localPath, remotePath *C.char, verifyDownload bool, isFinal bool) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: DownloadThumbnail ", r)
		}
	}()
	allocID := C.GoString(allocationID)

	alloc, err := getAllocation(allocID)
	if err != nil {
		return WithJSON(false, err)
	}

	r := C.GoString(remotePath)

	lookupHash := getLookupHash(allocID, r)
	statusBar := NewStatusBar(statusDownload, lookupHash+":thumbnail")

	err = alloc.DownloadThumbnail(C.GoString(localPath), r, verifyDownload, statusBar, isFinal)
	if err != nil {
		return WithJSON(false, err)
	}

	return WithJSON(true, nil)
}

// DownloadSharedFile - downalod shared file by authTicket
// ## Inputs
//   - localPath
//   - authTicket
//   - verifyDownload
//   - isFinal
//
// ## Outputs
//
//	{
//	"error":"",
//	"result":"{"AllocationID":"xxx","LookupHash":"xxxxxx" }",
//	}
//
//export DownloadSharedFile
func DownloadSharedFile(localPath, authTicket *C.char, verifyDownload bool, isFinal bool) *C.char {
	info := &SharedInfo{}
	t, at, err := getAuthTicket(authTicket)
	if err != nil {
		return WithJSON(info, err)
	}

	info.AllocationID = t.AllocationID
	info.LookupHash = t.FilePathHash

	alloc, err := getAllocation(t.AllocationID)
	if err != nil {
		return WithJSON(info, err)
	}

	statusBar := NewStatusBar(statusDownload, t.FilePathHash)

	err = alloc.DownloadFromAuthTicket(C.GoString(localPath), at, t.FilePathHash, t.FileName, verifyDownload, statusBar, isFinal)
	if err != nil {
		return WithJSON(info, err)
	}

	return WithJSON(info, nil)
}

// DownloadSharedThumbnail - downalod shared thumbnial by authTicket
// ## Inputs
//   - localPath
//   - authTicket
//   - verifyDownload
//   - isFinal
//
// ## Outputs
//
//	{
//	"error":"",
//	"result":"{"AllocationID":"xxx","LookupHash":"xxx" }",
//	}
//
//export DownloadSharedThumbnail
func DownloadSharedThumbnail(localPath, authTicket *C.char, verifyDownload bool, isFinal bool) *C.char {
	info := &SharedInfo{}
	t, at, err := getAuthTicket(authTicket)
	if err != nil {
		return WithJSON(info, err)
	}
	info.AllocationID = t.AllocationID
	info.LookupHash = t.FilePathHash

	alloc, err := getAllocation(t.AllocationID)
	if err != nil {
		return WithJSON(info, err)
	}

	statusBar := NewStatusBar(statusDownload, t.FilePathHash)

	err = alloc.DownloadThumbnailFromAuthTicket(C.GoString(localPath), at, t.FilePathHash, t.FileName, verifyDownload, statusBar, isFinal)
	if err != nil {
		return WithJSON(info, err)
	}

	return WithJSON(info, nil)
}

// DownloadFileBlocks - downalod file blocks
// ## Inputs
//   - allocationID
//   - localPath
//   - remotePath
//   - startBlock
//   - endBlock
//   - numBlocks
//   - verifyDownload
//   - isFinal
//
// ## Outputs
//
//	{
//	"error":"",
//	"result":"true",
//	}
//
//export DownloadFileBlocks
func DownloadFileBlocks(allocationID,
	localPath, remotePath *C.char, startBlock int64, endBlock int64,
	numBlocks int, verifyDownload bool, isFinal bool) *C.char {
	allocID := C.GoString(allocationID)

	alloc, err := getAllocation(allocID)
	if err != nil {
		return WithJSON(false, err)
	}

	r := C.GoString(remotePath)

	lookupHash := getLookupHash(allocID, r)
	statusBar := NewStatusBar(statusDownload, lookupHash+fmt.Sprintf(":%v-%v-%v", startBlock, endBlock, numBlocks))

	err = alloc.DownloadFileByBlock(C.GoString(localPath), r, startBlock, endBlock, numBlocks, verifyDownload, statusBar, isFinal)
	if err != nil {
		return WithJSON(false, err)
	}

	return WithJSON(true, nil)
}

// DownloadSharedFileBlocks - downalod shared file blocks
// ## Inputs
//   - allocationID
//   - localPath
//   - remotePath
//   - startBlock
//   - endBlock
//   - numBlocks
//   - verifyDownload
//   - isFinal
//
// ## Outputs
//
//	{
//	"error":"",
//	"result":"true",
//	}
//
//export DownloadSharedFileBlocks
func DownloadSharedFileBlocks(allocationID,
	localPath, authTicket *C.char, startBlock int64, endBlock int64,
	numBlocks int, verifyDownload bool, isFinal bool) *C.char {

	info := &SharedInfo{}
	t, at, err := getAuthTicket(authTicket)
	if err != nil {
		return WithJSON(info, err)
	}
	info.AllocationID = t.AllocationID
	info.LookupHash = t.FilePathHash

	alloc, err := getAllocation(t.AllocationID)
	if err != nil {
		return WithJSON(info, err)
	}

	statusBar := NewStatusBar(statusDownload, t.FilePathHash+fmt.Sprintf(":%v-%v-%v", startBlock, endBlock, numBlocks))

	err = alloc.DownloadFromAuthTicketByBlocks(C.GoString(localPath), at, startBlock, endBlock, numBlocks, t.FilePathHash, t.FileName, verifyDownload, statusBar, isFinal)
	if err != nil {
		return WithJSON(info, err)
	}

	return WithJSON(info, nil)
}

// GetDownloadStatus - get download status
// ## Inputs
//   - key: lookuphash/lookuphash:thumbnail/lookuphash:startBlock-endBlock-numBlocks/lookuphash:startBlock-endBlock-numBlocks:thumbnail
//
// ## Outputs
//
//	{
//	"error":"",
//	"result":"{'Started':false,'CompletedBytes': 0,Error:”,'Completed':false}",
//	}
//
//export GetDownloadStatus
func GetDownloadStatus(key *C.char, isThumbnail bool) *C.char {

	k := C.GoString(key)
	if isThumbnail {
		k += ":thumbnail"
	}

	s, ok := statusDownload.Get(k)

	if !ok {
		s = &Status{}
	}

	return WithJSON(s, nil)
}

// CreateAuthTicket - create AuthTicket for sharing
// ## Inputs
//   - allocationID
//   - remotePath
//   - refereeClientID
//   - refereePublicEncryptionKey
//   - availableAfter
//   - expirationSeconds
//
// ## Outputs
//
//	{
//	"error":"",
//	"result":"{}",
//	}
//
//export CreateAuthTicket
func CreateAuthTicket(allocationID, remotePath, refereeClientID, refereePublicEncryptionKey, availableAfter *C.char, expirationSeconds int64) *C.char {
	alloc, err := getAllocation(C.GoString(allocationID))
	if err != nil {
		log.Error("win: ", err)
		return WithJSON(nil, err)
	}

	rPath := C.GoString(remotePath)

	fileMeta, err := alloc.GetFileMeta(rPath)
	if err != nil {
		log.Error("win: ", err)
		return WithJSON(nil, err)
	}

	af := time.Now()
	availableAfterString := C.GoString(availableAfter)

	if len(availableAfterString) > 0 {
		aa, err := common.ParseTime(af, availableAfterString)
		if err != nil {
			log.Error("win: ", err)
			return WithJSON(nil, err)
		}
		af = *aa
	}

	at, err := alloc.GetAuthTicket(rPath, fileMeta.Name, fileMeta.Type, C.GoString(refereeClientID), C.GoString(refereePublicEncryptionKey), expirationSeconds, &af)
	if err != nil {
		log.Error("win: ", err)
		return WithJSON(nil, err)
	}

	return WithJSON(at, nil)

}

// DeleteAuthTicket - delete AuthTicket
// ## Inputs
//   - allocationID
//   - remotePath
//   - refereeClientID
//
// ## Outputs
//
//	{
//	"error":"",
//	"result":"true",
//	}
//
//export DeleteAuthTicket
func DeleteAuthTicket(allocationID, remotePath, refereeClientID *C.char) *C.char {
	alloc, err := getAllocation(C.GoString(allocationID))
	if err != nil {
		log.Error("win: ", err)
		return WithJSON(false, err)
	}

	rPath := C.GoString(remotePath)

	err = alloc.RevokeShare(rPath, C.GoString(refereeClientID))
	if err != nil {
		log.Error("win: ", err)
		return WithJSON(false, err)
	}

	return WithJSON(true, nil)

}
