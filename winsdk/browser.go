package main

/*
#include <stdlib.h>
*/
import (
	"C"
)

import (
	"errors"
	"strings"

	"github.com/0chain/gosdk/zboxcore/sdk"
)

type RemoteFile struct {
	sdk.FileInfo
	Name string `json:"name"`
	Path string `json:"path"`
}

// ListAll - list all files from blobbers
// ## Inputs
//   - allocationID
//
// ## Outputs
//
//	{
//	"error":"",
//	"result":[{},{}]",
//	}
//
//export ListAll
func ListAll(allocationID *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	alloc, err := getAllocation(C.GoString(allocationID))
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	ref, err := alloc.GetRemoteFileMap(nil, "/")
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	files := make([]RemoteFile, 0)
	for path, data := range ref {
		paths := strings.SplitAfter(path, "/")
		var f = RemoteFile{
			Name:     paths[len(paths)-1],
			Path:     path,
			FileInfo: data,
		}
		files = append(files, f)
	}

	return WithJSON(files, nil)
}

// List - list files from blobbers
// ## Inputs
//   - allocationID
//   - remotePath
//
// - authTicket
// - lookupHash
//
// ## Outputs
//
//	{
//	"error":"",
//	"result":[{},{}]",
//	}
//
//export List
func List(allocationID, remotePath, authTicket, lookupHash *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	allocID := C.GoString(allocationID)
	remotepath := C.GoString(remotePath)
	authticket := C.GoString(authTicket)
	lookuphash := C.GoString(lookupHash)

	if len(remotepath) == 0 && len(authticket) == 0 {
		return WithJSON("[]", errors.New("Error: remotepath / authticket flag is missing"))
	}

	if len(remotepath) > 0 {
		if len(allocID) == 0 {
			return WithJSON("[]", errors.New("Error: allocationID is missing"))
		}

		allocationObj, err := getAllocation(allocID)
		if err != nil {
			log.Error("win: ", err)
			return WithJSON("[]", err)
		}

		ref, err := allocationObj.ListDir(remotepath)
		if err != nil {
			if err != nil {
				log.Error("win: ", err)
				return WithJSON("[]", err)
			}
		}

		return WithJSON(ref.Children, nil)
	}
	if len(authticket) > 0 {

		if len(lookuphash) == 0 {
			return WithJSON("[]", errors.New("Error: lookuphash flag is missing"))
		}

		allocationObj, _, err := getAllocationWith(authticket)
		if err != nil {
			log.Error("win: ", err)
			return WithJSON("[]", err)
		}

		at := sdk.InitAuthTicket(authticket)
		lookuphash, err = at.GetLookupHash()
		if err != nil {
			log.Error("win: ", err)
			return WithJSON("[]", err)
		}

		ref, err := allocationObj.ListDirFromAuthTicket(authticket, lookuphash)
		if err != nil {
			log.Error("win: ", err)
			return WithJSON("[]", err)
		}

		return WithJSON(ref.Children, nil)
	}
	return WithJSON("[]", nil)
}
