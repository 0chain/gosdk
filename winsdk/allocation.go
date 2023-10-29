package main

/*
#include <stdlib.h>
*/
import (
	"C"
)
import (
	"encoding/json"

	"github.com/0chain/gosdk/zboxapi"
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
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
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
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	items, err := sdk.GetAllocations()
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	return WithJSON(items, nil)
}

// CreateFreeAllocation create a free allocation
// ## Inputs
//   - freeMarker
//     return
//     {
//     "error":"",
//     "result":"id",
//     }
//
//export CreateFreeAllocation
func CreateFreeAllocation(freemarker *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	marker := &zboxapi.FreeMarker{}
	js := C.GoString(freemarker)
	err := json.Unmarshal([]byte(js), marker)
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	lock := zcncore.ConvertToValue(marker.FreeTokens)

	allocationID, _, err := sdk.CreateFreeAllocation(js, lock)

	if err != nil {
		log.Error("win: ", err, "lock: ", lock, " marker:", js)
		return WithJSON("", err)
	}

	return WithJSON(allocationID, nil)
}
