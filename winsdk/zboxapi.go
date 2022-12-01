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

	"github.com/0chain/gosdk/zboxapi"
)

var (
	zboxApiClient            *zboxapi.Client
	ErrZboxApiNotInitialized = errors.New("0box: zboxapi client is not initialized")
)

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
func CreateJwtSession(phoneNumber string) *C.char {
	if zboxApiClient == nil {
		return WithJSON(0, ErrZboxApiNotInitialized)
	}
	return WithJSON(zboxApiClient.CreateJwtSession(context.TODO(), phoneNumber))
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
func CreateJwtToken(phoneNumber string, jwtSessionID int64, otp string) *C.char {
	if zboxApiClient == nil {
		return WithJSON("", ErrZboxApiNotInitialized)
	}
	return WithJSON(zboxApiClient.CreateJwtToken(context.TODO(), phoneNumber, jwtSessionID, otp))
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
func RefreshJwtToken(phoneNumber string, token string) *C.char {
	if zboxApiClient == nil {
		return WithJSON("", ErrZboxApiNotInitialized)
	}
	return WithJSON(zboxApiClient.RefreshJwtToken(context.TODO(), phoneNumber, token))
}
