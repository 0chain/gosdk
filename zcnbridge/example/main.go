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
	ConvertAmountWei = 100
)

// How should we manage nonce? - when user starts again on another server - how should we restore the value?

// Prerequisites:
// 1. Client must have enough amount of Ethereum on his wallet (any Ethereum transaction will fail)
// 2. Client must have enough WZCN tokens in Ethereum chain.

// Order of client initialization

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
	burnTrx, err := zcnbridge.BurnZCN(context.TODO(), config.Bridge.Value)
	burnTrxHash := burnTrx.Hash
	if err != nil {
		log.Logger.Fatal("failed to burn in ZCN", zap.Error(err), zap.String("hash", burnTrxHash))
	}

	// ASK authorizers for burn tickets to mint in Ethereum
	mintPayload, err := authorizer.CreateWZCNMintPayload(burnTrxHash)
	if err != nil {
		log.Logger.Fatal("failed to verify burn transactions in ZCN in CreateWZCNMintPayload", zap.Error(err), zap.String("hash", burnTrxHash))
	}

	tran, err := zcnbridge.MintWZCN(context.Background(), ConvertAmountWei, mintPayload)
	tranHash := tran.Hash().Hex()
	if err != nil {
		log.Logger.Fatal("failed to execute MintWZCN", zap.Error(err), zap.String("hash", tranHash))
	}

	// ASK for minting events from bridge contract but this is not necessary as we're going to check it by hash

	res, err := zcnbridge.ConfirmEthereumTransactionStatus(tranHash, 60, 2)
	if err != nil {
		log.Logger.Fatal(
			"failed to confirm transaction ConfirmEthereumTransactionStatus",
			zap.String("hash", tranHash),
			zap.Error(err),
		)
	}

	if res == 0 {
		log.Logger.Fatal("failed to confirm transaction ConfirmEthereumTransactionStatus", zap.String("hash", tranHash))
	}
}

func fromERCtoZCN() {
	// Example: https://ropsten.etherscan.io/tx/0xa28266fb44cfc2aa27b26bd94e268e40d065a05b1a8e6339865f826557ff9f0e
	transaction, err := zcnbridge.IncreaseBurnerAllowance(context.Background(), ConvertAmountWei)
	if err != nil {
		log.Logger.Fatal("failed to execute IncreaseBurnerAllowance", zap.Error(err))
	}

	hash := transaction.Hash().Hex()
	res, err := zcnbridge.ConfirmEthereumTransactionStatus(hash, 60, 2)
	if err != nil {
		log.Logger.Fatal(
			"failed to confirm transaction ConfirmEthereumTransactionStatus",
			zap.String("hash", hash),
			zap.Error(err),
		)
	}
	if res == 0 {
		log.Logger.Fatal("failed to confirm transaction", zap.String("hash", transaction.Hash().Hex()))
	}

	burnTrx, err := zcnbridge.BurnWZCN(context.Background(), ConvertAmountWei)
	burnTrxHash := burnTrx.Hash().Hex()
	if err != nil {
		log.Logger.Fatal("failed to execute BurnWZCN in wrapped chain", zap.Error(err), zap.String("hash", burnTrxHash))
	}

	res, err = zcnbridge.ConfirmEthereumTransactionStatus(burnTrxHash, 60, 2)
	if err != nil {
		log.Logger.Fatal(
			"failed to confirm transaction ConfirmEthereumTransactionStatus",
			zap.String("hash", burnTrxHash),
			zap.Error(err),
		)
	}
	if res == 0 {
		log.Logger.Fatal("failed to confirm burn transaction in ZCN in ConfirmEthereumTransactionStatus", zap.String("hash", burnTrxHash))
	}

	// ASK authorizers for burn tickets to mint in WZCN
	mintPayload, err := authorizer.CreateZCNMintPayload(burnTrxHash)
	if err != nil {
		log.Logger.Fatal("failed to CreateZCNMintPayload", zap.Error(err), zap.String("hash", burnTrxHash))
	}

	trx, err := zcnbridge.MintZCN(context.TODO(), mintPayload)
	if err != nil {
		log.Logger.Fatal("failed to MintZCN", zap.Error(err), zap.String("hash", trx.Hash))
	}
}
