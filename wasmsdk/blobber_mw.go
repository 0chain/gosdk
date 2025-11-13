//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"errors"

	"github.com/0chain/gosdk/constants"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

// NOTE: This file provides *MW suffixed* wrappers for several wasm blobber functions
// that accept an additional `key string` parameter (the MultiWalletSupportKey).
// Where the underlying SDK exposes a way to use a per-operation key, the wrapper
// forwards it. Where the SDK does not currently accept a key for that operation,
// the wrapper will either call the existing function (and ignore the key) or
// return an error indicating the key cannot be applied.

// listObjectsMW lists allocation objects but signs the list request using `key` when provided.
func listObjectsMW(allocationID string, remotePath string, offset, pageLimit int, key string) (*sdk.ListResult, error) {
	alloc, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}
	if key != "" {
		return alloc.ListDir(remotePath, sdk.WithListRequestOffset(offset), sdk.WithListRequestPageLimit(pageLimit), sdk.WithListRequestPubKey(key))
	}
	return alloc.ListDir(remotePath, sdk.WithListRequestOffset(offset), sdk.WithListRequestPageLimit(pageLimit))
}

// listObjectsFromAuthTicketMW lists allocation objects using an auth ticket and an optional signing key.
func listObjectsFromAuthTicketMW(allocationID, authTicket, lookupHash string, offset, pageLimit int, key string) (*sdk.ListResult, error) {
	alloc, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}
	if key != "" {
		return alloc.ListDirFromAuthTicket(authTicket, lookupHash, sdk.WithListRequestOffset(offset), sdk.WithListRequestPageLimit(pageLimit), sdk.WithListRequestPubKey(key))
	}
	return alloc.ListDirFromAuthTicket(authTicket, lookupHash, sdk.WithListRequestOffset(offset), sdk.WithListRequestPageLimit(pageLimit))
}

// createDirMW creates a directory and uses the provided key for signing the multi-operation.
func createDirMW(allocationID, remotePath, key string) error {
	if allocationID == "" {
		return errors.New("allocationID required")
	}
	if remotePath == "" {
		return errors.New("remotePath required")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return err
	}
	// Use DoMultiOperation and set the MultiWalletSupportKey on the MultiOperation
	return allocationObj.DoMultiOperation([]sdk.OperationRequest{{
		OperationType: constants.FileOperationCreateDir,
		RemotePath:    remotePath,
	}}, func(mo *sdk.MultiOperation) { mo.MultiWalletSupportKey = key })
}

// DeleteMW deletes a file using the provided MultiWalletSupportKey.
func DeleteMW(allocationID, remotePath, key string) (*FileCommandResponse, error) {
	if allocationID == "" {
		return nil, RequiredArg("allocationID")
	}
	if remotePath == "" {
		return nil, RequiredArg("remotePath")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}

	err = allocationObj.DoMultiOperation([]sdk.OperationRequest{{
		OperationType: constants.FileOperationDelete,
		RemotePath:    remotePath,
	}}, func(mo *sdk.MultiOperation) { mo.MultiWalletSupportKey = key })
	if err != nil {
		return nil, err
	}
	resp := &FileCommandResponse{CommandSuccess: true}
	return resp, nil
}

// RenameMW renames a file using the provided key for signing.
func RenameMW(allocationID, remotePath, destName, key string) (*FileCommandResponse, error) {
	if allocationID == "" {
		return nil, RequiredArg("allocationID")
	}
	if remotePath == "" {
		return nil, RequiredArg("remotePath")
	}
	if destName == "" {
		return nil, RequiredArg("destName")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}

	err = allocationObj.DoMultiOperation([]sdk.OperationRequest{{
		OperationType: constants.FileOperationRename,
		RemotePath:    remotePath,
		DestName:      destName,
	}}, func(mo *sdk.MultiOperation) { mo.MultiWalletSupportKey = key })
	if err != nil {
		return nil, err
	}
	resp := &FileCommandResponse{CommandSuccess: true}
	return resp, nil
}

// CopyMW copies a file and signs using the provided key.
func CopyMW(allocationID, remotePath, destPath, key string) (*FileCommandResponse, error) {
	if allocationID == "" {
		return nil, RequiredArg("allocationID")
	}
	if remotePath == "" {
		return nil, RequiredArg("remotePath")
	}
	if destPath == "" {
		return nil, RequiredArg("destPath")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}

	err = allocationObj.DoMultiOperation([]sdk.OperationRequest{{
		OperationType: constants.FileOperationCopy,
		RemotePath:    remotePath,
		DestPath:      destPath,
	}}, func(mo *sdk.MultiOperation) { mo.MultiWalletSupportKey = key })
	if err != nil {
		return nil, err
	}
	resp := &FileCommandResponse{CommandSuccess: true}
	return resp, nil
}

// MoveMW moves a file and signs using the provided key.
func MoveMW(allocationID, remotePath, destPath, key string) (*FileCommandResponse, error) {
	if allocationID == "" {
		return nil, RequiredArg("allocationID")
	}
	if remotePath == "" {
		return nil, RequiredArg("remotePath")
	}
	if destPath == "" {
		return nil, RequiredArg("destPath")
	}
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return nil, err
	}

	err = allocationObj.DoMultiOperation([]sdk.OperationRequest{{
		OperationType: constants.FileOperationMove,
		RemotePath:    remotePath,
		DestPath:      destPath,
	}}, func(mo *sdk.MultiOperation) { mo.MultiWalletSupportKey = key })
	if err != nil {
		return nil, err
	}
	resp := &FileCommandResponse{CommandSuccess: true}
	return resp, nil
}

// MultiOperationMW performs multiple operations and signs the whole multi-operation with `key`.
func MultiOperationMW(allocationID string, jsonMultiUploadOptions string, key string) error {
	if allocationID == "" {
		return errors.New("AllocationID is required")
	}
	if jsonMultiUploadOptions == "" {
		return errors.New("operations are empty")
	}
	var options []MultiOperationOption
	if err := json.Unmarshal([]byte(jsonMultiUploadOptions), &options); err != nil {
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
	allocationObj, err := getAllocation(allocationID, key)
	if err != nil {
		return err
	}
	return allocationObj.DoMultiOperation(operations, func(mo *sdk.MultiOperation) { mo.MultiWalletSupportKey = key })
}

// For other functions in blobber.go that are primarily read-only or don't expose a
// way to set a per-operation signing key in the SDK (for example: GetFileStats,
// GetBlobbers, downloadBlocks, upload wrappers that don't accept a wallet option),
// the SDK currently does not provide a simple way to use the MultiWalletSupportKey.
// Such functions are not implemented here with key-aware behavior and will
// effectively ignore the `key` parameter if a MW wrapper were added that calls
// the existing API.

// TODO: If you want, I can extend this file with additional MW wrappers (for
// example uploadMW that will use sdk.CreateChunkedUpload with sdk.WithWallet(key)),
// or implement key-aware variants for more of the blobber API. If so, tell me
// which specific functions you'd like prioritized.
