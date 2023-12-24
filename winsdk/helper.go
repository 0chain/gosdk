package main

/*
#include <stdlib.h>
*/
import (
	"C"
)
import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/0chain/gosdk/core/encryption"
	"github.com/0chain/gosdk/zboxcore/marker"
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

func decodeAuthTicket(authTicket *C.char) (*marker.AuthTicket, string, error) {
	at := C.GoString(authTicket)
	buf, err := base64.StdEncoding.DecodeString(at)
	if err != nil {
		return nil, at, err
	}
	t := &marker.AuthTicket{}
	err = json.Unmarshal(buf, t)
	if err != nil {
		return nil, at, err
	}

	return t, at, err
}
