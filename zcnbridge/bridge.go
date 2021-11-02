package zcnbridge

import (
	"context"
	"fmt"
	"github.com/0chain/gosdk/zcnbridge/ethereum/erc20"
	"github.com/0chain/gosdk/zcnbridge/ethereum/bridge"
	"github.com/0chain/gosdk/zcncore"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	//"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
	"math/big"
)

const (
	IncreaseAllowanceSigCode = "39509351"
	Bytes32                  = 32
)

const (
	Failed = types.ReceiptStatusFailed
	Success = types.ReceiptStatusSuccessful
)

var (
	IncreaseAllowanceSig = []byte(erc20.ERC20MetaData.Sigs[IncreaseAllowanceSigCode])
	BurnSig = []byte(bridge.BridgeMetaData.Sigs[IncreaseAllowanceSigCode])
	DefaultEncoder  = func(id string) []byte {
		return []byte(id)
	}
)

// Description:
// 1. Increase the amount for token
// 2. Call burn using same amount
// 3. Confirm transaction was executed

func InitBridge() {
	// Read config from file
	config.gasLimit = 300000 // TODO: InitBridge - wei, gwei, unit, tokens?
}

// IncreaseBurnerAllowance TODO: Is amount in wei?
// IncreaseBurnerAllowance Increases allowance for bridge contract address to transfer
// WZCN tokens on behalf of the token owner
func IncreaseBurnerAllowance(amountTokens int64) (*types.Transaction, error) {
	client, err := createClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}

	// To:
	tokenAddress := common.HexToAddress(config.wzcnAddress)

	// 1. Data Parameter:
	hash := sha3.NewLegacyKeccak256()
	hash.Write(IncreaseAllowanceSig)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0x39509351

	// 2. Data Parameter:
	spenderAddress := common.HexToAddress(config.bridgeAddress)
	spenderPaddedAddress := common.LeftPadBytes(spenderAddress.Bytes(), Bytes32)
	fmt.Println(hexutil.Encode(spenderPaddedAddress)) // 0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d

	// 3. Data Parameter:
	amount := new(big.Int)
	amount.SetInt64(amountTokens)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), Bytes32)
	fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000

	var data []byte
	data = append(data, methodID...)
	data = append(data, spenderPaddedAddress...)
	data = append(data, paddedAmount...)

	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	})
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	// TODO: This needs to fix
	gasLimit = gasLimit + gasLimit / 100

	ownerAddress, privKey, err := ownerPrivateKeyAndAddress()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read private key and ownerAddress")
	}

	transactOpts := createSignedTransaction(client, ownerAddress, privKey, gasLimit)

	wzcnToken, err := erc20.NewERC20(tokenAddress, client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize WZCN-ERC20 instance")
	}

	tran, err := wzcnToken.IncreaseAllowance(transactOpts, spenderAddress, amount)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send `IncreaseAllowance` transaction")
	}

	return tran, nil
}

func TransactionStatus(hash string) int {
	return zcncore.CheckEthHashStatus(hash)
}

//func BurnWZCN(amount, clientId string, ctx context.Context)(*types.Transaction, error) {
//	if DefaultEncoder == nil {
//		return nil, errors.New("DefaultEncoder must be setup")
//	}
//
//	client, err := createClient()
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to create client")
//	}
//
//	// To:
//	bridgeAddress := common.HexToAddress(config.bridgeAddress)
//
//	ownerAddress, privKey, err := ownerPrivateKeyAndAddress()
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to read private key and ownerAddress")
//	}
//
//	transactOpts := createSignedTransaction(client, ownerAddress, privKey, gasLimit)
//}