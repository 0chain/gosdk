package main

import (
	"github.com/0chain/gosdk/zcnbridge/config"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/zcnbridge/node"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"go.uber.org/zap"
)

func main() {
	// 1. Init config
	// 2. Init logs
	// 2. Init SDK
	// 3. Register wallet
	// 4. Init bridge and make transactions

	config.ParseClientConfig()
	config.Setup()

	client, err := wallet.Setup()
	if err != nil {
		log.Logger.Fatal("failed to setup wallet", zap.Error(err))
	}

	node.Start(client)
}
