package sdk

import (
	"github.com/0chain/gosdk/sdks"
	"github.com/0chain/gosdk/zboxcore/client"
)

var zbox *sdks.ZBox

// initZBox check and create zbox on sdk.InitStorage or zcncore.Init
func initZBox(client *client.Client, signatureScheme string) {
	zbox = sdks.New(client.ClientID, client.ClientKey, signatureScheme)
	zbox.Wallet = *client.Wallet
}
