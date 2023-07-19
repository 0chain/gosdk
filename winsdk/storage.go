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

// CreateDir create directory
//
//	return
//		{
//			"error":"",
//			"result":"true",
//		}
//
//export CreateDir
func CreateDir(allocationID, path *C.char) *C.char {
	allocID := C.GoString(allocationID)

	alloc, err := getAllocation(allocID)
	if err != nil {
		return WithJSON(false, err)
	}

	s := C.GoString(path)
	err = alloc.CreateDir(s)

	if err != nil {
		return WithJSON(false, err)
	}

	return WithJSON(true, nil)

}

// Rename rename path
//
//	return
//		{
//			"error":"",
//			"result":"true",
//		}
//
//export Rename
func Rename(allocationID, path, destName *C.char) *C.char {
	allocID := C.GoString(allocationID)

	alloc, err := getAllocation(allocID)
	if err != nil {
		return WithJSON(false, err)
	}

	s := C.GoString(path)
	d := C.GoString(destName)
	err = alloc.RenameObject(s, d)

	if err != nil {
		return WithJSON(false, err)
	}

	return WithJSON(true, nil)

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

	return WithJSON(true, nil)
}

type MultiOperationOption struct {
	OperationType string `json:"operationType,omitempty"`
	RemotePath    string `json:"remotePath,omitempty"`
	DestName      string `json:"destName,omitempty"` // Required only for rename operation
	DestPath      string `json:"destPath,omitempty"` // Required for copy and move operation`
}

type MultiUploadOption struct {
	FilePath      string `json:"filePath,omitempty"`
	FileName      string `json:"fileName,omitempty"`
	RemotePath    string `json:"remotePath,omitempty"`
	ThumbnailPath string `json:"thumbnailPath,omitempty"`
	Encrypt       bool   `json:"encrypt,omitempty"`
	ChunkNumber   int    `json:"chunkNumber,omitempty"`
	IsUpdate      bool   `json:"isUpdate,omitempty"`
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
		return WithJSON(nil, errors.New("AllocationID is required"))
	}
	var options []MultiOperationOption
	err := json.Unmarshal([]byte(jsonMultiOperationOptions), &options)
	if err != nil {
		return WithJSON(nil, err)
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
	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		return WithJSON(nil, err)
	}
	err = allocationObj.DoMultiOperation(operations)
	if err != nil {
		return WithJSON(nil, err)
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
			err = a.DownloadFile(options[i].LocalPath, options[i].RemotePath, false, &StatusCallback{}, false)
		} else {
			err = a.DownloadThumbnail(options[i].LocalPath, options[i].RemotePath, false, &StatusCallback{}, false)
		}
		if err != nil {
			return err
		}
	}
	if options[len(options)-1].DownloadOp == 1 {
		err = a.DownloadFile(options[len(options)-1].LocalPath, options[len(options)-1].RemotePath, false, &StatusCallback{}, true)
	} else {
		err = a.DownloadThumbnail(options[len(options)-1].LocalPath, options[len(options)-1].RemotePath, false, &StatusCallback{}, true)
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
	for idx, option := range options {
		filePaths[idx] = option.Path
		fileNames[idx] = option.Name
		thumbnailPaths[idx] = option.ThumbnailPath
		remotePaths[idx] = option.RemotePath
		chunkNumbers[idx] = option.ChunkNumber
		isUpdates[idx] = option.IsUpdate

	}

	a, err := getAllocation(allocID)
	if err != nil {
		return WithJSON(nil, err)
	}
	statusBar := &StatusCallback{}
	err = a.StartMultiUpload(workdir, filePaths, fileNames, thumbnailPaths, encrypts, chunkNumbers, remotePaths, isUpdates, statusBar)
	if err != nil {
		return WithJSON(nil, err)
	}
	return WithJSON(nil, nil)
}
