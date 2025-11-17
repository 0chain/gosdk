//go:build js && wasm
// +build js,wasm

package main

import (
	"errors"
	"strings"
	"sync"
	"syscall/js"
	"time"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/wasmsdk/jsbridge"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

// MW (multi-wallet) variants of allocation helpers. Each function accepts an
// additional `key string` argument (the multi-wallet support key). When the
// underlying SDK supports explicit key-based signing (via variadic keys) the
// key is forwarded. When selecting an allocation object the cached helper
// `getAllocation(allocationID, key)` is used to select the allocation for that
// key.

// createfreeallocationMW creates a free allocation
func createfreeallocationMW(freeStorageMarker string, key string) (string, error) {
	allocationID, _, err := sdk.CreateFreeAllocation(freeStorageMarker, 0, key)
	if err != nil {
		sdkLogger.Error("Error creating free allocation: ", err)
		return "", err
	}
	return allocationID, err
}

// getAllocationBlobbersMW retrieves allocation blobbers (key is ignored)
func getAllocationBlobbersMW(preferredBlobberURLs []string,
	dataShards, parityShards int, size int64,
	minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64, isRestricted int, force bool, key string) ([]string, error) {

	if len(preferredBlobberURLs) > 0 {
		return sdk.GetBlobberIds(preferredBlobberURLs)
	}

	return sdk.GetAllocationBlobbers(sdk.StorageV2, dataShards, parityShards, size, isRestricted, sdk.PriceRange{
		Min: uint64(minReadPrice),
		Max: uint64(maxReadPrice),
	}, sdk.PriceRange{
		Min: uint64(minWritePrice),
		Max: uint64(maxWritePrice),
	}, force)
}

// getBlobberIdsMW retrieves blobber ids from the given blobber urls (key ignored)
func getBlobberIdsMW(blobberUrls []string, key string) ([]string, error) {
	return sdk.GetBlobberIds(blobberUrls)
}

// reloadAllocationMW reload allocation from blockchain and update cache (supports key)
func reloadAllocationMW(allocationID, key string) (*sdk.Allocation, error) {
	a, err := sdk.GetAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}

	it := &cachedAllocation{
		Allocation: a,
		Expiration: time.Now().Add(5 * time.Minute),
	}

	cachedAllocations.Add(allocationID, it)

	return it.Allocation, nil
}

// createAllocationMW creates an allocation given allocation creation parameters
func createAllocationMW(datashards, parityshards int, size, authRoundExpiry int64,
	minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64, lock int64, blobberIds, blobberAuthTickets []string, setThirdPartyExtendable, IsEnterprise, force bool, key string) (
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
		StorageVersion:       sdk.StorageV2,
		BlobberAuthTickets:   blobberAuthTickets,
		Force:                force,
		AuthRoundExpiry:      authRoundExpiry,
	}

	sdkLogger.Info(options)
	_, _, txn, err := sdk.CreateAllocationWith(options, key)
	return txn, err
}

// listAllocationsMW retrieves the list of allocations owned by the client
func listAllocationsMW(key string) ([]*sdk.Allocation, error) {
	return sdk.GetAllocations(key)
}

// transferAllocationMW transfers the ownership of an allocation to a new owner
func transferAllocationMW(allocationID, newOwnerId, newOwnerPublicKey, key string) error {
	if allocationID == "" {
		return RequiredArg("allocationID")
	}
	if newOwnerId == "" {
		return RequiredArg("newOwnerId")
	}
	if newOwnerPublicKey == "" {
		return RequiredArg("newOwnerPublicKey")
	}

	_, _, err := sdk.TransferAllocation(allocationID, newOwnerId, newOwnerPublicKey, key)
	if err == nil {
		clearAllocation(allocationID)
	}
	return err
}

// UpdateForbidAllocationMW updates allocation file options with optional key
func UpdateForbidAllocationMW(allocationID string, forbidupload, forbiddelete, forbidupdate, forbidmove, forbidcopy, forbidrename bool, key string) (string, error) {
	fop := &sdk.FileOptionsParameters{
		ForbidUpload: sdk.FileOptionParam{Changed: forbidupload, Value: forbidupload},
		ForbidDelete: sdk.FileOptionParam{Changed: forbiddelete, Value: forbiddelete},
		ForbidUpdate: sdk.FileOptionParam{Changed: forbidupdate, Value: forbidupdate},
		ForbidMove:   sdk.FileOptionParam{Changed: forbidmove, Value: forbidmove},
		ForbidCopy:   sdk.FileOptionParam{Changed: forbidcopy, Value: forbidcopy},
		ForbidRename: sdk.FileOptionParam{Changed: forbidrename, Value: forbidrename},
	}

	hash, _, err := sdk.UpdateAllocation(0, 0, false, allocationID, 0, "", "", "", "", "", false, fop, "", key)
	return hash, err
}

// freezeAllocationMW freezes one of the client's allocations
func freezeAllocationMW(allocationID, key string) (string, error) {
	hash, _, err := sdk.UpdateAllocation(0, 0, false, allocationID, 0, "", "", "", "", "", false, &sdk.FileOptionsParameters{
		ForbidUpload: sdk.FileOptionParam{Changed: true, Value: true},
		ForbidDelete: sdk.FileOptionParam{Changed: true, Value: true},
		ForbidUpdate: sdk.FileOptionParam{Changed: true, Value: true},
		ForbidMove:   sdk.FileOptionParam{Changed: true, Value: true},
		ForbidCopy:   sdk.FileOptionParam{Changed: true, Value: true},
		ForbidRename: sdk.FileOptionParam{Changed: true, Value: true},
	}, "", key)
	if err == nil {
		clearAllocation(allocationID)
	}
	return hash, err
}

// cancelAllocationMW cancels one of the client's allocations
func cancelAllocationMW(allocationID, key string) (string, error) {
	hash, _, err := sdk.CancelAllocation(allocationID, key)
	if err == nil {
		clearAllocation(allocationID)
	}
	return hash, err
}

// updateAllocationWithRepairMW updates the allocation settings and repairs if necessary
func updateAllocationWithRepairMW(allocationID string,
	size, authRoundExpiry int64,
	extend bool,
	lock int64,
	addBlobberId, addBlobberAuthTicket, removeBlobberId, ownerSigninPublicKey, updateAllocTicket, callbackFuncName, key string) (string, error) {

	sdk.SetWasm()
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return "", err
	}

	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg, isRepair: true, totalBytesMap: make(map[string]int)}
	wg.Add(1)
	if callbackFuncName != "" {
		callback := js.Global().Get(callbackFuncName)
		statusBar.callback = func(totalBytes, completedBytes int, filename, objURL, err string) {
			callback.Invoke(totalBytes, completedBytes, filename, objURL, err)
		}
	}

	alloc, hash, isRepairRequired, err := allocationObj.UpdateWithStatus(size, authRoundExpiry, extend, uint64(lock), addBlobberId, addBlobberAuthTicket, removeBlobberId, ownerSigninPublicKey, false, &sdk.FileOptionsParameters{}, updateAllocTicket)
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

// updateAllocationMW updates the allocation settings
func updateAllocationMW(allocationID string,
	size, authRoundExpiry int64, extend bool,
	lock int64,
	addBlobberId, addBlobberAuthTicket, removeBlobberId, ownerSigninPublicKey string, setThirdPartyExtendable bool, key string) (string, error) {

	hash, _, err := sdk.UpdateAllocation(size, authRoundExpiry, extend, allocationID, uint64(lock), addBlobberId, addBlobberAuthTicket, removeBlobberId, "", ownerSigninPublicKey, setThirdPartyExtendable, &sdk.FileOptionsParameters{}, "", key)
	if err == nil {
		clearAllocation(allocationID)
	}
	return hash, err
}

// getUpdateAllocTicketMW forwards key to SDK's GetUpdateAllocTicket
func getUpdateAllocTicketMW(allocationID, userID, operationType string, roundExpiry int64, key string) (string, error) {
	return sdk.GetUpdateAllocTicket(allocationID, userID, operationType, roundExpiry, key)
}

// getAllocationMW returns allocation for given id using the provided key (if any)
func getAllocationMW(allocationID, key string) (*sdk.Allocation, error) {
	return getAllocation(allocationID, key)
}

// getAllocationMinLockMW retrieves the minimum lock value for allocation creation.
// The `key` parameter is accepted for API symmetry but not forwarded because the
// underlying SDK call does not require a signing key for this computation.
func getAllocationMinLockMW(datashards, parityshards int,
	size int64,
	maxwritePrice uint64,
	key string,
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

// getUpdateAllocationMinLockMW retrieves the minimum lock value required for an allocation update.
// The `key` parameter is accepted for API symmetry but is not used because the SDK
// function does not require a signing key.
func getUpdateAllocationMinLockMW(
	allocationID string,
	size int64,
	extend bool,
	addBlobberId, removeBlobberId, key string) (int64, error) {
	return sdk.GetUpdateAllocationMinLock(allocationID, size, extend, addBlobberId, removeBlobberId)
}

// getRemoteFileMapMW list all files in an allocation from the blobbers.
func getRemoteFileMapMW(allocationID string, key string) ([]*fileResp, error) {
	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}
	allocationObj, err := getAllocation(allocationID, key)
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

// lockWritePoolMW locks given number of tokens for duration in write pool
func lockWritePoolMW(allocID string, tokens, fee uint64, key string) (string, error) {
	hash, _, err := sdk.WritePoolLock(allocID, tokens, fee, key)
	return hash, err
}

// lockStakePoolMW stake number of tokens for a given provider
func lockStakePoolMW(providerType, tokens, fee uint64, providerID string, key string) (string, error) {
	hash, _, err := sdk.StakePoolLock(sdk.ProviderType(providerType), providerID, tokens, fee, key)
	return hash, err
}

// unlockStakePoolMW unlocks stake pool
func unlockStakePoolMW(providerType, fee uint64, providerID, clientID, key string) (int64, error) {
	unstake, _, err := sdk.StakePoolUnlock(sdk.ProviderType(providerType), providerID, clientID, fee, key)
	return unstake, err
}

// collectRewardsMW collects rewards
func collectRewardsMW(providerType int, providerID, key string) (string, error) {
	hash, _, err := sdk.CollectRewards(providerID, sdk.ProviderType(providerType), key)
	return hash, err
}

// getSkatePoolInfoMW gets stake pool info
func getSkatePoolInfoMW(providerType int, providerID, key string) (*sdk.StakePoolInfo, error) {
	info, err := sdk.GetStakePoolInfo(sdk.ProviderType(providerType), providerID, key)
	if err != nil {
		return nil, err
	}
	return info, err
}

// getAllocationWithMW retrieves allocation with auth ticket and optional key
func getAllocationWithMW(authTicket, key string) (*sdk.Allocation, error) {
	sdk.SetWasm()
	sdkAllocation, err := sdk.GetAllocationFromAuthTicket(authTicket, key)
	if err != nil {
		return nil, err
	}
	return sdkAllocation, err
}

// convertTokenToSASMW converts tokens in ZCN to SAS.
func convertTokenToSASMW(token float64, key string) uint64 {
	return uint64(token * float64(TOKEN_UNIT))
}

// allocationRepairMW issue repair process for an allocation
func allocationRepairMW(allocationID, remotePath, key string) error {
	if len(allocationID) == 0 {
		return RequiredArg("allocationID")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return err
	}
	sdk.SetWasm()
	wg := &sync.WaitGroup{}
	statusBar := &StatusBar{wg: wg, isRepair: true, totalBytesMap: make(map[string]int)}
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

// repairSizeMW retrieves the repair size for a specific path in an allocation
func repairSizeMW(allocationID, remotePath, key string) (sdk.RepairSize, error) {
	alloc, err := getAllocation(allocationID, key)
	if err != nil {
		return sdk.RepairSize{}, err
	}
	return alloc.RepairSize(remotePath)
}

// generateOwnerSigningKeyMW generates owner signing key (key is ignored)
func generateOwnerSigningKeyMW(ownerPublicKey, ownerID, key string) (string, error) {
	return sdk.GenerateOwnerSigningPublicKey(key)
}
