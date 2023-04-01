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

const (
	migrationDeploymentSH = "https://raw.githubusercontent.com/0chain/blobber/setup-blobber-quickly/docker.local/migration.sh"
	blimpDeploymentSH     = "https://raw.githubusercontent.com/0chain/blobber/setup-blobber-quickly/docker.local/blimp.sh"
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

func refreshJwtToken(phoneNumber string, token string) (string, error) {
	if zboxApiClient == nil {
		return "", ErrZboxApiNotInitialized
	}
	return zboxApiClient.RefreshJwtToken(context.TODO(), phoneNumber, token)
}

func generateBlimpScript(accessKey, secretKey, allocationID, blockWorker, minioToken, domain, walletid, walletpublickey, walletprivatekey string) (string, error) {
	command := "curl -fSsL " +
		blimpDeploymentSH + " | " +
		"sed 's/0chainminiousername/" + accessKey + "/; " +
		"s/0chainminiopassword/" + secretKey + "/; " +
		"s/0chainallocationid/" + allocationID + "/; " +
		"s/0chainblockworker/" + blockWorker + "/; " +
		"s/0chainminiotoken/" + minioToken + "/; " +
		"s/blimpdomain/" + domain + "/;' " +
		"s/0chainwalletid/" + walletid + "/;' " +
		"s/0chainwalletpublickey/" + walletpublickey + "/;' " +
		"s/0chainwalletprivatekey/" + walletprivatekey + "/;' " +
		"| bash"

	return command, nil
}

func generates3MigrationScript(accessKey, secretKey, blockWorker, allocationID, bucket, domain, optionalParams, walletid, walletpublickey, walletprivatekey string) (string, error) {

	command := "curl -fSsL " +
		migrationDeploymentSH + " | " +
		"sed 's/0chainaccesskey/" + accessKey + "/; " +
		"s/0chainsecretkey/" + secretKey + "/; " +
		"s/0chainblockworker/" + blockWorker + "/; " +
		"s/0chainallocation/" + allocationID + "/; " +
		"s/0chainbucket/" + bucket + "/; " +
		optionalParams +
		"s/blimpdomain/" + domain + "/;' " +
		"| bash"

	return command, nil
}
