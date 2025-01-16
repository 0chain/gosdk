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

// getCsrfToken gets csrf token from 0box api
func getCsrfToken() (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.GetCsrfToken(context.TODO())
}

// createJwtToken creates JWT token with the help of the provided userID.
func createJwtToken(userID, accessToken string) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.CreateJwtToken(context.TODO(), userID, accessToken)
}

// refreshJwtToken refreshes JWT token for the given userID.
func refreshJwtToken(userID string, token string) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.RefreshJwtToken(context.TODO(), userID, token)
}
