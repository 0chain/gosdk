package main

import (
	"context"
	"fmt"

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

	// To test this, authorizers must be installed
	PrintAuthorizers()
	PrintEthereumBurnTicketsPayloads()

	fromERCtoZCN()
	fromZCNtoERC()
}

func PrintEthereumBurnTicketsPayloads() {
	// Ropsten burn successful transactions for which we may receive burn tickets and mint payloads
	// to mint ZCN tokens
	tranHashes := []string{
		"0xa5049192c3622534e6195fbadcf21c9eb928ca3e5e8c7056f500f78f31c1c1aa",
		"0xd3583513ea4f76f25000e704c8fc12c5b7b71a1574138d4df20d948255bd7f9c",
		"0x468805e8bb268d584659ccd104e36bd5e552feec440d1a761aa8f9034a92b2fd",
		"0x39ba7befd88a6dc6abec1bd503a6c2ced9472b8643704e4048d673728fb373b5",
		"0x31925839586949a96e72cacf25fed7f47de5faff78adc20946183daf3c4cf230",
		"0xef7494153ca9ddb871f4ca385ebaf47c572fbe14c39f98b5decc6d91b9230dd3",
		"0x943f86ca64a87adc346bc46a6732ea4a4c0eb7dee1453b1c37fb86f144f88658",
		"0x29ce974e8a44e6628af4749d50df04b6555bd3b932f080b0447bbe4d61f09a90",
		"0xe0c3941fc74ea7e17a80750e5923e2fca8e7db3dcf9b67d2ab4e1528524fe808",
		"0x5f8efdce13d0235c273b3714bcad8817cacb6d60867b156032f3e52cd6f32ebe",
	}

	for _, hash := range tranHashes {
		payload, err := authorizer.CreateZChainMintPayload(hash)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(payload)
	}
}

func PrintAuthorizers() {
	authorizers, err := authorizer.GetAuthorizers()
	if err != nil {
		fmt.Print(err)
	}
	fmt.Println(authorizers)
}

func fromZCNtoERC() {
	burnTrx, err := zcnbridge.BurnZCN(context.TODO(), config.Bridge.Value)
	burnTrxHash := burnTrx.Hash
	if err != nil {
		log.Logger.Fatal("failed to burn in ZCN", zap.Error(err), zap.String("hash", burnTrxHash))
	}

	// ASK authorizers for burn tickets to mint in Ethereum
	mintPayload, err := authorizer.CreateEthereumMintPayload(burnTrxHash)
	if err != nil {
		log.Logger.Fatal("failed to verify burn transactions in ZCN in CreateEthereumMintPayload", zap.Error(err), zap.String("hash", burnTrxHash))
	}

	tran, err := zcnbridge.MintWZCN(context.Background(), ConvertAmountWei, mintPayload)
	tranHash := tran.Hash().Hex()
	if err != nil {
		log.Logger.Fatal("failed to execute MintWZCN", zap.Error(err), zap.String("hash", tranHash))
	}

	// ASK for minting events from bridge contract but this is not necessary as we're going to check it by hash

	res, err := zcnbridge.ConfirmEthereumTransaction(tranHash, 60, 2)
	if err != nil {
		log.Logger.Fatal(
			"failed to confirm transaction ConfirmEthereumTransaction",
			zap.String("hash", tranHash),
			zap.Error(err),
		)
	}

	if res == 0 {
		log.Logger.Fatal("failed to confirm transaction ConfirmEthereumTransaction", zap.String("hash", tranHash))
	}
}

func fromERCtoZCN() {
	// Example: https://ropsten.etherscan.io/tx/0xa28266fb44cfc2aa27b26bd94e268e40d065a05b1a8e6339865f826557ff9f0e
	transaction, err := zcnbridge.IncreaseBurnerAllowance(context.Background(), ConvertAmountWei)
	if err != nil {
		log.Logger.Fatal("failed to execute IncreaseBurnerAllowance", zap.Error(err))
	}

	hash := transaction.Hash().Hex()
	res, err := zcnbridge.ConfirmEthereumTransaction(hash, 60, 2)
	if err != nil {
		log.Logger.Fatal(
			"failed to confirm transaction ConfirmEthereumTransaction",
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

	res, err = zcnbridge.ConfirmEthereumTransaction(burnTrxHash, 60, 2)
	if err != nil {
		log.Logger.Fatal(
			"failed to confirm transaction ConfirmEthereumTransaction",
			zap.String("hash", burnTrxHash),
			zap.Error(err),
		)
	}
	if res == 0 {
		log.Logger.Fatal("failed to confirm burn transaction in ZCN in ConfirmEthereumTransaction", zap.String("hash", burnTrxHash))
	}

	// ASK authorizers for burn tickets to mint in WZCN
	mintPayload, err := authorizer.CreateZChainMintPayload(burnTrxHash)
	if err != nil {
		log.Logger.Fatal("failed to CreateZChainMintPayload", zap.Error(err), zap.String("hash", burnTrxHash))
	}

	trx, err := zcnbridge.MintZCN(context.TODO(), mintPayload)
	if err != nil {
		log.Logger.Fatal("failed to MintZCN", zap.Error(err), zap.String("hash", trx.Hash))
	}
}
