package main

/*
#include <stdlib.h>
*/

import (
	"C"
)

import (
	"context"
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
	if zboxApiClient == nil {
		return WithJSON(0, ErrZboxApiNotInitialized)
	}

	return WithJSON(zboxApiClient.GetCsrfToken(context.TODO()))
}

// CreateJwtSession create a jwt session
//
//	return
//		{
//			"error":"",
//			"result":"xxx",
//		}
//
//export CreateJwtSession
func CreateJwtSession(phoneNumber *C.char) *C.char {
	if zboxApiClient == nil {
		return WithJSON(0, ErrZboxApiNotInitialized)
	}
	return WithJSON(zboxApiClient.CreateJwtSession(context.TODO(), C.GoString(phoneNumber)))
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
func CreateJwtToken(phoneNumber *C.char, jwtSessionID int64, otp *C.char) *C.char {
	if zboxApiClient == nil {
		return WithJSON("", ErrZboxApiNotInitialized)
	}
	return WithJSON(zboxApiClient.CreateJwtToken(context.TODO(), C.GoString(phoneNumber), jwtSessionID, C.GoString(otp)))
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
func RefreshJwtToken(phoneNumber, token *C.char) *C.char {
	if zboxApiClient == nil {
		return WithJSON("", ErrZboxApiNotInitialized)
	}
	return WithJSON(zboxApiClient.RefreshJwtToken(context.TODO(), C.GoString(phoneNumber), C.GoString(token)))
}
