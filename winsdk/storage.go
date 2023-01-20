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

	fileStats, err := allocationObj.GetFileStats(path)
	if err != nil {
		return WithJSON(nil, err)
	}

	var stats []*sdk.FileStats

	for _, it := range fileStats {
		stats = append(stats, it)
	}

	return WithJSON(stats, nil)
}
