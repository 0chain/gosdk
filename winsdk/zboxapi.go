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
	"errors"

	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/zboxapi"
	"github.com/0chain/gosdk/zboxcore/client"
)

var (
	zboxApiClient            *zboxapi.Client
	ErrZboxApiNotInitialized = errors.New("zboxapi: zboxapi client is not initialized")
	ErrZboxApiInvalidWallet  = errors.New("zboxapi: invalid wallet")
	logging                  logger.Logger
)

// InitZbox init zbox api client with zbox host and zbox app type
//
//export InitZBox
func InitZBox(zboxHost, zboxAppType *C.char) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	if zboxApiClient == nil {
		zboxApiClient = zboxapi.NewClient()
	}

	zboxApiClient.SetRequest(C.GoString(zboxHost), C.GoString(zboxAppType))

	c := client.GetClient()
	if c != nil {
		zboxApiClient.SetWallet(client.GetClientID(), client.GetClientPrivateKey(), client.GetClientPublicKey())
	} else {
		logging.Info("SetWallet: skipped")
	}
}

// SetZBoxWallet set wallet on zbox api
//
//	return
//		{
//			"error":"",
//			"result":"true",
//		}
//
//export SetZBoxWallet
func SetZBoxWallet(clientID, clientPrivateKey, clientPublicKey *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	if zboxApiClient == nil {
		return WithJSON(false, ErrZboxApiNotInitialized)
	}
	id := C.GoString(clientID)
	prikey := C.GoString(clientPrivateKey)
	pubkey := C.GoString(clientPublicKey)

	if id != "" && prikey != "" && pubkey != "" {
		zboxApiClient.SetWallet(id, prikey, pubkey)
		return WithJSON(true, nil)
	}

	return WithJSON(false, ErrZboxApiInvalidWallet)
}

// GetCsrfToken get a fresh CSRF token
//
//	return
//		{
//			"error":"",
//			"result":"xxx",
//		}
//
//export GetCsrfToken
func GetCsrfToken() *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	if zboxApiClient == nil {
		return WithJSON(0, ErrZboxApiNotInitialized)
	}

	return WithJSON(zboxApiClient.GetCsrfToken(context.TODO()))
}

// CreateJwtToken create a fresh jwt token
//
//	return
//		{
//			"error":"",
//			"result":"xxx",
//		}
//
//export CreateJwtToken
func CreateJwtToken(userID, accessToken *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	if zboxApiClient == nil {
		return WithJSON("", ErrZboxApiNotInitialized)
	}
	return WithJSON(zboxApiClient.CreateJwtToken(context.TODO(), C.GoString(userID), C.GoString(accessToken)))
}

// RefreshJwtToken refresh jwt token
//
//	return
//		{
//			"error":"",
//			"result":"xxx",
//		}
//
//export RefreshJwtToken
func RefreshJwtToken(userID, token *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	if zboxApiClient == nil {
		return WithJSON("", ErrZboxApiNotInitialized)
	}
	return WithJSON(zboxApiClient.RefreshJwtToken(context.TODO(), C.GoString(userID), C.GoString(token)))
}

// GetFreeMarker create a free storage marker
// ## Inputs
//   - phoneNumber
//   - token
//     return
//     {
//     "error":"",
//     "result":"{}",
//     }
//
//export GetFreeMarker
func GetFreeMarker(phoneNumber, token *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	marker, err := zboxApiClient.GetFreeStorage(context.TODO(), C.GoString(phoneNumber), C.GoString(token))
	if err != nil {
		log.Error("win: ", err)
		return WithJSON("", err)
	}

	return WithJSON(marker, nil)
}

// CreateSharedInfo create a shareInfo on 0box db
// ## Inputs
//   - phoneNumber
//   - token
//   - sharedInfo
//
// ## Output
//
//	{
//	"error":"",
//	"result":true,
//	}
//
//export CreateSharedInfo
func CreateSharedInfo(phoneNumber, token, sharedInfo *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	js := C.GoString(sharedInfo)

	s := zboxapi.SharedInfo{}
	err := json.Unmarshal([]byte(js), &s)
	if err != nil {
		log.Error("win: ", js, err)
		return WithJSON(false, err)
	}

	err = zboxApiClient.CreateSharedInfo(context.TODO(), C.GoString(phoneNumber), C.GoString(token), s)
	if err != nil {
		log.Error("win: ", err)
		return WithJSON(false, err)
	}

	return WithJSON(true, nil)
}

// DeleteSharedInfo create a shareInfo on 0box db
// ## Inputs
//   - phoneNumber
//   - token
//   - authTicket
//   - lookupHash
//
// ## Output
//
//	{
//	"error":"",
//	"result":true,
//	}
//
//export DeleteSharedInfo
func DeleteSharedInfo(phoneNumber, token, authTicket, lookupHash *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	err := zboxApiClient.DeleteSharedInfo(context.TODO(), C.GoString(phoneNumber), C.GoString(token), C.GoString(authTicket), C.GoString(lookupHash))
	if err != nil {
		log.Error("win: ", err)
		return WithJSON(false, err)
	}

	return WithJSON(true, nil)
}

// GetSharedByMe get file list that is shared by me privatly
// ## Inputs
//   - phoneNumber
//   - token
//
// ## Output
//
//	{
//	"error":"",
//	"result":[{},{}],
//	}
//
//export GetSharedByMe
func GetSharedByMe(phoneNumber, token *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	list, err := zboxApiClient.GetSharedByMe(context.TODO(), C.GoString(phoneNumber), C.GoString(token))
	if err != nil {
		log.Error("win: ", err)
		return WithJSON(nil, err)
	}

	return WithJSON(list, nil)
}

// GetSharedByPublic get file list that is clicked by me
// ## Inputs
//   - phoneNumber
//   - token
//
// ## Output
//
//	{
//	"error":"",
//	"result":[{},{}],
//	}
//
//export GetSharedByPublic
func GetSharedByPublic(phoneNumber, token *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	list, err := zboxApiClient.GetSharedByPublic(context.TODO(), C.GoString(phoneNumber), C.GoString(token))
	if err != nil {
		log.Error("win: ", err)
		return WithJSON(nil, err)
	}

	return WithJSON(list, nil)
}

// GetSharedToMe get file list that is shared to me
// ## Inputs
//   - phoneNumber
//   - token
//
// ## Output
//
//	{
//	"error":"",
//	"result":[{},{}],
//	}
//
//export GetSharedToMe
func GetSharedToMe(phoneNumber, token *C.char) *C.char {
	defer func() {
		if r := recover(); r != nil {
			log.Error("win: crash ", r)
		}
	}()
	list, err := zboxApiClient.GetSharedToMe(context.TODO(), C.GoString(phoneNumber), C.GoString(token))
	if err != nil {
		log.Error("win: ", err)
		return WithJSON(nil, err)
	}

	return WithJSON(list, nil)
}
