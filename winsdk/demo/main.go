package main

// #cgo LDFLAGS: -L. -lzcn.windows
// #include "zcn.windows.h"
import "C"

import (
	"fmt"
	"os"
)

func main() {

	buf, err := os.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}
	configJson := C.CString(string(buf))
	buf, err = os.ReadFile("./client.json")
	if err != nil {
		panic(err)
	}
	clientJson := C.CString(string(buf))
	fmt.Println(clientJson)
	C.InitSDKs(configJson)
	C.InitWallet(clientJson)

	// caution: make sure that this allocation exists in wallet
	allocID := "f4d07362499b2bcfaccbb3a69fdb9642c5b004d08277e04e4359659989d68fd6"

	GetAllocation(allocID)
	UploadFile(allocID)
	CreateFolder(allocID)

}

// GetAllocation gets an allocation
func GetAllocation(allocID string) {
	alloc := C.GetAllocation(C.CString(allocID))
	fmt.Println(C.GoString(alloc))
}

// UploadFile uploads a file to the specified allocation
func UploadFile(allocID string) {
	buf, err := os.ReadFile("./upload.json")
	if err != nil {
		panic(err)
	}

	C.BulkUpload(C.CString(allocID), C.CString(string(buf)))
}

// CreateFolder creates a folder in the specifed allocation
func CreateFolder(allocID string) {
	buf, err := os.ReadFile("./multi.json")
	if err != nil {
		panic(err)
	}

	fmt.Println("started multiOps", allocID, string(buf))
	C.MultiOperation(C.CString(allocID), C.CString(string(buf)))
}
