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
)

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

// ListAllocations get allocation list
//
//	return
//		{
//			"error":"",
//			"result":"[{},{}]",
//		}
//
//export ListAllocations
func ListAllocations() *C.char {
	items, err := sdk.GetAllocations()
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	js, err := json.Marshal(items)
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	return WithJSON(string(js), nil)
}
