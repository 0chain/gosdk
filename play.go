package main

import (
	"fmt"
	"github.com/0chain/gosdk/zboxcore/sdk"
	//"github.com/0chain/gosdk/mobilesdk/sdk"
	"time"
)

func main() {
	fmt.Printf("Now = %v\n", Now())

	walletString := `{"client_id": "3bb87189fef4971ab475d4649e4f3eae3ad7335e4bca04d57af46e0a23e812f5", "client_key": "b546c07364e8f2470a8d1891d3dcbc799c19f0984a08823880b2d8d3b44ce3127d8fcb88f6b8fa185b47eefcd8aa5dd9f93335243ac3d0bc4e7641c8e9807d13", "keys": [{"public_key": "b546c07364e8f2470a8d1891d3dcbc799c19f0984a08823880b2d8d3b44ce3127d8fcb88f6b8fa185b47eefcd8aa5dd9f93335243ac3d0bc4e7641c8e9807d13", "private_key": "59c9a5900433edd40ef40eb03e9064e6e5ec67d480aeb2cfee290ab37dc1c01e"}], "mnemonics": "mom ghost camp hole story flight off waste brave inflict throw pen ostrich universe fetch verb tomato earn mixture flat notice pizza merge offer", "version": "1.0", "date_created": "2023-11-26T01:23:58Z", "nonce": 1, "ChainID": "0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe", "SignatureScheme": null}`
	blockWorker := "https://dev-st.devnet-0chain.net/dns"
	chainId := "0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe"

	err := sdk.InitStorageSDK(walletString, blockWorker, chainId, "bls0chain", nil, 0)

	if err != nil {
		fmt.Println(err)
		return
	}

	updateBlobberModel := &sdk.UpdateBlobber{
		ID: "invaild_id",
	}
	res, nonce, err := sdk.UpdateBlobberSettings(updateBlobberModel)

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("UpdateBlobberSettings response : ", res)
		fmt.Println("UpdateBlobberSettings nonce : ", nonce)
	}

}

// Timestamp represents Unix time (e.g. in seconds)
type Timestamp int64

// Now - current datetime
func Now() Timestamp {
	return Timestamp(time.Now().Unix())
}
