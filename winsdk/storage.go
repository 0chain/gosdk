package main

/*
#include <stdlib.h>
*/

import (
	"C"
)

import (
	"errors"

	"github.com/0chain/gosdk/zboxcore/sdk"
)

// GetFileStats get file stats of blobbers
//
//	return
//		{
//			"error":"",
//			"result":"xxx",
//		}
//
//export GetFileStats
func GetFileStats(allocationID, remotePath *C.char) *C.char {
	allocID := C.GoString(allocationID)
	path := C.GoString(remotePath)

	if len(allocID) == 0 {
		return WithJSON(nil, errors.New("allocationID is required"))
	}

	if len(path) == 0 {
		return WithJSON(nil, errors.New("remotePath is required"))
	}

	allocationObj, err := getAllocation(allocID)
	if err != nil {
		return WithJSON(nil, err)
	}

	stats, err := allocationObj.GetFileStats(path)
	if err != nil {
		return WithJSON(nil, err)
	}

	result := make([]*sdk.FileStats, 0, len(stats))

	//convert map[string]*sdk.FileStats to []*sdk.FileStats
	for _, v := range stats {
		result = append(result, v)
	}

	return WithJSON(result, nil)
}

// GetAllocation get allocation info
//
//	return
//		{
//			"error":"",
//			"result":"{}",
//		}
//
//export GetAllocation
func GetAllocation(allocationID *C.char) *C.char {
	allocID := C.GoString(allocationID)
	return WithJSON(getAllocation(allocID))
}

// Rename rename path
//
//	return
//		{
//			"error":"",
//			"result":"true",
//		}
//
//export Rename
func Rename(allocationID, path, destName *C.char) *C.char {
	allocID := C.GoString(allocationID)

	alloc, err := getAllocation(allocID)
	if err != nil {
		return WithJSON(false, err)
	}

	s := C.GoString(path)
	d := C.GoString(destName)
	err = alloc.RenameObject(s, d)

	if err != nil {
		return WithJSON(false, err)
	}

	return WithJSON(true, nil)

}

// Delete delete path
//
//	return
//		{
//			"error":"",
//			"result":"true",
//		}
//
//export Delete
func Delete(allocationID, path *C.char) *C.char {
	allocID := C.GoString(allocationID)

	alloc, err := getAllocation(allocID)
	if err != nil {
		return WithJSON(false, err)
	}

	s := C.GoString(path)

	err = alloc.DeleteFile(s)

	if err != nil {
		return WithJSON(false, err)
	}

	return WithJSON(true, nil)
}

// GetFileMeta get metadata by path
//
//	return
//		{
//			"error":"",
//			"result":"true",
//		}
//
//export GetFileMeta
func GetFileMeta(allocationID, path *C.char) *C.char {
	allocID := C.GoString(allocationID)

	alloc, err := getAllocation(allocID)
	if err != nil {
		return WithJSON(false, err)
	}

	s := C.GoString(path)

	f, err := alloc.GetFileMeta(s)

	if err != nil {
		return WithJSON(false, err)
	}

	return WithJSON(f, nil)
}
