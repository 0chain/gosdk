//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"errors"

	"github.com/0chain/gosdk/zboxapi"
)

var (
	zboxApiClient            = zboxapi.NewClient()
	ErrZboxApiNotInitialized = errors.New("0box: please call setWallet to create 0box api client")
)

func setZbox(host, appType string) {

}

func getCsrfToken() (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.GetCsrfToken(context.TODO())
}

func createJwtSession(phoneNumber string) (int64, error) {
	if zboxApiClient == nil {
		return 0, ErrZboxApiNotInitialized
	}
	return zboxApiClient.CreateJwtSession(context.TODO(), phoneNumber)
}

func createJwtToken(phoneNumber string, jwtSessionID int64, otp string) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.CreateJwtToken(context.TODO(), phoneNumber, jwtSessionID, otp)
}

func refreshJwtToken(phoneNumber string, token, signature string) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.RefreshJwtToken(context.TODO(), phoneNumber, token, signature)
}
