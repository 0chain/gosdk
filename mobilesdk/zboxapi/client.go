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

// CreateJwtToken creates JWT token with the help of provided userID.
func CreateJwtToken(userID, accessToken string) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.CreateJwtToken(context.TODO(), userID, accessToken)
}

// RefreshJwtToken refreshes JWT token
func RefreshJwtToken(userID string, token string) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.RefreshJwtToken(context.TODO(), userID, token)
}
