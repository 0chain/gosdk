package main

import (
	"context"

	"github.com/0chain/gosdk/zcnbridge/authorizer"

	"github.com/0chain/gosdk/zcnbridge"
	"github.com/0chain/gosdk/zcnbridge/config"
	"github.com/0chain/gosdk/zcnbridge/log"
	"go.uber.org/zap"
)

const (
	ConvertAmountWei = 10000000
)

// How should we manage nonce? - when user starts again on another server - how should we restore the value?

// 1. Init config
// 2. Init logs
// 2. Init SDK
// 3. Register wallet
// 4. Init bridge and make transactions

func main() {
	config.ParseClientConfig()
	config.Setup()
	zcnbridge.InitBridge()

	fromERCtoZCN()
	fromZCNtoERC()
}

func fromZCNtoERC() {
	trx, err := zcnbridge.BurnZCN(context.TODO(), config.Bridge.Value)
	if err != nil {
		log.Logger.Fatal("failed to burn", zap.Error(err), zap.String("hash", trx.Hash))
	}

	// ASK authorizers for burn tickets

	tran, err := zcnbridge.MintWZCN(ConvertAmountWei, nil)
	tranHash := tran.Hash().Hex()
	if err != nil {
		log.Logger.Fatal("failed to execute MintWZCN", zap.Error(err), zap.String("hash", tranHash))
	}

	// ASK for minting events from bridge contract

	res := zcnbridge.ConfirmEthereumTransactionStatus(tranHash, 60, 2)
	if res == 0 {
		log.Logger.Fatal("failed to confirm transaction ConfirmEthereumTransactionStatus", zap.String("hash", tranHash))
	}
}

func fromERCtoZCN() {
	transaction, err := zcnbridge.IncreaseBurnerAllowance(ConvertAmountWei)
	if err != nil {
		log.Logger.Fatal("failed to execute IncreaseBurnerAllowance", zap.Error(err))
	}

	res := zcnbridge.ConfirmEthereumTransactionStatus(transaction.Hash().Hex(), 60, 2)
	if res == 0 {
		log.Logger.Fatal("failed to confirm transaction", zap.String("hash", transaction.Hash().Hex()))
	}

	burnTrx, err := zcnbridge.BurnWZCN(ConvertAmountWei)
	burnTrxHash := burnTrx.Hash().Hex()
	if err != nil {
		log.Logger.Fatal("failed to execute BurnWZCN", zap.Error(err), zap.String("hash", burnTrxHash))
	}

	res = zcnbridge.ConfirmEthereumTransactionStatus(burnTrxHash, 60, 2)
	if res == 0 {
		log.Logger.Fatal("failed to confirm transaction ConfirmEthereumTransactionStatus", zap.String("hash", burnTrxHash))
	}

	mintPayload, err := authorizer.CreateZCNMintPayload(burnTrxHash)
	if err != nil {
		log.Logger.Fatal("failed to CreateZCNMintPayload", zap.Error(err), zap.String("hash", burnTrxHash))
	}

	trx, err := zcnbridge.MintZCN(context.TODO(), mintPayload)
	if err != nil {
		log.Logger.Fatal("failed to MintZCN", zap.Error(err), zap.String("hash", trx.Hash))
	}
}
