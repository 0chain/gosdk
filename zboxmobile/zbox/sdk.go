package zbox

import (
	"encoding/json"

	"0chain.net/clientsdk/zboxcore/client"
	"0chain.net/clientsdk/zboxcore/sdk"
	"0chain.net/clientsdk/zcncore"
)

type ChainConfig struct {
	ChainID         string   `json:"chain_id,omitempty"`
	Miners          []string `json:"miners"`
	Sharders        []string `json:"sharders"`
	SignatureScheme string   `json:"signaturescheme"`
}

type StorageSDK struct {
	chainconfig *ChainConfig
	client      *client.Client
}

func SetLogFile(logFile string, verbose bool) {
	zcncore.SetLogFile(logFile, verbose)
	sdk.SetLogFile(logFile, verbose)
}

func InitStorageSDK(clientjson string, configjson string) (*StorageSDK, error) {
	configObj := &ChainConfig{}
	err := json.Unmarshal([]byte(configjson), configObj)
	if err != nil {
		return nil, err
	}
	err = zcncore.InitZCNSDK(configObj.Miners, configObj.Sharders, configObj.SignatureScheme)
	if err != nil {
		return nil, err
	}
	err = sdk.InitStorageSDK(clientjson, configObj.Miners, configObj.Sharders, configObj.ChainID, configObj.SignatureScheme)
	if err != nil {
		return nil, err
	}
	return &StorageSDK{client: client.GetClient(), chainconfig: configObj}, nil
}

func (s *StorageSDK) CreateAllocation(datashards int, parityshards int, size int64, expiration int64) (*Allocation, error) {
	sdkAllocationID, err := sdk.CreateAllocation(datashards, parityshards, size, expiration)
	if err != nil {
		return nil, err
	}
	sdkAllocation, err := sdk.GetAllocation(sdkAllocationID)
	if err != nil {
		return nil, err
	}
	return &Allocation{ID: sdkAllocation.ID, DataShards: sdkAllocation.DataShards, ParityShards: sdkAllocation.ParityShards, Size: sdkAllocation.Size, Expiration: sdkAllocation.Expiration, blobbers: sdkAllocation.Blobbers, sdkAllocation: sdkAllocation}, nil
}

func (s *StorageSDK) GetAllocation(allocationID string) (*Allocation, error) {
	sdkAllocation, err := sdk.GetAllocation(allocationID)
	if err != nil {
		return nil, err
	}
	return &Allocation{ID: sdkAllocation.ID, DataShards: sdkAllocation.DataShards, ParityShards: sdkAllocation.ParityShards, Size: sdkAllocation.Size, Expiration: sdkAllocation.Expiration, blobbers: sdkAllocation.Blobbers, sdkAllocation: sdkAllocation}, nil
}

func (s *StorageSDK) GetAllocations() (string, error) {
	sdkAllocations, err := sdk.GetAllocations()
	if err != nil {
		return "", err
	}
	result := make([]*Allocation, len(sdkAllocations))
	for i, sdkAllocation := range sdkAllocations {
		allocationObj := &Allocation{ID: sdkAllocation.ID, DataShards: sdkAllocation.DataShards, ParityShards: sdkAllocation.ParityShards, Size: sdkAllocation.Size, Expiration: sdkAllocation.Expiration, blobbers: sdkAllocation.Blobbers, sdkAllocation: sdkAllocation}
		result[i] = allocationObj
	}
	retBytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

func (s *StorageSDK) GetAllocationFromAuthTicket(authTicket string) (*Allocation, error) {
	sdkAllocation, err := sdk.GetAllocationFromAuthTicket(authTicket)
	if err != nil {
		return nil, err
	}
	return &Allocation{ID: sdkAllocation.ID, DataShards: sdkAllocation.DataShards, ParityShards: sdkAllocation.ParityShards, Size: sdkAllocation.Size, Expiration: sdkAllocation.Expiration, blobbers: sdkAllocation.Blobbers, sdkAllocation: sdkAllocation}, nil
}

func (s *StorageSDK) GetAllocationStats(allocationID string) (string, error) {
	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		return "", err
	}
	stats := allocationObj.GetStats()
	retBytes, err := json.Marshal(stats)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}
