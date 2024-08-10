// Subpackage to provide interface for zboxapi SDK (dealing with apps backend) to be used to build the mobile SDK.
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

	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/zboxapi"
	"github.com/0chain/gosdk/zboxcore/client"
	"go.uber.org/zap"
)

var (
	zboxApiClient            *zboxapi.Client
	ErrZboxApiNotInitialized = errors.New("zboxapi: zboxapi client is not initialized")
	ErrZboxApiInvalidWallet  = errors.New("zboxapi: invalid wallet")
	logging                  logger.Logger
)

// Init initialize the zbox api client for the mobile sdk
//   - baseUrl is the base url of the server
//   - appType is the type of the application
func Init(baseUrl, appType string) {
	zboxApiClient = zboxapi.NewClient()
	zboxApiClient.SetRequest(baseUrl, appType)

	c := client.GetClient()
	if c != nil {
		err := SetWallet(client.GetClientID(), client.GetClientPrivateKey(), client.GetClientPublicKey()) //nolint: errcheck
		if err != nil {
			logging.Error("SetWallet", zap.Error(err))
		}
	} else {
		logging.Info("SetWallet: skipped")
	}
}

// SetWallet set the client's wallet information for the zbox api client
//   - clientID is the client id
//   - clientPrivateKey is the client private key
//   - clientPublicKey is the client public key
func SetWallet(clientID, clientPrivateKey, clientPublicKey string) error {
	if zboxApiClient == nil {
		return ErrZboxApiNotInitialized
	}
	if clientID != "" && clientPrivateKey != "" && clientPublicKey != "" {
		zboxApiClient.SetWallet(clientID, clientPrivateKey, clientPublicKey)
		return nil
	}

	return ErrZboxApiInvalidWallet
}

// GetCsrfToken create a fresh CSRF token
func GetCsrfToken() (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.GetCsrfToken(context.TODO())
}

// CreateJwtSession create a jwt session for the given phone number
//   - phoneNumber is the phone number
func CreateJwtSession(phoneNumber string) (int64, error) {
	if zboxApiClient == nil {
		return 0, ErrZboxApiNotInitialized
	}
	return zboxApiClient.CreateJwtSession(context.TODO(), phoneNumber)
}

// CreateJwtToken create a fresh jwt token for the given phone number
//   - phoneNumber is the phone number
//   - jwtSessionID is the jwt session id
//   - otp is the one time password
func CreateJwtToken(phoneNumber string, jwtSessionID int64, otp string) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.CreateJwtToken(context.TODO(), phoneNumber, jwtSessionID, otp)
}

// RefreshJwtToken refresh jwt token
//   - phoneNumber is the phone number for which the token is to be refreshed
//   - token is the token to be refreshed
func RefreshJwtToken(phoneNumber string, token string) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.RefreshJwtToken(context.TODO(), phoneNumber, token)
}
