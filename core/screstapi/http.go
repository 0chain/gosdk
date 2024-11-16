package screstapi

import (
	"context"

	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/zboxapi"
)

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
