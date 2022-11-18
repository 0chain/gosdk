package main

/*
#include <stdlib.h>
#include <stdbool.h>
*/

import (
	"C"
)
import (
	"encoding/json"
	"os"
)

func getHomeDir() string {
	dir, _ := os.UserHomeDir()

	return dir
}

type JsonResult struct {
	Error  string      `json:"error,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

func WithJSON(obj interface{}, err error) *C.char {

	r := &JsonResult{
		Result: obj,
	}
	if err != nil {
		r.Error = err.Error()
	}

	js, _ := json.Marshal(r)

	return C.CString(string(js))
}
