package zbox

import (
	"encoding/json"

	"github.com/0chain/gosdk/zboxcore/blockchain"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

// type allocationImpl interface {
// 	SetConfig(clientJson, dirTreeJson, blobbersJson string, iDataShards, iParityShard int) error
// 	GetDirTree() string
// 	AddDir(path string) error
// 	ListDir(path string) string
// 	GetBlobbers() string
// 	UploadFile(localPath, remotePath string, statusCb StatusCallback) error
// 	UpdateFile(localPath, remotePath string, statusCb StatusCallback) error
// 	RepairFile(localPath, remotePath string, statusCb StatusCallback) error
// 	Commit() error
// 	DownloadFile(remotePath, localPath string, statusCb StatusCallback) error
// 	DownloadCancel()
// 	DeleteFile() error
// 	GetShareAuthToken(remotePath string, clientID string) string
// 	DownloadFileFromShareLink(localPath string, authTokenB64 string, statusCb StatusCallback) error
// 	GetFileStats(remotePath string) string
// }

type StatusCallback interface {
	sdk.StatusCallback
}

type Allocation struct {
	ID           string `json:"id"`
	DataShards   int    `json:"data_shards"`
	ParityShards int    `json:"parity_shards"`
	Size         int64  `json:"size"`
	Expiration   int64  `json:"expiration_date"`

	blobbers      []*blockchain.StorageNode
	sdkAllocation *sdk.Allocation
}

func (a *Allocation) ListDir(path string) (string, error) {
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

func (a *Allocation) DownloadFile(remotePath, localPath string, statusCb StatusCallback) error {
	return a.sdkAllocation.DownloadFile(localPath, remotePath, statusCb)
}

func (a *Allocation) UploadFile(localPath, remotePath string, statusCb StatusCallback) error {
	return a.sdkAllocation.UploadFile(localPath, remotePath, statusCb)
}

func (a *Allocation) DeleteFile(remotePath string) error {
	return a.sdkAllocation.DeleteFile(remotePath)
}

func (a *Allocation) GetStats() (string, error) {
	stats := a.sdkAllocation.GetStats()
	retBytes, err := json.Marshal(stats)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

func (a *Allocation) GetShareAuthToken(path string, refereeClientID string) (string, error) {
	return a.sdkAllocation.GetAuthTicketForShare(path, refereeClientID)
}

func (a *Allocation) DownloadFromAuthTicket(localPath string, authTicket string, status StatusCallback) error {
	return a.sdkAllocation.DownloadFromAuthTicket(localPath, authTicket, status)
}

func (a *Allocation) GetFileStats(path string) (string, error) {
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
