package zboxapi

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
	"github.com/0chain/gosdk/zboxcore/client"
)

var (
	zboxApiClient            *zboxapi.Client
	ErrZboxApiNotInitialized = errors.New("zboxapi: zboxapi client is not initialized")
	ErrZboxApiInvalidWallet  = errors.New("zboxapi: invalid wallet")
)

func Init(baseUrl, appType string) {
	zboxApiClient = zboxapi.NewClient(baseUrl, appType)
	SetWallet(client.GetClientID(), client.GetClientPrivateKey(), client.GetClientPublicKey()) //nolint: errcheck
}

func SetWallet(clientID, clientPrivateKey, clientPublicKey string) error {
	if zboxApiClient == nil {
		return ErrZboxApiNotInitialized
	}
	if clientID != "" && clientPrivateKey != "" && clientPublicKey != "" {
		zboxApiClient.SetWallet(clientID, clientPrivateKey, clientPublicKey)
		return nil
	}

	return nil
}

// GetCsrfToken create a fresh CSRF token
func GetCsrfToken() (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.GetCsrfToken(context.TODO())
}

// CreateJwtSession create a jwt session
func CreateJwtSession(phoneNumber string) (int64, error) {
	if zboxApiClient == nil {
		return 0, ErrZboxApiNotInitialized
	}
	return zboxApiClient.CreateJwtSession(context.TODO(), phoneNumber)
}

// CreateJwtToken create a fresh jwt token
func CreateJwtToken(phoneNumber string, jwtSessionID int64, otp string) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.CreateJwtToken(context.TODO(), phoneNumber, jwtSessionID, otp)
}

// RefreshJwtToken refresh jwt token
func RefreshJwtToken(phoneNumber string, token string) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.RefreshJwtToken(context.TODO(), phoneNumber, token)
}
