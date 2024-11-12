package scRestApi

import (
	"context"

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
	"/v1/mint_nonce":                 "/user",
	"/v1/not_processed_burn_tickets": "/not_processed_burn_tickets",
}

func MakeSCRestAPICall(scAddress string, relativePath string, params map[string]string, isWasm bool, restApiUrls ...string) (resp []byte, err error) {
	if isWasm {
		resp, err = MakeSCRestAPICallToZbox(urlPathSharderToZboxMap[relativePath], params)
		if err != nil {
			resp, err = client.MakeSCRestAPICallToSharder(scAddress, relativePath, params)
		}
	} else {
		resp, err = client.MakeSCRestAPICallToSharder(scAddress, relativePath, params)
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
