package screstapi

import (
	"context"
	"encoding/json"

	"github.com/0chain/gosdk/core/client"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/zboxapi"
)

var urlPathSharderToZboxMap = map[string]string{
	"/getStakePoolStat":              "/getStakePoolStat",
	"/getUserStakePoolStat":          "/getUserStakePoolStat",
	"/getChallengePoolStat":          "/getChallengePoolStat",
	"/getBlobber":                    "/blobber",
	"/getblobbers":                   "/blobbers",
	"/blobber_ids":                   "/blobber_ids",
	"/alloc_blobbers":                "/blobbers/allocation",
	"/get_validator":                 "/validator",
	"/validators":                    "/validators",
	"/allocation":                    "/getAllocation",
	"/allocations":                   "/getAllocations",
	"/v1/mint_nonce":                 "/mintNonce",
	"client/get/balance":             "/balance",
	"/v1/not_processed_burn_tickets": "/not_processed_burn_tickets",
}

func MakeSCRestAPICall(scAddress string, relativePath string, params map[string]string, restApiUrls ...string) (resp []byte, err error) {
	_, ok := urlPathSharderToZboxMap[relativePath]
	if client.IsSdkFlow && ok {
		resp, err = MakeSCRestAPICallToZbox(urlPathSharderToZboxMap[relativePath], params)
		if err != nil {
			resp, err = client.MakeSCRestAPICallToSharder(scAddress, relativePath, params)
		}
	} else {
		resp, err = client.MakeSCRestAPICallToSharder(scAddress, relativePath, params, restApiUrls...)
	}

	return resp, err
}

func MakeSCRestAPICallToZbox(relativePath string, params map[string]string) ([]byte, error) {
	// req, err := http.NewRequest(method, relativePath)
	zboxApiClient := zboxapi.NewClient()
	configObj := &conf.Config{}
	zboxApiClient.SetRequest(configObj.ZboxHost, configObj.ZboxAppType)

	resp, err := zboxApiClient.MakeRestApiCallToZbox(context.TODO(), relativePath, params)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GetBalance(clientIDs ...string) (*client.GetBalanceResponse, error) {
	var clientID string
	if len(clientIDs) > 0 {
		clientID = clientIDs[0]
	} else {
		clientID = client.Id()
	}

	var (
		balance client.GetBalanceResponse
		err     error
		resp    []byte
	)

	if resp, err = MakeSCRestAPICall("", client.GetBalanceUrl, map[string]string{
		"client_id": clientID,
	}, "v1/"); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(resp, &balance); err != nil {
		return nil, err
	}

	return &balance, err
}
