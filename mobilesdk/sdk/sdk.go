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

	"github.com/0chain/gosdk/core/sys"
	"github.com/pkg/errors"

	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/core/version"
	"github.com/0chain/gosdk/zboxcore/client"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zboxcore/sdk"

	"github.com/0chain/gosdk/mobilesdk/zbox"
	"github.com/0chain/gosdk/mobilesdk/zboxapi"
	"github.com/0chain/gosdk/zcncore"
)

var nonce = int64(0)

type Autorizer interface {
	Auth(msg string) (string, error)
}

// ChainConfig - blockchain config
type ChainConfig struct {
	ChainID           string   `json:"chain_id,omitempty"`
	PreferredBlobbers []string `json:"preferred_blobbers"`
	BlockWorker       string   `json:"block_worker"`
	SignatureScheme   string   `json:"signature_scheme"`
	// ZboxHost 0box api host host: "https://0box.dev.0chain.net"
	ZboxHost string `json:"zbox_host"`
	// ZboxAppType app type name
	ZboxAppType string `json:"zbox_app_type"`
}

// StorageSDK - storage SDK config
type StorageSDK struct {
	chainconfig *ChainConfig
	client      *client.Client
}

// SetLogFile setup log level for core libraries
//   - logFile: the output file of logs
//   - verbose: output detail logs
func SetLogFile(logFile string, verbose bool) {
	zcncore.SetLogFile(logFile, verbose)
	sdk.SetLogFile(logFile, verbose)
}

// SetLogLevel set the log level.
//
//	`lvl` - 0 disabled; higher number (upto 4) more verbosity
func SetLogLevel(logLevel int) {
	sdk.SetLogLevel(logLevel)
}

// Init init the sdk with chain config
//   - chainConfigJson: chain config json string
func Init(chainConfigJson string) error {
	return zcncore.Init(chainConfigJson)
}

// InitStorageSDK init storage sdk from config
//   - clientJson example
//     {
//     "client_id":"8f6ce6457fc04cfb4eb67b5ce3162fe2b85f66ef81db9d1a9eaa4ffe1d2359e0",
//     "client_key":"c8c88854822a1039c5a74bdb8c025081a64b17f52edd463fbecb9d4a42d15608f93b5434e926d67a828b88e63293b6aedbaf0042c7020d0a96d2e2f17d3779a4",
//     "keys":[
//     {
//     "public_key":"c8c88854822a1039c5a74bdb8c025081a64b17f52edd463fbecb9d4a42d15608f93b5434e926d67a828b88e63293b6aedbaf0042c7020d0a96d2e2f17d3779a4",
//     "private_key":"72f480d4b1e7fb76e04327b7c2348a99a64f0ff2c5ebc3334a002aa2e66e8506"
//     }],
//     "mnemonics":"abandon mercy into make powder fashion butter ignore blade vanish plastic shock learn nephew matrix indoor surge document motor group barely offer pottery antenna",
//     "version":"1.0",
//     "date_created":"1668667145",
//     "nonce":0
//     }
//   - configJson example
//     {
//     "block_worker": "https://dev.0chain.net/dns",
//     "signature_scheme": "bls0chain",
//     "min_submit": 50,
//     "min_confirmation": 50,
//     "confirmation_chain_length": 3,
//     "max_txn_query": 5,
//     "query_sleep_time": 5,
//     "preferred_blobbers": ["https://dev.0chain.net/blobber02","https://dev.0chain.net/blobber03"],
//     "chain_id":"0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe",
//     "ethereum_node":"https://ropsten.infura.io/v3/xxxxxxxxxxxxxxx",
//     "zbox_host":"https://0box.dev.0chain.net",
//     "zbox_app_type":"vult",
//     }
func InitStorageSDK(clientJson string, configJson string) (*StorageSDK, error) {
	l.Logger.Info("Start InitStorageSDK")
	configObj := &ChainConfig{}
	err := json.Unmarshal([]byte(configJson), configObj)
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
	if err = sdk.InitStorageSDK(clientJson,
		configObj.BlockWorker,
		configObj.ChainID,
		configObj.SignatureScheme,
		configObj.PreferredBlobbers,
		0); err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	l.Logger.Info("InitStorageSDK success")

	if configObj.ZboxHost != "" && configObj.ZboxAppType != "" {
		zboxapi.Init(configObj.ZboxHost, configObj.ZboxAppType)
		l.Logger.Info("InitZboxApi success")
	} else {
		l.Logger.Info("InitZboxApi skipped")
	}

	l.Logger.Info("Init successful")

	return &StorageSDK{client: client.GetClient(), chainconfig: configObj}, nil
}

// CreateAllocation creating new allocation
//   - datashards: number of data shards, effects upload and download speeds
//   - parityshards: number of parity shards, effects availability
//   - size: size of space reserved on blobbers
//   - expiration: duration to allocation expiration
//   - lock: lock write pool with given number of tokens
//   - blobberAuthTickets: list of blobber auth tickets needed for the restricted blobbers
func (s *StorageSDK) CreateAllocation(datashards, parityshards int, size, expiration int64, lock string, blobberAuthTickets []string) (*zbox.Allocation, error) {
	readPrice := sdk.PriceRange{Min: 0, Max: math.MaxInt64}
	writePrice := sdk.PriceRange{Min: 0, Max: math.MaxInt64}

	l, err := util.ParseCoinStr(lock)
	if err != nil {
		return nil, err
	}

	options := sdk.CreateAllocationOptions{
		DataShards:         datashards,
		ParityShards:       parityshards,
		Size:               size,
		ReadPrice:          readPrice,
		WritePrice:         writePrice,
		Lock:               uint64(l),
		BlobberIds:         []string{},
		FileOptionsParams:  &sdk.FileOptionsParameters{},
		BlobberAuthTickets: blobberAuthTickets,
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

// CreateAllocationWithBlobbers - creating new allocation with list of blobbers
//   - name: allocation name
//   - datashards: number of data shards, effects upload and download speeds
//   - parityshards: number of parity shards, effects availability
//   - size: size of space reserved on blobbers
//   - expiration: duration to allocation expiration
//   - lock: lock write pool with given number of tokens
//   - blobberUrls: concat blobber urls with comma. leave it as empty if you don't have any preferred blobbers
//   - blobberIds: concat blobber ids with comma. leave it as empty if you don't have any preferred blobbers
func (s *StorageSDK) CreateAllocationWithBlobbers(name string, datashards, parityshards int, size int64, lock string, blobberUrls, blobberIds string, blobberAuthTickets []string) (*zbox.Allocation, error) {
	readPrice := sdk.PriceRange{Min: 0, Max: math.MaxInt64}
	writePrice := sdk.PriceRange{Min: 0, Max: math.MaxInt64}

	l, err := util.ParseCoinStr(lock)
	if err != nil {
		return nil, err
	}

	options := sdk.CreateAllocationOptions{
		DataShards:         datashards,
		ParityShards:       parityshards,
		Size:               size,
		Lock:               l,
		WritePrice:         writePrice,
		ReadPrice:          readPrice,
		BlobberAuthTickets: blobberAuthTickets,
	}

	if blobberUrls != "" {
		urls := strings.Split(blobberUrls, ",")
		if len(urls) > 0 {
			ids, err := sdk.GetBlobberIds(urls)
			if err != nil {
				return nil, err
			}
			options.BlobberIds = ids
		}
	}

	if blobberIds != "" {
		ids := strings.Split(blobberIds, ",")
		if len(ids) > 0 {
			options.BlobberIds = ids
		}
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

// GetAllocation retrieve allocation from ID
//   - allocationID: allocation ID
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

// GetAllocations retrieve list of allocations owned by the wallet
func (s *StorageSDK) GetAllocations() (string, error) {
	sdkAllocations, err := sdk.GetAllocations()
	if err != nil {
		return "", err
	}
	result := make([]*zbox.Allocation, len(sdkAllocations))
	for i, sdkAllocation := range sdkAllocations {
		allocationObj := zbox.ToAllocation(sdkAllocation)
		result[i] = allocationObj
	}
	retBytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
}

// GetAllocationFromAuthTicket retrieve allocation from Auth ticket
// AuthTicket is a signed message from the blobber authorizing the client to access the allocation.
// It's issued by the allocation owner and can be used by a non-owner to access the allocation.
//   - authTicket: auth ticket
func (s *StorageSDK) GetAllocationFromAuthTicket(authTicket string) (*zbox.Allocation, error) {
	sdkAllocation, err := sdk.GetAllocationFromAuthTicket(authTicket)
	if err != nil {
		return nil, err
	}
	return zbox.ToAllocation(sdkAllocation), nil
}

// GetAllocationStats retrieve allocation stats by allocation ID
//   - allocationID: allocation ID
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

// FinalizeAllocation finalize allocation
//   - allocationID: allocation ID
func (s *StorageSDK) FinalizeAllocation(allocationID string) (string, error) {
	hash, _, err := sdk.FinalizeAllocation(allocationID)
	return hash, err
}

// CancelAllocation cancel allocation by ID
//   - allocationID: allocation ID
func (s *StorageSDK) CancelAllocation(allocationID string) (string, error) {
	hash, _, err := sdk.CancelAllocation(allocationID)
	return hash, err
}

// GetReadPoolInfo is to get information about the read pool for the allocation
//   - clientID: client ID
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
// WritePoolLock lock write pool with given number of tokens
//   - durInSeconds: duration in seconds
//   - tokens: number of tokens
//   - fee: fee of the transaction
//   - allocID: allocation ID
func (s *StorageSDK) WritePoolLock(durInSeconds int64, tokens, fee float64, allocID string) error {
	_, _, err := sdk.WritePoolLock(
		allocID,
		zcncore.ConvertTokenToSAS(tokens),
		zcncore.ConvertTokenToSAS(fee))
	return err
}

// GetVersion getting current version for gomobile lib
func (s *StorageSDK) GetVersion() string {
	return version.VERSIONSTR
}

// UpdateAllocation update allocation settings with new expiry and size
//   - size: size of space reserved on blobbers
//   - extend: extend allocation
//   - allocationID: allocation ID
//   - lock: Number of tokens to lock to the allocation after the update
func (s *StorageSDK) UpdateAllocation(size int64, extend bool, allocationID string, lock uint64) (hash string, err error) {
	if lock > math.MaxInt64 {
		return "", errors.Errorf("int64 overflow in lock")
	}

	hash, _, err = sdk.UpdateAllocation(size, extend, allocationID, lock, "", "", "", false, &sdk.FileOptionsParameters{})
	return hash, err
}

// GetBlobbersList get list of active blobbers, and format them as array json string
func (s *StorageSDK) GetBlobbersList() (string, error) {
	blobbs, err := sdk.GetBlobbers(true, false)
	if err != nil {
		return "", err
	}
	retBytes, err := json.Marshal(blobbs)
	if err != nil {
		return "", err
	}
	return string(retBytes), nil
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

// RedeeemFreeStorage given a free storage ticket, create a new free allocation
//   - ticket: free storage ticket
func (s *StorageSDK) RedeemFreeStorage(ticket string) (string, error) {
	recipientPublicKey, marker, lock, err := decodeTicket(ticket)
	if err != nil {
		return "", err
	}

	if recipientPublicKey != client.GetClientPublicKey() {
		return "", fmt.Errorf("invalid_free_marker: free marker is not assigned to your wallet")
	}

	hash, _, err := sdk.CreateFreeAllocation(marker, lock)
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
	return string(recipientPublicKey), string(markerStr), zcncore.ConvertTokenToSAS(s), nil
}

// RegisterAuthorizer Client can extend interface and FaSS implementation to this register like this:
//
//	public class Autorizer extends Pkg.Autorizer {
//		public void Auth() {
//			// do something here
//		}
//	}
func RegisterAuthorizer(auth Autorizer) {
	sys.Authorize = auth.Auth
}
