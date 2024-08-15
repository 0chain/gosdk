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

// createJwtSession creates jwt session for the given phone number
//   - phoneNumber is the phone number of the user
func createJwtSession(userID string) (int64, error) {
	if zboxApiClient == nil {
		return 0, ErrZboxApiNotInitialized
	}
	return zboxApiClient.CreateJwtSession(context.TODO(), userID)
}

// createJwtToken creates jwt token for the given phone number
//   - phoneNumber is the phone number of the user
//   - jwtSessionID is the jwt session id
//   - otp is the one time password
func createJwtToken(userID string, jwtSessionID int64) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.CreateJwtToken(context.TODO(), userID, jwtSessionID)
}

// refreshJwtToken refreshes jwt token for the given phone number
//   - phoneNumber is the phone number of the user
//   - token is the jwt token to refresh
func refreshJwtToken(userID string, token string) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.RefreshJwtToken(context.TODO(), userID, token)
}
