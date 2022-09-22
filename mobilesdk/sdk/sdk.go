package sdk

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/zboxcore/client"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zboxcore/zboxutil"
	zcn "github.com/0chain/gosdk/zcncore"

	"github.com/0chain/gosdk/mobilesdk/zbox"
	"github.com/0chain/gosdk/mobilesdk/zcncore"
	"github.com/0chain/gosdk/mobilesdk/zcncoremobile"
	"github.com/0chain/gosdk/mobilesdk/zcncrypto"
	"go.uber.org/zap"
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
		uint64(zcncoremobile.ConvertToValue(tokens)),
		uint64(zcncoremobile.ConvertToValue(fee)))
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
	input, err, lock := decodeTicket(ticket)
	if err != nil {
		return "", err
	}

	blobbers, err := getFreeAllocationBlobbers(input)
	if err != nil {
		return "", err
	}
	if len(blobbers) == 0 {
		return "", fmt.Errorf("unable to get free blobbers for allocation")
	}

	input["blobbers"] = blobbers

	payload, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	return smartContractTxn(
		zcncore.StorageSmartContractAddress,
		transaction.NEW_FREE_ALLOCATION,
		string(payload),
		lock,
	)
}

func decodeTicket(ticket string) (map[string]interface{}, error, uint64) {
	decoded, err := base64.StdEncoding.DecodeString(ticket)
	if err != nil {
		return nil, err, 0
	}

	input := make(map[string]interface{})
	if err = json.Unmarshal(decoded, &input); err != nil {
		return nil, err, 0
	}

	str := fmt.Sprintf("%v", input["marker"])
	decodedMarker, _ := base64.StdEncoding.DecodeString(str)
	markerInput := make(map[string]interface{})
	if err = json.Unmarshal(decodedMarker, &markerInput); err != nil {
		return nil, err, 0
	}

	result := make(map[string]interface{})
	result["recipient_public_key"] = input["recipient_public_key"]

	lock := markerInput["free_tokens"]
	markerStr, _ := json.Marshal(markerInput)
	result["marker"] = string(markerStr)

	s, _ := strconv.ParseFloat(string(fmt.Sprintf("%v", lock)), 64)
	return result, nil, uint64(zcncoremobile.ConvertToValue(s))
}

func getFreeAllocationBlobbers(request map[string]interface{}) ([]string, error) {
	data, _ := json.Marshal(request)

	params := make(map[string]string)
	params["free_allocation_data"] = string(data)

	allocBlobber, err := zboxutil.MakeSCRestAPICall(sdk.STORAGE_SCADDRESS, "/free_alloc_blobbers", params, nil)
	if err != nil {
		return nil, err
	}
	var allocBlobberIDs []string

	err = json.Unmarshal(allocBlobber, &allocBlobberIDs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal blobber IDs")
	}

	return allocBlobberIDs, nil
}

type TransactionCallback struct {
	wg      *sync.WaitGroup
	success bool
	balance int64
}

func smartContractTxn(address, method, input string, value uint64) (string, error) {
	tcb := &TransactionCallback{}
	tcb.wg = &sync.WaitGroup{}
	tcb.wg.Add(1)

	zcntxn, err := zcncore.NewTransaction(tcb, 0, 1)
	if err != nil {
		return "", err
	}

	l.Logger.Info("Calling SC txn with values :", zap.Any("method", method), zap.Any("input", input), zap.Any("value", value))
	err = zcntxn.ExecuteSmartContract(address, method, input, value)
	if err != nil {
		tcb.wg.Done()
		return "", err
	}
	tcb.wg.Wait()
	if len(zcntxn.GetTransactionError()) > 0 {
		return "", errors.New("smart_contract_txn_get_error", zcntxn.GetTransactionError())
	}
	tcb.wg.Add(1)
	err = zcntxn.Verify()
	if err != nil {
		tcb.wg.Done()
		return "", errors.New("smart_contract_txn_verify_error", err.Error())
	}
	tcb.wg.Wait()

	if len(zcntxn.GetVerifyError()) > 0 {
		return "", errors.New("smart_contract_txn_verify_error", zcntxn.GetVerifyError())
	}
	return zcntxn.GetTransactionHash(), nil
}

func (t *TransactionCallback) OnBalanceAvailable(status int, value int64, info string) {
	defer t.wg.Done()
	if status == zcncore.StatusSuccess {
		t.success = true
	} else {
		t.success = false
	}
	t.balance = value
}

func (t *TransactionCallback) OnTransactionComplete(zcntxn *zcncore.Transaction, status int) {
	defer t.wg.Done()
	if status == zcncore.StatusSuccess {
		t.success = true
	} else {
		t.success = false
	}
}

func (t *TransactionCallback) OnVerifyComplete(zcntxn *zcncore.Transaction, status int) {
	defer t.wg.Done()
	if status == zcncore.StatusSuccess {
		t.success = true
	} else {
		t.success = false
	}
}

func (t *TransactionCallback) OnAuthComplete(zcntxn *zcncore.Transaction, status int) {}
