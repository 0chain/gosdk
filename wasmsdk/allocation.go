//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

const TOKEN_UNIT int64 = 1e10

type fileResp struct {
	sdk.FileInfo
	Name string `json:"name"`
	Path string `json:"path"`
}

type decodeAuthTokenResp struct {
	RecipientPublicKey string `json:"recipient_public_key"`
	Marker             string `json:"marker"`
	Tokens             uint64 `json:"tokens"`
}

// getBlobberIds retrieves blobber ids from the given blobber urls
//   - blobberUrls is the list of blobber urls
func getBlobberIds(blobberUrls []string) ([]string, error) {
	return sdk.GetBlobberIds(blobberUrls)
}

// createfreeallocation creates a free allocation
//   - freeStorageMarker is the free storage marker
func createfreeallocation(freeStorageMarker string) (string, error) {
	allocationID, _, err := sdk.CreateFreeAllocation(freeStorageMarker, 0)
	if err != nil {
		sdkLogger.Error("Error creating free allocation: ", err)
		return "", err
	}
	return allocationID, err
}

// getAllocationBlobbers retrieves allocation blobbers
//   - preferredBlobberURLs is the list of preferred blobber urls
//   - dataShards is the number of data shards
//   - parityShards is the number of parity shards
//   - size is the size of the allocation
//   - minReadPrice is the minimum read price
//   - maxReadPrice is the maximum read price
//   - minWritePrice is the minimum write price
//   - maxWritePrice is the maximum write price
//   - isRestricted is the restricted flag
//   - force is the force flag
func getAllocationBlobbers(preferredBlobberURLs []string,
	dataShards, parityShards int, size int64,
	minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64, isRestricted int, force bool) ([]string, error) {

	if len(preferredBlobberURLs) > 0 {
		return sdk.GetBlobberIds(preferredBlobberURLs)
	}

	return sdk.GetAllocationBlobbers(dataShards, parityShards, size, isRestricted, sdk.PriceRange{
		Min: uint64(minReadPrice),
		Max: uint64(maxReadPrice),
	}, sdk.PriceRange{
		Min: uint64(minWritePrice),
		Max: uint64(maxWritePrice),
	}, force)
}

// createAllocation creates an allocation given allocation creation parameters
//   - datashards is the number of data shards. Data uploaded to the allocation will be split and distributed across these shards.
//   - parityshards is the number of parity shards. Parity shards are used to replicate datashards for redundancy.
//   - size is the size of the allocation in bytes.
//   - minReadPrice is the minimum read price set by the client.
//   - maxReadPrice is the maximum read price set by the client.
//   - minWritePrice is the minimum write price set by the client.
//   - maxWritePrice is the maximum write price set by the client.
//   - lock is the lock value to add to the allocation.
//   - blobberIds is the list of blobber ids.
//   - blobberAuthTickets is the list of blobber auth tickets in case of using restricted blobbers.
func createAllocation(datashards, parityshards int, size int64,
	minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64, lock int64, blobberIds, blobberAuthTickets []string, setThirdPartyExtendable, IsEnterprise, force bool) (
	*transaction.Transaction, error) {

	options := sdk.CreateAllocationOptions{
		DataShards:   datashards,
		ParityShards: parityshards,
		Size:         size,
		ReadPrice: sdk.PriceRange{
			Min: uint64(minReadPrice),
			Max: uint64(maxReadPrice),
		},
		WritePrice: sdk.PriceRange{
			Min: uint64(minWritePrice),
			Max: uint64(maxWritePrice),
		},
		Lock:                 uint64(lock),
		BlobberIds:           blobberIds,
		ThirdPartyExtendable: setThirdPartyExtendable,
		IsEnterprise:         IsEnterprise,
		BlobberAuthTickets:   blobberAuthTickets,
		Force:                force,
	}

	sdkLogger.Info(options)
	_, _, txn, err := sdk.CreateAllocationWith(options)

	return txn, err
}

// listAllocations retrieves the list of allocations owned by the client
func listAllocations() ([]*sdk.Allocation, error) {
	return sdk.GetAllocations()
}

// transferAllocation transfers the ownership of an allocation to a new owner
//   - allocationID is the allocation id
//   - newOwnerId is the new owner id
//   - newOwnerPublicKey is the new owner public key
func transferAllocation(allocationID, newOwnerId, newOwnerPublicKey string) error {
	if allocationID == "" {
		return RequiredArg("allocationID")
	}

	if newOwnerId == "" {
		return RequiredArg("newOwnerId")
	}

	if newOwnerPublicKey == "" {
		return RequiredArg("newOwnerPublicKey")
	}

	_, _, err := sdk.TransferAllocation(allocationID, newOwnerId, newOwnerPublicKey)

	if err == nil {
		clearAllocation(allocationID)
	}

	return err
}

// UpdateForbidAllocation updates the permissions of an allocation, given the permission parameters in a forbid-first manner.
//   - allocationID: allocation ID
//   - forbidupload: forbid upload flag, if true, uploading files to the allocation is forbidden
//   - forbiddelete: forbid delete flag, if true, deleting files from the allocation is forbidden
//   - forbidupdate: forbid update flag, if true, updating files in the allocation is forbidden
//   - forbidmove: forbid move flag, if true, moving files in the allocation is forbidden
//   - forbidcopy: forbid copy flag, if true, copying files in the allocation is forbidden
//   - forbidrename: forbid rename flag, if true, renaming files in the allocation is forbidden
func UpdateForbidAllocation(allocationID string, forbidupload, forbiddelete, forbidupdate, forbidmove, forbidcopy, forbidrename bool) (string, error) {

	hash, _, err := sdk.UpdateAllocation(
		0,            //size,
		false,        //extend,
		allocationID, // allocID,
		0,            //lock,
		"",           //addBlobberId,
		"",           //addBlobberAuthTicket
		"",           //removeBlobberId,
		false,        //thirdPartyExtendable,
		&sdk.FileOptionsParameters{
			ForbidUpload: sdk.FileOptionParam{Changed: forbidupload, Value: forbidupload},
			ForbidDelete: sdk.FileOptionParam{Changed: forbiddelete, Value: forbiddelete},
			ForbidUpdate: sdk.FileOptionParam{Changed: forbidupdate, Value: forbidupdate},
			ForbidMove:   sdk.FileOptionParam{Changed: forbidmove, Value: forbidmove},
			ForbidCopy:   sdk.FileOptionParam{Changed: forbidcopy, Value: forbidcopy},
			ForbidRename: sdk.FileOptionParam{Changed: forbidrename, Value: forbidrename},
		},
	)

	return hash, err

}

// freezeAllocation freezes one of the client's allocations, given its ID
// Freezing the allocation means to forbid all the operations on the files in the allocation.
//   - allocationID: allocation ID
func freezeAllocation(allocationID string) (string, error) {

	hash, _, err := sdk.UpdateAllocation(
		0,            //size,
		false,        //extend,
		allocationID, // allocID,
		0,            //lock,
		"",           //addBlobberId,
		"",           //addBlobberAuthTicket
		"",           //removeBlobberId,
		false,        //thirdPartyExtendable,
		&sdk.FileOptionsParameters{
			ForbidUpload: sdk.FileOptionParam{Changed: true, Value: true},
			ForbidDelete: sdk.FileOptionParam{Changed: true, Value: true},
			ForbidUpdate: sdk.FileOptionParam{Changed: true, Value: true},
			ForbidMove:   sdk.FileOptionParam{Changed: true, Value: true},
			ForbidCopy:   sdk.FileOptionParam{Changed: true, Value: true},
			ForbidRename: sdk.FileOptionParam{Changed: true, Value: true},
		},
	)

	if err == nil {
		clearAllocation(allocationID)
	}

	return hash, err

}

// cancelAllocation cancels one of the client's allocations, given its ID
//   - allocationID: allocation ID
func cancelAllocation(allocationID string) (string, error) {
	hash, _, err := sdk.CancelAllocation(allocationID)

	if err == nil {
		clearAllocation(allocationID)
	}

	return hash, err
}

// updateAllocationWithRepair updates the allocation settings and repairs it if necessary.
// Repair means to sync the user's data under the allocation on all the blobbers
// and fill the missing data on the blobbers that have missing data.
// Check the system documentation for more information about the repoair process.
//   - allocationID: allocation ID
//   - size: size of the allocation
//   - extend: extend flag
//   - lock: lock value to add to the allocation
//   - addBlobberId: blobber ID to add to the allocation
//   - addBlobberAuthTicket: blobber auth ticket to add to the allocation, in case of restricted blobbers
//   - removeBlobberId: blobber ID to remove from the allocation
func updateAllocationWithRepair(allocationID string,
	size int64,
	extend bool,
	lock int64,
	addBlobberId, addBlobberAuthTicket, removeBlobberId string) (string, error) {
	sdk.SetWasm()
	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		return "", err
	}

	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg, isRepair: true}
	wg.Add(1)

	alloc, hash, isRepairRequired, err := allocationObj.UpdateWithStatus(size, extend, uint64(lock), addBlobberId, addBlobberAuthTicket, removeBlobberId, false, &sdk.FileOptionsParameters{}, statusBar)
	if err != nil {
		return hash, err
	}
	clearAllocation(allocationID)

	if isRepairRequired {
		addWebWorkers(alloc)
		if removeBlobberId != "" {
			jsbridge.RemoveWorker(removeBlobberId)
		}
		err := alloc.RepairAlloc(statusBar)
		if err != nil {
			return "", err
		}
		wg.Wait()
		if statusBar.err != nil {
			return "", statusBar.err
		}
	}

	return hash, err
}

// updateAllocation updates the allocation settings
//   - allocationID: allocation ID
//   - size: new size of the allocation
//   - extend: extend flag, whether to extend the allocation's expiration date
//   - lock: lock value to add to the allocation
//   - addBlobberId: blobber ID to add to the allocation
//   - addBlobberAuthTicket: blobber auth ticket to add to the allocation, in case of restricted blobbers
//   - removeBlobberId: blobber ID to remove from the allocation
//   - setThirdPartyExtendable: third party extendable flag, if true, the allocation can be extended (in terms of size) by a non-owner client
func updateAllocation(allocationID string,
	size int64, extend bool,
	lock int64,
	addBlobberId, addBlobberAuthTicket, removeBlobberId string, setThirdPartyExtendable bool) (string, error) {
	hash, _, err := sdk.UpdateAllocation(size, extend, allocationID, uint64(lock), addBlobberId, addBlobberAuthTicket, removeBlobberId, setThirdPartyExtendable, &sdk.FileOptionsParameters{})

	if err == nil {
		clearAllocation(allocationID)
	}

	return hash, err
}

// getAllocationMinLock retrieves the minimum lock value for the allocation creation, as calculated by the network.
// Lock value is the amount of tokens that the client needs to lock in the allocation's write pool
// to be able to pay for the write operations.
//   - datashards: number of data shards
//   - parityshards: number of parity shards.
//   - size: size of the allocation.
//   - maxwritePrice: maximum write price set by the client.
func getAllocationMinLock(datashards, parityshards int,
	size int64,
	maxwritePrice uint64,
) (int64, error) {
	writePrice := sdk.PriceRange{Min: 0, Max: maxwritePrice}

	value, err := sdk.GetAllocationMinLock(datashards, parityshards, size, writePrice)
	if err != nil {
		sdkLogger.Error(err)
		return 0, err
	}
	sdkLogger.Info("allocation Minlock value", value)
	return value, nil
}

// getUpdateAllocationMinLock retrieves the minimum lock value for the allocation after update, as calculated by the network based on the update parameters.
// Lock value is the amount of tokens that the client needs to lock in the allocation's write pool
// to be able to pay for the write operations.
//   - allocationID: allocation ID
//   - size: new size of the allocation
//   - extend: extend flag, whether to extend the allocation's expiration date
//   - addBlobberId: blobber ID to add to the allocation
//   - removeBlobberId: blobber ID to remove from the allocation
func getUpdateAllocationMinLock(
	allocationID string,
	size int64,
	extend bool,
	addBlobberId, removeBlobberId string) (int64, error) {
	return sdk.GetUpdateAllocationMinLock(allocationID, size, extend, addBlobberId, removeBlobberId)
}

// getRemoteFileMap list all files in an allocation from the blobbers.
//   - allocationID: allocation ID
func getRemoteFileMap(allocationID string) ([]*fileResp, error) {
	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}
	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		return nil, err
	}

	ref, err := allocationObj.GetRemoteFileMap(nil, "/")
	if err != nil {
		sdkLogger.Error(err)
		return nil, err
	}

	fileResps := make([]*fileResp, 0)
	for path, data := range ref {
		paths := strings.SplitAfter(path, "/")
		var resp = fileResp{
			Name:     paths[len(paths)-1],
			Path:     path,
			FileInfo: data,
		}
		fileResps = append(fileResps, &resp)
	}

	return fileResps, nil
}

// lockWritePool locks given number of tokes for given duration in write pool.
//   - allocID: allocation id
//   - tokens:  sas tokens
//   - fee: sas tokens
func lockWritePool(allocID string, tokens, fee uint64) (string, error) {
	hash, _, err := sdk.WritePoolLock(allocID, tokens, fee)
	return hash, err
}

// lockStakePool stake number of tokens for a given provider given its type and id
//   - providerType: provider type (1: miner, 2:sharder, 3:blobber, 4:validator, 5:authorizer)
//   - tokens: amount of tokens to lock (in SAS)
//   - fee: transaction fees (in SAS)
//   - providerID: provider id
func lockStakePool(providerType, tokens, fee uint64, providerID string) (string, error) {

	hash, _, err := sdk.StakePoolLock(sdk.ProviderType(providerType), providerID,
		tokens, fee)
	return hash, err
}

// unlockWritePool unlocks the read pool
//   - tokens: amount of tokens to lock (in SAS)
//   - fee: transaction fees (in SAS)
func lockReadPool(tokens, fee uint64) (string, error) {
	hash, _, err := sdk.ReadPoolLock(tokens, fee)
	return hash, err
}

// unLockWritePool unlocks the write pool
//   - fee: transaction fees (in SAS)
func unLockReadPool(fee uint64) (string, error) {
	hash, _, err := sdk.ReadPoolUnlock(fee)
	return hash, err
}

// unlockWritePool unlocks the write pool
//   - providerType: provider type (1: miner, 2:sharder, 3:blobber, 4:validator, 5:authorizer)
//   - fee: transaction fees (in SAS)
//   - providerID: provider id
func unlockStakePool(providerType, fee uint64, providerID string) (int64, error) {
	unstake, _, err := sdk.StakePoolUnlock(sdk.ProviderType(providerType), providerID, fee)
	return unstake, err
}

// getSkatePoolInfo is to get information about the stake pool for the allocation
//   - providerType: provider type (1: miner, 2:sharder, 3:blobber, 4:validator, 5:authorizer)
//   - providerID: provider id
func getSkatePoolInfo(providerType int, providerID string) (*sdk.StakePoolInfo, error) {

	info, err := sdk.GetStakePoolInfo(sdk.ProviderType(providerType), providerID)

	if err != nil {
		return nil, err
	}
	return info, err
}

// getReadPoolInfo is to get information about the read pool for the allocation
//   - clientID: client id
func getReadPoolInfo(clientID string) (*sdk.ReadPool, error) {
	readPool, err := sdk.GetReadPoolInfo(clientID)
	if err != nil {
		return nil, err
	}

	return readPool, nil
}

// getAllocationWith retrieves the information of a free or a shared allocation object given the auth ticket.
// A free allocation is an allocation that is created to the user using Vult app for the first time with no fees.
// A shared allocation is an allocation that has some shared files. The user who needs
// to access those files needs first to read the information of this allocation.
//   - authTicket: auth ticket usually used by a non-owner to access a shared allocation
func getAllocationWith(authTicket string) (*sdk.Allocation, error) {
	sdk.SetWasm()
	sdkAllocation, err := sdk.GetAllocationFromAuthTicket(authTicket)
	if err != nil {
		return nil, err
	}
	return sdkAllocation, err
}

// decodeAuthTicket decodes the auth ticket and returns the recipient public key and the tokens
//   - ticket: auth ticket
func decodeAuthTicket(ticket string) (*decodeAuthTokenResp, error) {
	resp := &decodeAuthTokenResp{}

	decoded, err := base64.StdEncoding.DecodeString(ticket)
	if err != nil {
		sdkLogger.Error("error decoding", err.Error())
		return resp, err
	}

	input := make(map[string]interface{})
	if err = json.Unmarshal(decoded, &input); err != nil {
		sdkLogger.Error("error unmarshalling json", err.Error())
		return resp, err
	}

	if marker, ok := input["marker"]; ok {
		str := fmt.Sprintf("%v", marker)
		decodedMarker, _ := base64.StdEncoding.DecodeString(str)
		markerInput := make(map[string]interface{})
		if err = json.Unmarshal(decodedMarker, &markerInput); err != nil {
			sdkLogger.Error("error unmarshaling markerInput", err.Error())
			return resp, err
		}
		lock := markerInput["free_tokens"]
		markerStr, _ := json.Marshal(markerInput)
		resp.Marker = string(markerStr)
		s, _ := strconv.ParseFloat(string(fmt.Sprintf("%v", lock)), 64)
		resp.Tokens = convertTokenToSAS(s)
	}

	if public_key, ok := input["recipient_public_key"]; ok {
		recipientPublicKey, ok := public_key.(string)
		if !ok {
			return resp, fmt.Errorf("recipient_public_key is required")
		}
		resp.RecipientPublicKey = string(recipientPublicKey)
	}

	return resp, nil
}

// convertTokenToSAS converts tokens in ZCN to SAS.
// 1 ZCN = 1e10 SAS
//   - token: token value in ZCN
func convertTokenToSAS(token float64) uint64 {
	return uint64(token * float64(TOKEN_UNIT))
}

// allocationRepair issue repair process for an allocation, starting from a specific path.
// Repair means to sync the user's data under the allocation on all the blobbers
// and fill the missing data on the blobbers that have missing data.
// Check the system documentation for more information about the repoair process.
//   - allocationID: allocation ID
//   - remotePath: remote path
func allocationRepair(allocationID, remotePath string) error {
	if len(allocationID) == 0 {
		return RequiredArg("allocationID")
	}
	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		return err
	}
	sdk.SetWasm()
	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg, isRepair: true}
	wg.Add(1)

	err = allocationObj.StartRepair("/tmp", remotePath, statusBar)
	if err != nil {
		PrintError("Upload failed.", err)
		return err
	}
	wg.Wait()
	if !statusBar.success {
		return errors.New("upload failed: unknown")
	}
	return nil
}

// repairSize retrieves the repair size for a specific path in an allocation.
// Repair size is the size of the data that needs to be repaired in the allocation.
//   - allocationID: allocation ID
//   - remotePath: remote path
func repairSize(allocationID, remotePath string) (sdk.RepairSize, error) {
	alloc, err := sdk.GetAllocation(allocationID)
	if err != nil {
		return sdk.RepairSize{}, err
	}
	return alloc.RepairSize(remotePath)
}
