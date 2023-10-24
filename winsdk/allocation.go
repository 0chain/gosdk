package main

/*
#include <stdlib.h>
*/
import (
	"C"
)
import (
	"context"
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
func CreateFreeAllocation(phoneNumber, token *C.char) *C.char {

	marker, err := zboxApiClient.GetFreeStorage(context.TODO(), C.GoString(phoneNumber), C.GoString(token))
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	fs, err := json.Marshal(marker)
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	lock := zcncore.ConvertToValue(marker.FreeTokens)

	allocationID, _, err := sdk.CreateFreeAllocation(string(fs), lock)

	if err != nil {
		log.Error("win: ", err, "lock: ", lock, " marker:", fs)
		return WithJSON("", err)
	}

	return WithJSON(allocationID, nil)
}
