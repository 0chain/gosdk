package main

import (
	"github.com/0chain/gosdk/zcnbridge"
	"github.com/0chain/gosdk/zcnbridge/config"
)

func main() {
	// 1. Init config
	// 2. Init logs
	// 2. Init SDK
	// 3. Register wallet
	// 4. Init bridge and make transactions

	config.ParseClientConfig()
	config.Setup()
	zcnbridge.InitBridge()
}
