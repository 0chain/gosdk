//go:build mobile
// +build mobile

package sdk

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/zboxcore/client"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/sdk"
	zcn "github.com/0chain/gosdk/zcncore"

	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/0chain/gosdk/mobilesdk/zbox"
	"github.com/0chain/gosdk/zcncore"
)

var nonce = int64(0)

// ChainConfig - blockchain config
type ChainConfig struct {
	ChainID           string   `json:"chain_id,omitempty"`
	PreferredBlobbers []string `json:"preferred_blobbers"`
	BlockWorker       string   `json:"block_worker"`
	SignatureScheme   string   `json:"signature_scheme"`
}

// StorageSDK - storage SDK config
type StorageSDK struct {
	chainconfig *ChainConfig
	client      *client.Client
}

// SetLogFile - setting up log level for core libraries
func SetLogFile(logFile string, verbose bool) {
	zcncore.SetLogFile(logFile, verbose)
	sdk.SetLogFile(logFile, verbose)
}

// SetLogLevel set the log level.
// lvl - 0 disabled; higher number (upto 4) more verbosity
func SetLogLevel(logLevel int) {
	sdk.SetLogLevel(logLevel)
}

func Init(chainConfigJson string) error {
	return zcncore.Init(chainConfigJson)
}

// InitStorageSDK - init storage sdk from config
func InitStorageSDK(clientjson string, configjson string) (*StorageSDK, error) {
	l.Logger.Info("Start InitStorageSDK")
	configObj := &ChainConfig{}
	err := json.Unmarshal([]byte(configjson), configObj)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}
	err = zcncore.InitZCNSDK(configObj.BlockWorker, configObj.SignatureScheme)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}
	l.Logger.Info("InitZCNSDK success")
	l.Logger.Info(configObj.BlockWorker)
	l.Logger.Info(configObj.ChainID)
	l.Logger.Info(configObj.SignatureScheme)
	l.Logger.Info(configObj.PreferredBlobbers)
	err = sdk.InitStorageSDK(clientjson, configObj.BlockWorker, configObj.ChainID, configObj.SignatureScheme, configObj.PreferredBlobbers, 1)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}
	l.Logger.Info("InitStorageSDK success")
	l.Logger.Info("Init successful")
	return &StorageSDK{client: client.GetClient(), chainconfig: configObj}, nil
}

// CreateAllocation - creating new allocation
func (s *StorageSDK) CreateAllocation(name string, datashards, parityshards int, size, expiration int64, lock uint64) (*zbox.Allocation, error) {
	readPrice := sdk.PriceRange{Min: 0, Max: math.MaxInt64}
	writePrice := sdk.PriceRange{Min: 0, Max: math.MaxInt64}
	sdkAllocationID, _, _, err := sdk.CreateAllocation(name, datashards, parityshards, size, expiration, readPrice, writePrice, lock)
	if err != nil {
		return nil, err
	}
	sdkAllocation, err := sdk.GetAllocation(sdkAllocationID)
	if err != nil {
		return nil, err
	}
	return &zbox.Allocation{ID: sdkAllocation.ID, DataShards: sdkAllocation.DataShards, ParityShards: sdkAllocation.ParityShards, Size: sdkAllocation.Size, Expiration: sdkAllocation.Expiration}, nil
}

// CreateAllocationWithBlobbers - creating new allocation with list of blobbers
func (s *StorageSDK) CreateAllocationWithBlobbers(name string, datashards, parityshards int, size, expiration int64, lock uint64, blobbersRaw string) (*zbox.Allocation, error) {
	readPrice := sdk.PriceRange{Min: 0, Max: math.MaxInt64}
	writePrice := sdk.PriceRange{Min: 0, Max: math.MaxInt64}

	options := sdk.CreateAllocationOptions{
		Name:         name,
		DataShards:   datashards,
		ParityShards: parityshards,
		Size:         size,
		Expiry:       expiration,
		Lock:         lock,
		WritePrice:   writePrice,
		ReadPrice:    readPrice,
	}

	blobberUrls := strings.Split(blobbersRaw, "/n")
	if len(blobberUrls) > 0 {
		blobberIds, err := sdk.GetBlobberIds(blobberUrls)
		if err != nil {
			return nil, err
		}

		options.BlobberIds = blobberIds

	}

	sdkAllocationID, _, _, err := sdk.CreateAllocationWith(options)
	if err != nil {
		return nil, err
	}

	sdkAllocation, err := sdk.GetAllocation(sdkAllocationID)
	if err != nil {
		return nil, err
	}

	return zbox.ToAllocation(sdkAllocation), nil
}

// GetAllocation - get allocation from ID
func (s *StorageSDK) GetAllocation(allocationID string) (*zbox.Allocation, error) {
	sdkAllocation, err := sdk.GetAllocation(allocationID)
	if err != nil {
		return nil, err
	}

	stats := sdkAllocation.GetStats()
	retBytes, err := json.Marshal(stats)
	if err != nil {
		return nil, err
	}

	alloc := zbox.ToAllocation(sdkAllocation)
	alloc.Stats = string(retBytes)

	return alloc, nil
}

// GetAllocations - get list of allocations
func (s *StorageSDK) GetAllocations() (string, error) {
	sdkAllocations, err := sdk.GetAllocations()
	if err != nil {
		return "", err
	}
	result := make([]*zbox.Allocation, len(sdkAllocations))
	for i, sdkAllocation := range sdkAllocations {
		allocationObj := &zbox.Allocation{ID: sdkAllocation.ID, DataShards: sdkAllocation.DataShards, ParityShards: sdkAllocation.ParityShards, Size: sdkAllocation.Size, Expiration: sdkAllocation.Expiration}
		result[i] = allocationObj
	}
	retBytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

// GetAllocationFromAuthTicket - get allocation from Auth ticket
func (s *StorageSDK) GetAllocationFromAuthTicket(authTicket string) (*zbox.Allocation, error) {
	sdkAllocation, err := sdk.GetAllocationFromAuthTicket(authTicket)
	if err != nil {
		return nil, err
	}
	return &zbox.Allocation{ID: sdkAllocation.ID, DataShards: sdkAllocation.DataShards, ParityShards: sdkAllocation.ParityShards, Size: sdkAllocation.Size, Expiration: sdkAllocation.Expiration}, nil
}

// GetAllocationStats - get allocation stats by allocation ID
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

// FinalizeAllocation - finalize allocation
func (s *StorageSDK) FinalizeAllocation(allocationID string) (string, error) {
	hash, _, err := sdk.FinalizeAllocation(allocationID)
	return hash, err
}

// CancelAllocation - cancel allocation by ID
func (s *StorageSDK) CancelAllocation(allocationID string) (string, error) {
	hash, _, err := sdk.CancelAllocation(allocationID)
	return hash, err
}

//GetReadPoolInfo is to get information about the read pool for the allocation
func (s *StorageSDK) GetReadPoolInfo(clientID string) (string, error) {
	readPool, err := sdk.GetReadPoolInfo(clientID)
	if err != nil {
		return "", err
	}

	retBytes, err := json.Marshal(readPool)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

// WRITE POOL METHODS

func (s *StorageSDK) WritePoolLock(durInSeconds int64, tokens, fee float64, allocID string) error {
	_, _, err := sdk.WritePoolLock(
		allocID,
		zcn.ConvertTokenToSAS(tokens),
		zcn.ConvertTokenToSAS(fee))
	return err
}

// GetVersion getting current version for gomobile lib
func (s *StorageSDK) GetVersion() string {
	return version.VERSIONSTR
}

// UpdateAllocation with new expiry and size
func (s *StorageSDK) UpdateAllocation(name string, size, expiry int64, allocationID string, lock uint64) (hash string, err error) {
	hash, _, err = sdk.UpdateAllocation(name, size, expiry, allocationID, lock, false, true, "", "")
	return hash, err
}

// GetBlobbersList get list of blobbers in string
func (s *StorageSDK) GetBlobbersList() (string, error) {
	blobbs, err := sdk.GetBlobbers()
	if err != nil {
		return "", err
	}
	retBytes, err := json.Marshal(blobbs)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

//GetReadPoolInfo is to get information about the read pool for the allocation
func RegisterToMiners(clientId, pubKey string, callback zcn.WalletCallback) error {
	wallet := zcncrypto.Wallet{ClientID: clientId, ClientKey: pubKey}
	return zcncore.RegisterToMiners(&wallet, callback)
}

// GetAllocations return back list of allocations for the wallet
// Extracted from main method, bcz of class fields
func GetAllocations() (string, error) {
	allocs, err := sdk.GetAllocations()
	if err != nil {
		return "", err
	}
	retBytes, err := json.Marshal(allocs)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

func (s *StorageSDK) RedeemFreeStorage(ticket string) (string, error) {
	recipientPublicKey, marker, lock, err := decodeTicket(ticket)
	if err != nil {
		return "", err
	}

	hash, _, err := sdk.CreateFreeAllocationFor(recipientPublicKey, marker, lock)
	return hash, err
}

func decodeTicket(ticket string) (string, string, uint64, error) {
	decoded, err := base64.StdEncoding.DecodeString(ticket)
	if err != nil {
		return "", "", 0, err
	}

	input := make(map[string]interface{})
	if err = json.Unmarshal(decoded, &input); err != nil {
		return "", "", 0, err
	}

	str := fmt.Sprintf("%v", input["marker"])
	decodedMarker, _ := base64.StdEncoding.DecodeString(str)
	markerInput := make(map[string]interface{})
	if err = json.Unmarshal(decodedMarker, &markerInput); err != nil {
		return "", "", 0, err
	}

	recipientPublicKey, ok := input["recipient_public_key"].(string)
	if !ok {
		return "", "", 0, fmt.Errorf("recipient_public_key is required")
	}

	lock := markerInput["free_tokens"]
	markerStr, _ := json.Marshal(markerInput)

	s, _ := strconv.ParseFloat(string(fmt.Sprintf("%v", lock)), 64)
	return string(recipientPublicKey), string(markerStr), zcn.ConvertTokenToSAS(s), nil
}
