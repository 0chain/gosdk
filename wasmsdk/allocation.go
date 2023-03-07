//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/0chain/gosdk/core/transaction"
	"github.com/0chain/gosdk/core/util"
	"github.com/0chain/gosdk/zboxcore/client"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

const TOKEN_UNIT int64 = 1e10

type fileResp struct {
	sdk.FileInfo
	Name string `json:"name"`
	Path string `json:"path"`
}

func getBlobberIds(blobberUrls []string) ([]string, error) {
	return sdk.GetBlobberIds(blobberUrls)
}

func getAllocationBlobbers(preferredBlobberURLs []string,
	dataShards, parityShards int, size, expiry int64,
	minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64) ([]string, error) {

	if len(preferredBlobberURLs) > 0 {
		return sdk.GetBlobberIds(preferredBlobberURLs)
	}

	return sdk.GetAllocationBlobbers(dataShards, parityShards, size, expiry, sdk.PriceRange{
		Min: uint64(minReadPrice),
		Max: uint64(maxReadPrice),
	}, sdk.PriceRange{
		Min: uint64(minWritePrice),
		Max: uint64(maxWritePrice),
	})
}

func createAllocation(datashards, parityshards int, size, expiry int64,
	minReadPrice, maxReadPrice, minWritePrice, maxWritePrice int64, lock int64, blobberIds []string) (
	*transaction.Transaction, error) {

	options := sdk.CreateAllocationOptions{
		DataShards:   datashards,
		ParityShards: parityshards,
		Size:         size,
		Expiry:       expiry,
		ReadPrice: sdk.PriceRange{
			Min: uint64(minReadPrice),
			Max: uint64(maxReadPrice),
		},
		WritePrice: sdk.PriceRange{
			Min: uint64(minWritePrice),
			Max: uint64(maxWritePrice),
		},
		Lock:       uint64(lock),
		BlobberIds: blobberIds,
	}

	sdkLogger.Info(options)
	_, _, txn, err := sdk.CreateAllocationWith(options)

	return txn, err
}

func listAllocations() ([]*sdk.Allocation, error) {
	return sdk.GetAllocations()
}

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

	_, _, err := sdk.CuratorTransferAllocation(allocationID, newOwnerId, newOwnerPublicKey)

	if err == nil {
		clearAllocation(allocationID)
	}

	return err
}

func freezeAllocation(allocationID string) (string, error) {

	hash, _, err := sdk.UpdateAllocation(
		0,            //size,
		0,            //int64(expiry/time.Second),
		allocationID, // allocID,
		0,            //lock,
		false,        //updateTerms,
		"",           //addBlobberId,
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

func cancelAllocation(allocationID string) (string, error) {
	hash, _, err := sdk.CancelAllocation(allocationID)

	if err == nil {
		clearAllocation(allocationID)
	}

	return hash, err
}

func updateAllocation(allocationID string,
	size, expiry int64,
	lock int64,
	updateTerms bool,
	addBlobberId, removeBlobberId string) (string, error) {
	hash, _, err := sdk.UpdateAllocation(size, expiry, allocationID, uint64(lock), updateTerms, addBlobberId, removeBlobberId, false, &sdk.FileOptionsParameters{})

	if err == nil {
		clearAllocation(allocationID)
	}

	return hash, err
}

func getAllocationMinLock(datashards, parityshards int,
	size, expiry int64,
	maxreadPrice, maxwritePrice uint64,
) (int64, error) {
	readPrice := sdk.PriceRange{Min: 0, Max: maxreadPrice}
	writePrice := sdk.PriceRange{Min: 0, Max: maxwritePrice}

	value, err := sdk.GetAllocationMinLock(datashards, parityshards, size, expiry, readPrice, writePrice)
	if err != nil {
		sdkLogger.Error(err)
		return 0, err
	}
	sdkLogger.Info("allocation Minlock value", value)
	return value, nil
}

func getRemoteFileMap(allocationID string) ([]*fileResp, error) {
	if len(allocationID) == 0 {
		return nil, RequiredArg("allocationID")
	}
	allocationObj, err := sdk.GetAllocation(allocationID)
	if err != nil {
		return nil, err
	}

	ref, err := allocationObj.GetRemoteFileMap(nil)
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
// ## Inputs
//   - allocID: allocation id
//   - tokens:  sas tokens
//   - fee: sas tokens
func lockWritePool(allocID, tokens, fee string) (string, error) {
	t, err := util.ParseCoinStr(tokens)
	if err != nil {
		return "", err
	}

	f, err := util.ParseCoinStr(fee)
	if err != nil {
		return "", err
	}
	hash, _, err := sdk.WritePoolLock(allocID, t, f)
	return hash, err
}

// GetReadPoolInfo is to get information about the read pool for the allocation
func getReadPoolInfo(clientID string) (*sdk.ReadPool, error) {
	readPool, err := sdk.GetReadPoolInfo(clientID)
	if err != nil {
		return nil, err
	}

	return readPool, nil
}

// GetAllocationFromAuthTicket - get allocation from Auth ticket
func getAllocationFromAuthTicket(authTicket string) (*sdk.Allocation, error) {
	sdkAllocation, err := sdk.GetAllocationFromAuthTicket(authTicket)
	if err != nil {
		return nil, err
	}
	return sdkAllocation, err
}

func redeemFreeStorage(ticket string) (string, error) {
	recipientPublicKey, marker, lock, err := decodeAuthTicket(ticket)
	if err != nil {
		return "", err
	}

	if recipientPublicKey != client.GetClientPublicKey() {
		return "", fmt.Errorf("invalid_free_marker: free marker is not assigned to your wallet")
	}

	hash, _, err := sdk.CreateFreeAllocation(marker, lock)
	return hash, err
}

func decodeAuthTicket(ticket string) (string, string, uint64, error) {
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
	return string(recipientPublicKey), string(markerStr), convertTokenToSAS(s), nil
}

func convertTokenToSAS(token float64) uint64 {
	return uint64(token * float64(TOKEN_UNIT))
}
