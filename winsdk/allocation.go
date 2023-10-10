package main

/*
#include <stdlib.h>
*/
import (
	"C"
)
import (
	"encoding/json"

	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/0chain/gosdk/zcncore"
)

// GetAllocation get allocation info
// ## Inputs
//   - allocationID
//
// ## Output
//
//	{
//	"error":"",
//	"result":"{}",
//	}
//
//export GetAllocation
func GetAllocation(allocationID *C.char) *C.char {
	allocID := C.GoString(allocationID)
	return WithJSON(getAllocation(allocID))
}

// ListAllocations get allocation list
// ## Output
//
//	{
//		"error":"",
//		"result":"[{},{}]",
//	}
//
//export ListAllocations
func ListAllocations() *C.char {
	items, err := sdk.GetAllocations()
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	return WithJSON(items, nil)
}

type FreeMarker struct {
	FreeTokens float64 `json:"free_tokens"`
}

// CreateFreeAllocation create a free allocation
// ## Inputs
//   - freeStorageMarker
//     return
//     {
//     "error":"",
//     "result":"id",
//     }
//
//export CreateFreeAllocation
func CreateFreeAllocation(freeStorageMarker *C.char) *C.char {

	marker := &FreeMarker{}
	fs := C.GoString(freeStorageMarker)
	err := json.Unmarshal([]byte(fs), marker)
	if err != nil {
		log.Fatal("unmarshalling marker", err)
	}
	lock := zcncore.ConvertToValue(marker.FreeTokens)

	allocationID, _, err := sdk.CreateFreeAllocation(fs, lock)

	if err != nil {
		log.Error("win: ", err, "lock: ", lock, " marker:", fs)
		return WithJSON("", err)
	}

	return WithJSON(allocationID, nil)
}
