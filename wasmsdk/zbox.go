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

func createJwtSession(userID string) (int64, error) {
	if zboxApiClient == nil {
		return 0, ErrZboxApiNotInitialized
	}
	return zboxApiClient.CreateJwtSession(context.TODO(), userID)
}

func createJwtToken(userID string, jwtSessionID int64) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.CreateJwtToken(context.TODO(), userID, jwtSessionID)
}

func refreshJwtToken(userID string, token string) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.RefreshJwtToken(context.TODO(), userID, token)
}
