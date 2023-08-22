package main

/*
#include <stdlib.h>
*/
import (
	"C"
)
import (
	"encoding/json"
	"os"

	"github.com/0chain/gosdk/core/encryption"
)

func getHomeDir() string {
	dir, _ := os.UserHomeDir()

	return dir
}

type JsonResult struct {
	Error  string `json:"error,omitempty"`
	Result string `json:"result,omitempty"`
}

func WithJSON(obj interface{}, err error) *C.char {

	r := &JsonResult{}

	if err != nil {
		r.Error = err.Error()
	}

	if obj != nil {
		result, _ := json.Marshal(obj)
		r.Result = string(result)
	}

	js, _ := json.Marshal(r)

	return C.CString(string(js))
}

func getLookupHash(allocationID, path string) string {
	return encryption.Hash(allocationID + ":" + path)
}
