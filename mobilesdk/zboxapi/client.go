package zboxapi

import (
	"github.com/0chain/gosdk/zboxapi"
	"github.com/0chain/gosdk/zboxcore/client"
)

var zboxApiClient *zboxapi.Client

func InitZboxApi(baseUrl, appType string) {
	zboxApiClient = zboxapi.NewClient(baseUrl, appType, client.GetClientID(), client.GetClientPrivateKey(), client.GetClientPublicKey())
}
