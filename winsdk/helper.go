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
	"path/filepath"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zboxcore/marker"
	"github.com/0chain/gosdk/zboxcore/sdk"
)

func getZcnWorkDir() (string, error) {
	d, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	z := filepath.Join(d, ".zcn")

	// create ~/.zcn folder if it doesn't exists
	os.MkdirAll(z, 0766) //nolint: errcheck

	return z, nil
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

func getAuthTicket(authTicket *C.char) (*marker.AuthTicket, string, error) {
	at := C.GoString(authTicket)
	t, err := sdk.InitAuthTicket(at).Unmarshall()

	return t, at, err
}
