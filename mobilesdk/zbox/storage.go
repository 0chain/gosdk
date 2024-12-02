package zbox

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"time"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zboxcore/fileref"
	"github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

type fileResp struct {
	sdk.FileInfo
	Name string `json:"name"`
	Path string `json:"path"`
}

type MultiOperationOption struct {
	OperationType string `json:"operationType,omitempty"`
	RemotePath    string `json:"remotePath,omitempty"`
	DestName      string `json:"destName,omitempty"` // Required only for rename operation
	DestPath      string `json:"destPath,omitempty"` // Required for copy and move operation`
}

type MultiUploadOption struct {
	FilePath       string `json:"filePath,omitempty"`
	FileName       string `json:"fileName,omitempty"`
	RemotePath     string `json:"remotePath,omitempty"`
	ThumbnailPath  string `json:"thumbnailPath,omitempty"`
	Encrypt        bool   `json:"encrypt,omitempty"`
	ChunkNumber    int    `json:"chunkNumber,omitempty"`
	IsUpdate       bool   `json:"isUpdate,omitempty"`
	IsWebstreaming bool   `json:"isWebstreaming,omitempty"`
}

type MultiDownloadOption struct {
	RemotePath       string `json:"remotePath"`
	LocalPath        string `json:"localPath"`
	DownloadOp       int    `json:"downloadOp"`
	RemoteFileName   string `json:"remoteFileName,omitempty"`   //Required only for file download with auth ticket
	RemoteLookupHash string `json:"remoteLookupHash,omitempty"` //Required only for file download with auth ticket
}

type memFileInfo struct {
	size int64
}

func (m memFileInfo) Name() string       { return "memfile" }
func (m memFileInfo) Size() int64        { return m.size }
func (m memFileInfo) Mode() fs.FileMode  { return 0444 }
func (m memFileInfo) ModTime() time.Time { return time.Now() }
func (m memFileInfo) IsDir() bool        { return false }
func (m memFileInfo) Sys() interface{}   { return nil }

type memBuffer struct {
	*bytes.Buffer
	pos int64
}

func (m *memBuffer) Close() error { return nil }

func (m *memBuffer) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case 0: // io.SeekStart
		abs = offset
	case 1: // io.SeekCurrent
		abs = int64(m.pos) + offset
	case 2: // io.SeekEnd
		abs = int64(m.Buffer.Len()) + offset
	default:
		return 0, errors.New("invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("negative position")
	}
	m.pos = abs
	return abs, nil
}

func (m *memBuffer) Stat() (fs.FileInfo, error) {
	return memFileInfo{size: int64(m.Buffer.Len())}, nil
}

func (m *memBuffer) Sync() error {
	return nil
}

func (m *memBuffer) Read(p []byte) (n int, err error) {
	n, err = m.Buffer.Read(p)
	m.pos += int64(n)
	return
}

func (m *memBuffer) Write(p []byte) (n int, err error) {
	n, err = m.Buffer.Write(p)
	m.pos += int64(n)
	return
}

func newMemBuffer() *memBuffer {
	return &memBuffer{
		Buffer: &bytes.Buffer{},
		pos:    0,
	}
}

// MultiOperation - do copy, move, delete and createdir operation together
// ## Inputs
//   - allocationID
//   - jsonMultiOperationOptions: Json Array of MultiOperationOption. eg: "[{"operationType":"move","remotePath":"/README.md","destPath":"/folder1/"},{"operationType":"delete","remotePath":"/t3.txt"}]"
//
// ## Outputs
//   - error
func MultiOperation(allocationID string, jsonMultiOperationOptions string) error {
	if allocationID == "" {
		return errors.New("AllocationID is required")
	}
	var options []MultiOperationOption
	err := json.Unmarshal([]byte(jsonMultiOperationOptions), &options)
	if err != nil {
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
	allocationObj, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return allocationObj.DoMultiOperation(operations)
}

// ListDir - listing files from path
// ## Inputs
//   - allocatonID
//   - remotePath
//
// ## Outputs
//   - the json string of sdk.ListResult
//   - error
func ListDir(allocationID, remotePath string) (string, error) {
	a, err := getAllocation(allocationID)
	if err != nil {
		return "", err
	}

	listResult, err := a.ListDir(remotePath)
	if err != nil {
		return "", err
	}
	retBytes, err := json.Marshal(listResult)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

// ListDirFromAuthTicket - listing files from path with auth ticket
// ## Inputs
//   - allocatonID
//   - authTicket
//   - lookupHash
//
// ## Outputs
//   - the json string of sdk.ListResult
//   - error
func ListDirFromAuthTicket(allocationID, authTicket string, lookupHash string) (string, error) {
	a, err := sdk.GetAllocationFromAuthTicket(authTicket)
	if err != nil {
		return "", err
	}

	at := sdk.InitAuthTicket(authTicket)
	if len(lookupHash) == 0 {
		lookupHash, err = at.GetLookupHash()
		if err != nil {
			return "", err
		}
	}

	listResult, err := a.ListDirFromAuthTicket(authTicket, lookupHash)
	if err != nil {
		return "", err
	}
	retBytes, err := json.Marshal(listResult)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

// GetFileMeta - getting file meta details from file path
// ## Inputs
//   - allocationID
//   - remotePath
//
// ## Outputs
//
//   - the json string of sdk.ConsolidatedFileMeta
//   - error
func GetFileMeta(allocationID, path string) (string, error) {
	a, err := getAllocation(allocationID)
	if err != nil {
		return "", err
	}

	fileMetaData, err := a.GetFileMeta(path)
	if err != nil {
		return "", err
	}
	retBytes, err := json.Marshal(fileMetaData)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

// GetFileMetaFromAuthTicket - getting file meta details from file path and auth ticket
// ## Inputs
//   - allocationID
//   - authTicket
//   - lookupHash
//
// ## Outpus
//   - the json string of sdk.ConsolidatedFileMeta
//   - error
func GetFileMetaFromAuthTicket(allocationID, authTicket string, lookupHash string) (string, error) {
	allocationObj, err := sdk.GetAllocationFromAuthTicket(authTicket)
	if err != nil {
		return "", err
	}

	at := sdk.InitAuthTicket(authTicket)
	if len(lookupHash) == 0 {
		lookupHash, err = at.GetLookupHash()
		if err != nil {
			return "", err
		}
	}

	fileMetaData, err := allocationObj.GetFileMetaFromAuthTicket(authTicket, lookupHash)
	if err != nil {
		return "", err
	}

	retBytes, err := json.Marshal(fileMetaData)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

// DownloadFile - start download file from remote path to localpath
// ## Inputs
//   - allocationID
//   - remotePath
//   - localPath: the full local path of file
//   - statusCb: callback of status
//   - isFinal: is final download request(for example if u want to download 10 files in
//     parallel, the last one should be true)
//
// ## Outputs
//   - error
func DownloadFile(allocationID, remotePath, localPath string, statusCb StatusCallbackMocked, isFinal bool) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return a.DownloadFile(localPath, remotePath, false, &StatusCallbackWrapped{Callback: statusCb}, isFinal)
}

// DownloadFileByBlock - start download file from remote path to localpath by blocks number
// ## Inputs
//
//   - allocationID
//   - remotePath
//   - localPath
//   - startBlock
//   - endBlock
//   - numBlocks
//   - statusCb: callback of status
//   - isFinal: is final download request(for example if u want to download 10 files in
//     parallel, the last one should be true)
//
// ## Outputs
//
//   - error
func DownloadFileByBlock(allocationID, remotePath, localPath string, startBlock, endBlock int64, numBlocks int, statusCb StatusCallbackMocked, isFinal bool) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return a.DownloadFileByBlock(localPath, remotePath, startBlock, endBlock, numBlocks, false, &StatusCallbackWrapped{Callback: statusCb}, isFinal)
}

// DownloadThumbnail - start download file thumbnail from remote path to localpath
// ## Inputs
//   - allocationID
//   - remotePath
//   - localPath
//   - statusCb: callback of status
//   - isFinal: is final download request(for example if u want to download 10 files in
//     parallel, the last one should be true)
//
// ## Outputs
//   - error
func DownloadThumbnail(allocationID, remotePath, localPath string, statusCb StatusCallbackMocked, isFinal bool) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}

	return a.DownloadThumbnail(localPath, remotePath, false, &StatusCallbackWrapped{Callback: statusCb}, isFinal)
}

// MultiDownloadFile - upload files from local path to remote path
// ## Inputs
//   - allocationID
//   - jsonMultiDownloadOptions: Json Array of MultiDownloadOption eg: "[{"remotePath":"/","localPath":"/t2.txt","downloadOp":1}]"
// downloadOp: 1 for file, 2 for thumbnail
// ## Outputs
//   - error

func MultiDownload(allocationID, jsonMultiDownloadOptions string, statusCb StatusCallbackMocked) error {

	var options []MultiDownloadOption
	err := json.Unmarshal([]byte(jsonMultiDownloadOptions), &options)
	if err != nil {
		return err
	}

	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	lastOpInd := len(options) - 1
	for ind, option := range options {
		if option.DownloadOp == 1 {
			err = a.DownloadFile(option.LocalPath, option.RemotePath, false, &StatusCallbackWrapped{Callback: statusCb}, ind == lastOpInd)
		} else {
			err = a.DownloadThumbnail(option.LocalPath, option.RemotePath, false, &StatusCallbackWrapped{Callback: statusCb}, ind == lastOpInd)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// RepairFile - repair file if it exists in remote path
// ## Inputs
//   - allocationID
//   - workdir: set a workdir as ~/.zcn on mobile apps
//   - localPath: the local full path of file. eg /usr/local/files/zcn.png
//   - remotePath:
//   - thumbnailPath: the local full path of thumbnail
//   - encrypt: the file should be ecnrypted or not on uploading
//   - statusCb: callback of status
//
// ## Outputs
//   - error
func RepairFile(allocationID, workdir, localPath, remotePath, thumbnailPath string, encrypt bool, webStreaming bool, statusCb StatusCallbackMocked) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err

	}
	return a.StartChunkedUpload(workdir, localPath, remotePath, &StatusCallbackWrapped{Callback: statusCb}, true, true, thumbnailPath, encrypt, webStreaming)
}

// MultiUploadFile - upload files from local path to remote path
// ## Inputs
//   - allocationID
//   - workdir: set a workdir as ~/.zcn on mobile apps
//   - jsonMultiUploadOpetions: Json Array of MultiOperationOption. eg: "[{"remotePath":"/","filePath":"/t2.txt"},{"remotePath":"/","filePath":"/t3.txt"}]"
//
// ## Outputs
//   - error
func MultiUpload(allocationID string, workdir string, jsonMultiUploadOptions string, statusCb StatusCallbackMocked) error {
	var options []MultiUploadOption
	logger.Logger.Info("multiupload: ", jsonMultiUploadOptions)
	err := json.Unmarshal([]byte(jsonMultiUploadOptions), &options)
	if err != nil {
		logger.Logger.Error("multiupload: ", err)
		return err
	}
	totalUploads := len(options)
	filePaths := make([]string, totalUploads)
	fileNames := make([]string, totalUploads)
	remotePaths := make([]string, totalUploads)
	thumbnailPaths := make([]string, totalUploads)
	encrypts := make([]bool, totalUploads)
	chunkNumbers := make([]int, totalUploads)
	isUpdates := make([]bool, totalUploads)
	isWebstreaming := make([]bool, totalUploads)
	for idx, option := range options {
		filePaths[idx] = option.FilePath
		fileNames[idx] = option.FileName
		thumbnailPaths[idx] = option.ThumbnailPath
		remotePaths[idx] = option.RemotePath
		chunkNumbers[idx] = option.ChunkNumber
		encrypts[idx] = option.Encrypt
		isUpdates[idx] = false
		isWebstreaming[idx] = option.IsWebstreaming
	}

	a, err := getAllocation(allocationID)
	if err != nil {
		logger.Logger.Error("multiupload: ", err)
		return err
	}
	return a.StartMultiUpload(workdir, filePaths, fileNames, thumbnailPaths, encrypts, chunkNumbers, remotePaths, isUpdates, isWebstreaming, &StatusCallbackWrapped{Callback: statusCb})

}

// MultiUpdateFile - update files from local path to remote path
// ## Inputs
//   - allocationID
//   - workdir: set a workdir as ~/.zcn on mobile apps
//   - jsonMultiUploadOpetions: Json Array of MultiOperationOption. eg: "[{"remotePath":"/","filePath":"/t2.txt"},{"remotePath":"/","filePath":"/t3.txt"}]"
//
// ## Outputs
//   - error
func MultiUpdate(allocationID string, workdir string, jsonMultiUploadOptions string, statusCb StatusCallbackMocked) error {
	var options []MultiUploadOption
	err := json.Unmarshal([]byte(jsonMultiUploadOptions), &options)
	totalUploads := len(options)
	filePaths := make([]string, totalUploads)
	fileNames := make([]string, totalUploads)
	remotePaths := make([]string, totalUploads)
	thumbnailPaths := make([]string, totalUploads)
	encrypts := make([]bool, totalUploads)
	chunkNumbers := make([]int, totalUploads)
	isUpdates := make([]bool, totalUploads)
	isWebstreaming := make([]bool, totalUploads)
	for idx, option := range options {
		filePaths[idx] = option.FilePath
		fileNames[idx] = option.FileName
		thumbnailPaths[idx] = option.ThumbnailPath
		remotePaths[idx] = option.RemotePath
		chunkNumbers[idx] = option.ChunkNumber
		encrypts[idx] = option.Encrypt
		isUpdates[idx] = true
		isWebstreaming[idx] = option.IsWebstreaming
	}
	if err != nil {
		return err
	}

	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return a.StartMultiUpload(workdir, filePaths, fileNames, thumbnailPaths, encrypts, chunkNumbers, remotePaths, isUpdates, isWebstreaming, &StatusCallbackWrapped{Callback: statusCb})

}

// UploadFile - upload file/thumbnail from local path to remote path
// ## Inputs
//   - allocationID
//   - workdir: set a workdir as ~/.zcn on mobile apps
//   - localPath: the local full path of file. eg /usr/local/files/zcn.png
//   - remotePath:
//   - thumbnailPath: the local full path of thumbnail
//   - encrypt: the file should be ecnrypted or not on uploading
//   - statusCb: callback of status
//
// ## Outputs
//   - error
func UploadFile(allocationID, workdir, localPath, remotePath, thumbnailPath string, encrypt bool, webStreaming bool, statusCb StatusCallbackMocked) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return a.StartChunkedUpload(workdir, localPath, remotePath, &StatusCallbackWrapped{Callback: statusCb}, false, false, thumbnailPath, encrypt, webStreaming)
}

// UploadFile - update file/thumbnail from local path to remote path
// ## Inputs
//   - workdir: set a workdir as ~/.zcn on mobile apps
//   - localPath: the local full path of file. eg /usr/local/files/zcn.png
//   - remotePath:
//   - thumbnailPath: the local full path of thumbnail
//   - encrypt: the file should be ecnrypted or not on uploading
//   - statusCb: callback of status
//
// ## Ouputs
//   - error
func UpdateFile(allocationID, workdir, localPath, remotePath, thumbnailPath string, encrypt bool, webStreaming bool, statusCb StatusCallbackMocked) error {
	a, err := getAllocation(allocationID)

	if err != nil {
		return err
	}

	return a.StartChunkedUpload(workdir, localPath, remotePath, &StatusCallbackWrapped{Callback: statusCb}, true, false, thumbnailPath, encrypt, webStreaming)
}

// DeleteFile - delete file from remote path
// ## Inputs
//   - allocationID
//   - remotePath
//
// ## Outputs
func DeleteFile(allocationID, remotePath string) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	a.SetCheckStatus(true)
	defer a.SetCheckStatus(false)
	return a.DoMultiOperation([]sdk.OperationRequest{
		{
			OperationType: constants.FileOperationDelete,
			RemotePath:    remotePath,
		},
	})
}

// RenameObject - rename or move file
// ## Inputs
//   - allocationID
//   - remotePath
//   - destName
//
// ## Outputs
//   - error
func RenameObject(allocationID, remotePath string, destName string) error {

	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return a.DoMultiOperation([]sdk.OperationRequest{
		{
			OperationType: constants.FileOperationRename,
			RemotePath:    remotePath,
			DestName:      destName,
		},
	})
}

// GetStatistics - get allocation stats
// ## Inputs
//   - allocationID
//
// ## Outputs
// - the json string of sdk.AllocationStats
// - error
func GetAllocationStats(allocationID string) (string, error) {
	a, err := getAllocation(allocationID)
	if err != nil {
		return "", err
	}
	stats := a.GetStats()
	retBytes, err := json.Marshal(stats)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

// GetBlobberStats - get blobbers stats
// ## Inputs
//   - allocationID
//
// ## Outputs
//   - the json string of map[string]*sdk.BlobberAllocationStats
//   - error
func GetBlobberStats(allocationID string) (string, error) {
	a, err := getAllocation(allocationID)
	if err != nil {
		return "", err
	}

	stats := a.GetBlobberStats()
	retBytes, err := json.Marshal(stats)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

// GetAuthToken - get auth token from refereeClientID
// ## Inputs
//   - allocationID
//   - path
//   - fileName
//   - referenceType: f: file, d: directory
//   - refereeClientID
//   - refereeEncryptionPublicKey
//   - expiration:  seconds in unix time
//   - availableAfter: seconds in unix time
//
// ## Outputs
//   - the json string of *marker.AuthTicket
//   - error
func GetAuthToken(allocationID, path string, filename string, referenceType string, refereeClientID string, refereeEncryptionPublicKey string, expiration int64, availableAfter int64) (string, error) {
	a, err := getAllocation(allocationID)
	if err != nil {
		return "", err
	}
	aa := time.Unix(availableAfter, 0)
	return a.GetAuthTicket(path, filename, referenceType, refereeClientID, refereeEncryptionPublicKey, expiration, &aa)
}

// DownloadFromAuthTicket - download file from Auth ticket
//
//	## Inputs
//	- allocationID
//	- localPath
//	- authTicket
//	- remoteLookupHash
//	- remoteFilename
//	- status: callback of status
func DownloadFromAuthTicket(allocationID, localPath string, authTicket string, remoteLookupHash string, remoteFilename string, status StatusCallbackMocked, isFinal bool) error {
	at, err := sdk.InitAuthTicket(authTicket).Unmarshall()
	if err != nil {
		return err
	}

	a, err := sdk.GetAllocationFromAuthTicket(authTicket)
	if err != nil {
		return err
	}

	if at.RefType == fileref.FILE {
		remoteFilename = at.FileName
		remoteLookupHash = at.FilePathHash
	} else if len(remoteLookupHash) > 0 {
		fileMeta, err := a.GetFileMetaFromAuthTicket(authTicket, remoteLookupHash)
		if err != nil {
			return err
		}
		remoteFilename = fileMeta.Name
	}

	return a.DownloadFromAuthTicket(localPath, authTicket, remoteLookupHash, remoteFilename, false, &StatusCallbackWrapped{Callback: status}, isFinal)
}

// DownloadFromAuthTicketByBlocks - download file from Auth ticket by blocks number
// ## Inputs
//   - allocationID
//   - localPath
//   - authTicket: the base64 string of *marker.AuthTicket
//   - startBlock:
//   - endBlock
//   - numBlocks
//   - remoteLookupHash
//   - remoteFilename
//   - status: callback of status
func DownloadFromAuthTicketByBlocks(allocationID, localPath string, authTicket string, startBlock, endBlock int64, numBlocks int, remoteLookupHash string, remoteFilename string, status StatusCallbackMocked, isFinal bool) error {
	a, err := sdk.GetAllocationFromAuthTicket(authTicket)
	if err != nil {
		return err
	}

	at := sdk.InitAuthTicket(authTicket)
	if len(remoteLookupHash) == 0 {
		remoteLookupHash, err = at.GetLookupHash()
		if err != nil {
			return err
		}
	}

	return a.DownloadFromAuthTicketByBlocks(localPath, authTicket, startBlock, endBlock, numBlocks, remoteLookupHash, remoteFilename, false, &StatusCallbackWrapped{Callback: status}, isFinal)
}

// DownloadThumbnailFromAuthTicket - downloadThumbnail from Auth ticket
// ## Inputs
//   - allocationID
//   - localPath
//   - authTicket: the base64 string of *marker.AuthTicket
//   - remoteLookupHash
//   - remoteFilename
//   - status: callback of status
func DownloadThumbnailFromAuthTicket(allocationID, localPath string, authTicket string, remoteLookupHash string, remoteFilename string, status StatusCallbackMocked, isFinal bool) error {
	at, err := sdk.InitAuthTicket(authTicket).Unmarshall()
	if err != nil {
		return err
	}

	a, err := sdk.GetAllocationFromAuthTicket(authTicket)
	if err != nil {
		return err
	}

	if at.RefType == fileref.FILE {
		remoteFilename = at.FileName
		remoteLookupHash = at.FilePathHash
	} else if len(remoteLookupHash) > 0 {
		fileMeta, err := a.GetFileMetaFromAuthTicket(authTicket, remoteLookupHash)
		if err != nil {
			return err
		}
		remoteFilename = fileMeta.Name
	}

	return a.DownloadThumbnailFromAuthTicket(localPath, authTicket, remoteLookupHash, remoteFilename, false, &StatusCallbackWrapped{Callback: status}, isFinal)
}

func MultiDownloadFromAuthTicket(allocationID, authTicket, jsonMultiDownloadOptions string, status StatusCallbackMocked) error {

	var options []MultiDownloadOption
	err := json.Unmarshal([]byte(jsonMultiDownloadOptions), &options)
	if err != nil {
		return err
	}

	a, err := sdk.GetAllocationFromAuthTicket(authTicket)
	if err != nil {
		return err
	}
	lastOpIndex := len(options) - 1
	for ind, option := range options {
		if option.DownloadOp == 1 {
			err = a.DownloadFromAuthTicket(option.LocalPath, authTicket, option.RemoteLookupHash, option.RemoteFileName, false, &StatusCallbackWrapped{Callback: status}, ind == lastOpIndex)
		} else {
			err = a.DownloadThumbnailFromAuthTicket(option.LocalPath, authTicket, option.RemoteLookupHash, option.RemoteFileName, false, &StatusCallbackWrapped{Callback: status}, ind == lastOpIndex)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// GetFileStats - get file stats from path
// ## Inputs
//   - allocationID
//   - path
//
// ## Outputs
//   - the json string of map[string]*sdk.FileStats
func GetFileStats(allocationID, path string) (string, error) {
	a, err := getAllocation(allocationID)
	if err != nil {
		return "", err
	}
	stats, err := a.GetFileStats(path)
	if err != nil {
		return "", err
	}
	result := make([]*sdk.FileStats, 0)
	for _, v := range stats {
		result = append(result, v)
	}
	retBytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

// CancelDownload - cancel file download
//
//	## Inputs
//	- allocationID
//	- remotePath
func CancelDownload(allocationID, remotepath string) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return a.CancelDownload(remotepath)
}

// CancelUpload - cancel file upload
//
//	## Inputs
//	- allocationID
//	- remotePath
func CancelUpload(allocationID, remotePath string) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return a.CancelUpload(remotePath)
}

// PauseUpload - pause file upload

//	## Inputs
//	- allocationID
//	- remotePath

func PauseUpload(allocationID, remotePath string) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return a.PauseUpload(remotePath)

}

// StartRepair - start repair files from path
//
//	## Inputs
//	- allocationID
//	- localRootPath
//	- pathToRepair
//	- status: callback of status
func StartRepair(allocationID, localRootPath, pathToRepair string, statusCb StatusCallbackMocked) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return a.StartRepair(localRootPath, pathToRepair, &StatusCallbackWrapped{Callback: statusCb})
}

// CancelRepair - cancel repair files from path
//
//	## Inputs
//	- allocationID
func CancelRepair(allocationID string) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return a.CancelRepair()
}

func RepairSize(allocationID, remotePath string) (string, error) {
	alloc, err := getAllocation(allocationID)
	if err != nil {
		return "", err
	}
	size, err := alloc.RepairSize(remotePath)
	if err != nil {
		return "", err
	}
	retBytes, err := json.Marshal(size)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

// CopyObject - copy object from path to dest
// ## Inputs
//   - allocationID
//   - path
//   - destPath
func CopyObject(allocationID, path string, destPath string) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return a.DoMultiOperation([]sdk.OperationRequest{
		{
			OperationType: constants.FileOperationCopy,
			RemotePath:    path,
			DestPath:      destPath,
		},
	})
}

// MoveObject - move object from path to dest
// ## Inputs
//   - allocationID
//   - path
//   - destPath
func MoveObject(allocationID, path string, destPath string) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return a.DoMultiOperation([]sdk.OperationRequest{
		{
			OperationType: constants.FileOperationMove,
			RemotePath:    path,
			DestPath:      destPath,
		},
	})

}

// CreateDir create empty directoy on remote blobbers
//
//	## Inputs
//	- allocationID
//	- dirName
func CreateDir(allocationID, dirName string) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return a.DoMultiOperation([]sdk.OperationRequest{
		{
			OperationType: constants.FileOperationCreateDir,
			RemotePath:    dirName,
		},
	})
}

// RevokeShare revoke authTicket
//
//	## Inputs
//	- allocationID
//	- path
//	- refereeClientID
func RevokeShare(allocationID, path, refereeClientID string) error {
	a, err := getAllocation(allocationID)
	if err != nil {
		return err
	}
	return a.RevokeShare(path, refereeClientID)
}

// GetRemoteFileMap returns the remote
//
//	## Inputs
//	- allocationID
func GetRemoteFileMap(allocationID string) (string, error) {
	a, err := getAllocation(allocationID)
	if err != nil {
		return "", err
	}

	ref, err := a.GetRemoteFileMap(nil, "/")
	if err != nil {
		return "", err
	}

	fileResps := make([]*fileResp, len(ref))
	for path, data := range ref {
		paths := strings.SplitAfter(path, "/")
		var resp = fileResp{
			Name:     paths[len(paths)-1],
			Path:     path,
			FileInfo: data,
		}
		fileResps = append(fileResps, &resp)
	}

	retBytes, err := json.Marshal(fileResps)
	if err != nil {
		return "", err
	}

	return string(retBytes), nil
}

// SetWorkingDir set working dir
//
//	## Inputs
//	- workDir
func SetWorkingDir(workDir string) {
	sdk.Workdir = workDir
}

func SetMultiOpBatchSize(size int) {
	sdk.SetMultiOpBatchSize(size)
}

// SetUploadMode sets upload mode
//
//	## Inputs
//	- mode: 0 for low, 1 for medium, 2 for high
func SetUploadMode(mode int) {
	switch mode {
	case 0:
		sdk.SetUploadMode(sdk.UploadModeLow)
	case 1:
		sdk.SetUploadMode(sdk.UploadModeMedium)
	case 2:
		sdk.SetUploadMode(sdk.UploadModeHigh)
	}
}

// DownloadBlocks downloads specific blocks from an allocation
//   - allocID: allocation ID
//   - remotePath: path to the file in allocation
//   - authTicket: auth ticket for accessing the file
//   - lookupHash: hash for file lookup
//   - startBlock: starting block number
//   - endBlock: ending block number
func DownloadBlocks(allocID, remotePath, authTicket, lookupHash string, startBlock, endBlock int64) ([]byte, error) {
	if len(remotePath) == 0 && len(authTicket) == 0 {
		return nil, errors.New("remotePath/authTicket is required")
	}

	alloc, err := getAllocation(allocID)
	if err != nil {
		return nil, fmt.Errorf("error fetching the allocation: %v", err)
	}

	var (
		wg        = &sync.WaitGroup{}
		statusBar = NewStatusBar(wg)
		buffer    = newMemBuffer()
	)

	if lookupHash == "" {
		lookupHash = getLookupHash(allocID, remotePath)
	}

	wg.Add(1)
	if authTicket != "" {
		err = alloc.DownloadByBlocksToFileHandlerFromAuthTicket(buffer, authTicket, lookupHash, startBlock, endBlock, 100, remotePath, false, statusBar, true)
	} else {
		err = alloc.DownloadByBlocksToFileHandler(
			buffer,
			remotePath,
			startBlock,
			endBlock,
			100,
			false,
			statusBar,
			true)
	}
	if err != nil {
		return nil, err
	}
	wg.Wait()
	return buffer.Bytes(), nil
}

func getLookupHash(allocationID string, path string) string {
	return encryption.Hash(allocationID + ":" + path)
}

// ListObjects list allocation objects from its blobbers
//   - allocationID is the allocation id
//   - remotePath is the remote path of the file
//   - offset is the offset of the list
//   - pageLimit is the limit of the page
func ListObjects(allocationID string, remotePath string, offset, pageLimit int) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in listObjects Error", r)
		}
	}()
	alloc, err := getAllocation(allocationID)
	if err != nil {
		return "", err
	}

	list, err := alloc.ListDir(remotePath, sdk.WithListRequestOffset(offset), sdk.WithListRequestPageLimit(pageLimit))
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(list)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// listObjectsFromAuthTicket list allocation objects from its blobbers using auth ticket
//   - allocationID is the allocation id
//   - authTicket is the auth ticket, provided usually by a non-owner to be able to access a shared source
//   - lookupHash is the lookup hash
//   - offset is the offset of the list
//   - pageLimit is the limit of the page
func ListObjectsFromAuthTicket(allocationID, authTicket, lookupHash string, offset, pageLimit int) (string, error) {
	alloc, err := getAllocation(allocationID)
	if err != nil {
		return "", err
	}
	list, err := alloc.ListDirFromAuthTicket(authTicket, lookupHash, sdk.WithListRequestOffset(offset), sdk.WithListRequestPageLimit(pageLimit))
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(list)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
