package main

import (
	"context"

	"github.com/0chain/gosdk/zcnbridge"
	"github.com/0chain/gosdk/zcnbridge/config"
	"github.com/0chain/gosdk/zcnbridge/log"
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
	zcnbridge.InitBridge()

	transaction, err := zcnbridge.IncreaseBurnerAllowance(10000000)
	if err != nil {
		log.Logger.Fatal("failed to execute IncreaseBurnerAllowance", zap.Error(err))
	}

	res := zcnbridge.ConfirmTransactionStatus(transaction.Hash().Hex(), 60, 2)
	if res == 0 {
		log.Logger.Fatal("failed to confirm transaction", zap.String("hash", transaction.Hash().Hex()))
	}

	burnTrx, err := zcnbridge.BurnWZCN(10000000)
	trxHash := burnTrx.Hash().Hex()
	if err != nil {
		log.Logger.Fatal("failed to execute BurnWZCN", zap.Error(err), zap.String("hash", trxHash))
	}

	res = zcnbridge.ConfirmTransactionStatus(trxHash, 60, 2)
	if res == 0 {
		log.Logger.Fatal("failed to confirm transaction ConfirmTransactionStatus", zap.String("hash", transaction.Hash().Hex()))
	}

	_, err = zcnbridge.CreateMintPayload(context.TODO(), trxHash)
	if err != nil {
		log.Logger.Fatal("failed to CreateMintPayload", zap.Error(err), zap.String("hash", trxHash))
	}
}
