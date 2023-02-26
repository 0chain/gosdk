package zbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

var ErrInvalidAllocation = errors.New("zbox: invalid allocation")

// Allocation - structure for allocation object
type Allocation struct {
	ID           string `json:"id"`
	DataShards   int    `json:"data_shards"`
	ParityShards int    `json:"parity_shards"`
	Size         int64  `json:"size"`
	Expiration   int64  `json:"expiration_date"`
	Name         string `json:"name"`
	Stats        string `json:"stats"`

	blobbers      []*blockchain.StorageNode `json:"-"`
	sdkAllocation *sdk.Allocation           `json:"-"`
}

func ToAllocation(sdkAllocation *sdk.Allocation) *Allocation {
	return &Allocation{
		ID:            sdkAllocation.ID,
		DataShards:    sdkAllocation.DataShards,
		ParityShards:  sdkAllocation.ParityShards,
		Size:          sdkAllocation.Size,
		Expiration:    sdkAllocation.Expiration,
		sdkAllocation: sdkAllocation,
		blobbers:      sdkAllocation.Blobbers,
	}
}

// MinMaxCost - keeps cost for allocation update/creation
type MinMaxCost struct {
	minW float64
	minR float64
	maxW float64
	maxR float64
}

// ListDir - listing files from path
func (a *Allocation) ListDir(path string) (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	listResult, err := a.sdkAllocation.ListDir(path)
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
func (a *Allocation) ListDirFromAuthTicket(authTicket string, lookupHash string) (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	listResult, err := a.sdkAllocation.ListDirFromAuthTicket(authTicket, lookupHash)
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
func (a *Allocation) GetFileMeta(path string) (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	fileMetaData, err := a.sdkAllocation.GetFileMeta(path)
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
func (a *Allocation) GetFileMetaFromAuthTicket(authTicket string, lookupHash string) (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	fileMetaData, err := a.sdkAllocation.GetFileMetaFromAuthTicket(authTicket, lookupHash)
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
func (a *Allocation) DownloadFile(remotePath, localPath string, statusCb StatusCallbackMocked) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.DownloadFile(localPath, remotePath, &StatusCallbackWrapped{Callback: statusCb})
}

// DownloadFileByBlock - start download file from remote path to localpath by blocks number
func (a *Allocation) DownloadFileByBlock(remotePath, localPath string, startBlock, endBlock int64, numBlocks int, statusCb StatusCallbackMocked) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.DownloadFileByBlock(localPath, remotePath, startBlock, endBlock, numBlocks, &StatusCallbackWrapped{Callback: statusCb})
}

// DownloadThumbnail - start download file thumbnail from remote path to localpath
func (a *Allocation) DownloadThumbnail(remotePath, localPath string, statusCb StatusCallbackMocked) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.DownloadThumbnail(localPath, remotePath, &StatusCallbackWrapped{Callback: statusCb})
}

// RepairFile - repair file if it exists in remote path
// ## Inputs
//   - workdir: set a workdir as ~/.zcn on mobile apps
//   - localPath: the local full path of file. eg /usr/local/files/zcn.png
//   - remotePath:
//   - thumbnailPath: the local full path of thumbnail
//   - encrypt: the file should be ecnrypted or not on uploading
//   - statusCb: callback of status
func (a *Allocation) RepairFile(workdir, localPath, remotePath, thumbnailPath string, encrypt bool, statusCb StatusCallbackMocked) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.StartChunkedUpload(workdir, localPath, remotePath, &StatusCallbackWrapped{Callback: statusCb}, true, true, thumbnailPath, encrypt)
}

// UploadFile - upload file/thumbnail from local path to remote path
// ## Inputs
//   - workdir: set a workdir as ~/.zcn on mobile apps
//   - localPath: the local full path of file. eg /usr/local/files/zcn.png
//   - remotePath:
//   - thumbnailPath: the local full path of thumbnail
//   - encrypt: the file should be ecnrypted or not on uploading
//   - statusCb: callback of status
func (a *Allocation) UploadFile(workdir, localPath, remotePath, thumbnailPath string, encrypt bool, statusCb StatusCallbackMocked) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.StartChunkedUpload(workdir, localPath, remotePath, &StatusCallbackWrapped{Callback: statusCb}, false, false, thumbnailPath, encrypt)
}

// UploadFile - update file/thumbnail from local path to remote path
// ## Inputs
//   - workdir: set a workdir as ~/.zcn on mobile apps
//   - localPath: the local full path of file. eg /usr/local/files/zcn.png
//   - remotePath:
//   - thumbnailPath: the local full path of thumbnail
//   - encrypt: the file should be ecnrypted or not on uploading
//   - statusCb: callback of status
func (a *Allocation) UpdateFile(workdir, localPath, remotePath, thumbnailPath string, encrypt bool, statusCb StatusCallbackMocked) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}

	return a.sdkAllocation.StartChunkedUpload(workdir, localPath, remotePath, &StatusCallbackWrapped{Callback: statusCb}, true, false, thumbnailPath, encrypt)
}

// DeleteFile - delete file from remote path
func (a *Allocation) DeleteFile(remotePath string) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.DeleteFile(remotePath)
}

// RenameObject - rename or move file
func (a *Allocation) RenameObject(remotePath string, destName string) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.RenameObject(remotePath, destName)
}

// GetStatistics - get allocation stats
func (a *Allocation) GetAllocationStats() (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	stats := a.sdkAllocation.GetStats()
	retBytes, err := json.Marshal(stats)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

// GetBlobberStats - get blobbers stats
func (a *Allocation) GetBlobberStats() (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	stats := a.sdkAllocation.GetBlobberStats()
	retBytes, err := json.Marshal(stats)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

// RevokeShare - revokes authTicket from refereeClientID
func (a *Allocation) RevokeShare(path string, refereeClientID string) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.RevokeShare(path, refereeClientID)
}

// GetShareAuthToken - get auth ticket from refereeClientID
func (a *Allocation) GetShareAuthToken(path string, filename string, referenceType string, refereeClientID string) (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	return a.sdkAllocation.GetAuthTicketForShare(path, filename, referenceType, refereeClientID)
}

// GetAuthToken - get auth token from refereeClientID
func (a *Allocation) GetAuthToken(path string, filename string, referenceType string, refereeClientID string, refereeEncryptionPublicKey string, expiration int64) (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	availableAfter := time.Now()
	return a.sdkAllocation.GetAuthTicket(path, filename, referenceType, refereeClientID, refereeEncryptionPublicKey, expiration, &availableAfter)
}

// DownloadFromAuthTicket - download file from Auth ticket
func (a *Allocation) DownloadFromAuthTicket(localPath string, authTicket string, remoteLookupHash string, remoteFilename string, status StatusCallbackMocked) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.DownloadFromAuthTicket(localPath, authTicket, remoteLookupHash, remoteFilename, &StatusCallbackWrapped{Callback: status})
}

// DownloadFromAuthTicketByBlocks - download file from Auth ticket by blocks number
func (a *Allocation) DownloadFromAuthTicketByBlocks(localPath string, authTicket string, startBlock, endBlock int64, numBlocks int, remoteLookupHash string, remoteFilename string, status StatusCallbackMocked) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.DownloadFromAuthTicketByBlocks(localPath, authTicket, startBlock, endBlock, numBlocks, remoteLookupHash, remoteFilename, &StatusCallbackWrapped{Callback: status})
}

// DownloadThumbnailFromAuthTicket - downloadThumbnail from Auth ticket
func (a *Allocation) DownloadThumbnailFromAuthTicket(localPath string, authTicket string, remoteLookupHash string, remoteFilename string, status StatusCallbackMocked) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.DownloadThumbnailFromAuthTicket(localPath, authTicket, remoteLookupHash, remoteFilename, &StatusCallbackWrapped{Callback: status})
}

// GetFileStats - get file stats from path
func (a *Allocation) GetFileStats(path string) (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	stats, err := a.sdkAllocation.GetFileStats(path)
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
func (a *Allocation) CancelDownload(remotepath string) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.CancelDownload(remotepath)
}

// CancelUpload - cancel file upload
func (a *Allocation) CancelUpload(localpath string) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.CancelUpload(localpath)
}

// GetDiff - cancel file diff
func (a *Allocation) GetDiff(lastSyncCachePath string, localRootPath string, localFileFilters string, remoteExcludePaths string) (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	var filterArray []string
	err := json.Unmarshal([]byte(localFileFilters), &filterArray)
	if err != nil {
		return "", fmt.Errorf("invalid local file filter JSON. %v", err)
	}
	var exclPathArray []string
	err = json.Unmarshal([]byte(remoteExcludePaths), &exclPathArray)
	if err != nil {
		return "", fmt.Errorf("invalid remote exclude path JSON. %v", err)
	}
	lFdiff, err := a.sdkAllocation.GetAllocationDiff(lastSyncCachePath, localRootPath, filterArray, exclPathArray)
	if err != nil {
		return "", fmt.Errorf("get allocation diff in sdk failed. %v", err)
	}
	retBytes, err := json.Marshal(lFdiff)
	if err != nil {
		return "", fmt.Errorf("failed to convert JSON. %v", err)
	}

	return string(retBytes), nil
}

// SaveRemoteSnapshot - saving remote snapshot
func (a *Allocation) SaveRemoteSnapshot(pathToSave string, remoteExcludePaths string) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	var exclPathArray []string
	err := json.Unmarshal([]byte(remoteExcludePaths), &exclPathArray)
	if err != nil {
		return fmt.Errorf("invalid remote exclude path JSON. %v", err)
	}
	return a.sdkAllocation.SaveRemoteSnapshot(pathToSave, exclPathArray)
}

// StartRepair - start repair files from path
func (a *Allocation) StartRepair(localRootPath, pathToRepair string, statusCb StatusCallbackMocked) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.StartRepair(localRootPath, pathToRepair, &StatusCallbackWrapped{Callback: statusCb})
}

// CancelRepair - cancel repair files from path
func (a *Allocation) CancelRepair() error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.CancelRepair()
}

// CopyObject - copy object from path to dest
func (a *Allocation) CopyObject(path string, destPath string) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.CopyObject(path, destPath)
}

// MoveObject - move object from path to dest
func (a *Allocation) MoveObject(path string, destPath string) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	return a.sdkAllocation.MoveObject(path, destPath)
}

// GetMinWriteRead - getting back cost for allocation
func (a *Allocation) GetMinWriteRead() (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	minW, minR, _ := a.sdkAllocation.GetMinWriteRead()
	maxW, maxR, _ := a.sdkAllocation.GetMaxWriteRead()

	minMaxCost := &MinMaxCost{}
	minMaxCost.maxR = maxR
	minMaxCost.maxW = maxW
	minMaxCost.minR = minR
	minMaxCost.minW = minW

	retBytes, err := json.Marshal(minMaxCost)
	if err != nil {
		return "", fmt.Errorf("failed to convert JSON. %v", err)
	}

	return string(retBytes), nil
}

// GetMaxStorageCost - getting back max cost for allocation
func (a *Allocation) GetMaxStorageCost(size int64) (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	cost, err := a.sdkAllocation.GetMaxStorageCost(size)
	return fmt.Sprintf("%f", cost), err
}

// GetMinStorageCost - getting back min cost for allocation
func (a *Allocation) GetMinStorageCost(size int64) (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	cost, err := a.sdkAllocation.GetMinStorageCost(size)
	return fmt.Sprintf("%f", float64(cost)), err
}

// GetMaxStorageCostWithBlobbers - getting cost for listed blobbers
func (a *Allocation) GetMaxStorageCostWithBlobbers(size int64, blobbersJson string) (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	var selBlobbers *[]*sdk.BlobberAllocation
	err := json.Unmarshal([]byte(blobbersJson), selBlobbers)
	if err != nil {
		return "", err
	}

	cost, err := a.sdkAllocation.GetMaxStorageCostFromBlobbers(size, *selBlobbers)
	return fmt.Sprintf("%f", cost), err
}

// GetFirstSegment - getting the amount of segments in maxSegments for very first playback
func (a *Allocation) GetFirstSegment(localPath, remotePath, tmpPath string, delay, maxSegments int) (string, error) {
	if a == nil || a.sdkAllocation == nil {
		return "", ErrInvalidAllocation
	}
	return CreateStreamingService(a).GetFirstSegment(localPath, remotePath, tmpPath, delay, maxSegments)
}

func (a *Allocation) CreateDir(dirName string) error {
	return a.sdkAllocation.CreateDir(dirName)
}

var currentPlayback StreamingImpl

// GetMinStorageCost - getting back min cost for allocation
func (a *Allocation) PlayStreaming(localPath, remotePath, authTicket, lookupHash, initSegment string, delay int, statusCb StatusCallbackWrapped) error {
	if a == nil || a.sdkAllocation == nil {
		return ErrInvalidAllocation
	}
	currentPlayback = CreateStreamingService(a)
	return currentPlayback.PlayStreaming(localPath, remotePath, authTicket, lookupHash, initSegment, delay, statusCb)
}

func (a *Allocation) StopStreaming() error {
	if currentPlayback == nil {
		return fmt.Errorf("no active playback found")
	}

	return currentPlayback.Stop()
}

func (a *Allocation) GetCurrentManifest() string {

	if currentPlayback == nil {
		return ""
	}

	return currentPlayback.GetCurrentManifest()
}
