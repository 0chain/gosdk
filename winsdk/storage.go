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
	"os"

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

// GetAllocation get allocation info
//
//	return
//		{
//			"error":"",
//			"result":"{}",
//		}
//
//export GetAllocation
func GetAllocation(allocationID *C.char) *C.char {
	allocID := C.GoString(allocationID)
	return WithJSON(getAllocation(allocID))
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
//
//	return
//		{
//			"error":"",
//			"result":"true",
//		}
//
//export GetFileMeta
func GetFileMeta(allocationID, path *C.char) *C.char {
	allocID := C.GoString(allocationID)

	alloc, err := getAllocation(allocID)
	if err != nil {
		return WithJSON(nil, err)
	}

	s := C.GoString(path)

	f, err := alloc.GetFileMeta(s)

	if err != nil {
		return WithJSON(nil, err)
	}

	return WithJSON(f, nil)
}

type MultiDownloadOption struct {
	RemotePath       string `json:"remotePath"`
	LocalPath        string `json:"localPath"`
	DownloadOp       int    `json:"downloadOp"`
	RemoteFileName   string `json:"remoteFileName,omitempty"`   //Required only for file download with auth ticket
	RemoteLookupHash string `json:"remoteLookupHash,omitempty"` //Required only for file download with auth ticket
}

// MultiDownloadFile - upload files from local path to remote path
// ## Inputs
//   - allocationID
//   - jsonMultiDownloadOptions: Json Array of MultiDownloadOption eg: "[{"remotePath":"/","localPath":"/t2.txt","downloadOp":1}]"
//
// downloadOp: 1 for file, 2 for thumbnail
// ## Outputs
//   - error
//
// export MultiDownload
func MultiDownload(_allocationID, _jsonMultiDownloadOptions *C.char) error {
	allocationID := C.GoString(_allocationID)
	jsonMultiUploadOptions := C.GoString(_jsonMultiDownloadOptions)
	var options []MultiDownloadOption
	err := json.Unmarshal([]byte(jsonMultiUploadOptions), &options)
	if err != nil {
		return err
	}

	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}

	for i := 0; i < len(options)-1; i++ {
		if options[i].DownloadOp == 1 {
			err = a.DownloadFile(options[i].LocalPath, options[i].RemotePath, false, NewStatusBar(statusDownload), false)
		} else {
			err = a.DownloadThumbnail(options[i].LocalPath, options[i].RemotePath, false, NewStatusBar(statusDownload), false)
		}
		if err != nil {
			return err
		}
	}
	if options[len(options)-1].DownloadOp == 1 {
		err = a.DownloadFile(options[len(options)-1].LocalPath, options[len(options)-1].RemotePath, false, NewStatusBar(statusDownload), true)
	} else {
		err = a.DownloadThumbnail(options[len(options)-1].LocalPath, options[len(options)-1].RemotePath, false, NewStatusBar(statusDownload), true)
	}

	return err
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
		statusUpload.Add(getLookupHash(allocID, option.RemotePath+option.Name), &Status{})
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

	s, ok := statusUpload.Get(C.GoString(lookupHash))

	if !ok {
		s = &Status{}
	}

	return WithJSON(s, nil)
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

// GetDownloadStatus - get download status
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
//export GetDownloadStatus
func GetDownloadStatus(lookupHash *C.char, isThumbnail bool) *C.char {

	key := C.GoString(lookupHash)
	if isThumbnail {
		key += ":thumbnail"
	}

	s, ok := statusDownload.Get(key)

	if !ok {
		s = &Status{}
	}

	return WithJSON(s, nil)
}
