package main

import (
	"fmt"
	"github.com/0chain/gosdk/mobilesdk/sdk"
	"github.com/0chain/gosdk/zcncore"
	"os"
)

var (
	chainConfig = `{
      "chain_id": "0afc093ffb509f059c55478bc1a60351cef7b4e9c008a53a6cc8241ca8617dfe",
      "signature_scheme": "bls0chain",
      "block_worker": "https://mob.zus.network/dns",
      "min_submit": 20,
      "min_confirmation": 10,
      "confirmation_chain_length": 3,
      "num_keys": 1,
      "eth_node": "https://ropsten.infura.io/v3/f0a254d8d18b4749bd8540da63b3292b"
    }`

	walletJson = `{
  "client_id": "aa7ff1bb8b81599b00068c332278a5c35647dacf84d4bc6f0cdabcbf68efffe9",
  "client_key": "2713241d62d29cba9c785a9ed54dc89e4841ab19004b7b4fbae2eb6c79993908a9525d19de501401ce96a5af0a8d71e02603e5b9f2588a992d1c2ffca589170c",
  "peer_public_key": "",
  "keys": [
    {
      "public_key": "2713241d62d29cba9c785a9ed54dc89e4841ab19004b7b4fbae2eb6c79993908a9525d19de501401ce96a5af0a8d71e02603e5b9f2588a992d1c2ffca589170c",
      "private_key": "30d18bf2ea2f371b38c6822f70e2b4830de4ab81128344b90e97b9a874826510"
    }
  ],
  "mnemonics": "stable fall surround sort help pond furnace catch shallow knife update all notable together vocal shield jeans nature achieve game ladder swim wing rhythm",
  "version": "1.0",
  "date_created": "2024-10-11T21:31:34Z",
  "nonce": 0,
  "is_split": false
}`

	clientId = "aa7ff1bb8b81599b00068c332278a5c35647dacf84d4bc6f0cdabcbf68efffe9"
)

func initBase() {
	fmt.Println("chain config is ", chainConfig)

	err := sdk.Init(chainConfig)

	if err != nil {
		fmt.Println("error initializing sdk ", err)
		return
	}
}

func setWalletInfo() {
	err := zcncore.SetWalletInfo(walletJson, "bls0chain", false)

	if err != nil {
		fmt.Println("error setting wallet info ", err)
		return
	}
}

func initSDK() {
	fmt.Println("wallet json is ", walletJson)
	fmt.Println("Chain config is ", chainConfig)

	_, err := sdk.InitStorageSDK(walletJson, chainConfig)
	if err != nil {
		fmt.Println("error initializing sdk ", err)
		return
	}

}

func main() {
	initBase()

	setWalletInfo()

	initSDK()

	testGetTransactions()
	testGetStakePoolUserInfo()
	testGetBlobbers()
	testThumbnailGeneration()
}

func testGetBlobbers() {
	res, err := zcncore.GetBlobbers(true, true, 20, 0)

	if err != nil {
		fmt.Println("error getting blobbers ", err)
		return
	}

	fmt.Println("blobbers ", string(res))
}

func testGetTransactions() {
	res, err := zcncore.GetTransactions("", clientId, "DESC", 20, 0)

	if err != nil {
		fmt.Println("error getting transactions ", err)
		return
	}

	fmt.Println("transactions ", string(res))
}

func testGetStakePoolUserInfo() {
	res, err := zcncore.GetStakePoolUserInfo(clientId)

	if err != nil {
		fmt.Println("error getting stake pool user info ", err)
		return
	}

	fmt.Println("stake pool user info ", string(res))
}

func testThumbnailGeneration() {
	//Get a image from this path
	imagePath := "/home/ubuntu/Downloads/0chain/Slack/image (4).png"
	file, err := os.ReadFile(imagePath)
	if err != nil {
		fmt.Println("error reading file ", err)
		return
	}

	res, err := sdk.CreateThumbnail(file, 100, 100)

	if err != nil {
		fmt.Println("error creating thumbnail ", err)
	}

	err = os.WriteFile("/home/ubuntu/Pictures/thumbnail.png", res, 0777)
	if err != nil {
		fmt.Println("error writing thumbnail ", err)
	}
}
