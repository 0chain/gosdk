package zcnbridge

import (
	"context"
	"fmt"

	"github.com/0chain/gosdk/zcnbridge/ethereum/bridge"
	"github.com/0chain/gosdk/zcnbridge/ethereum/erc20"
	"github.com/0chain/gosdk/zcncore"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"

	//"github.com/ethereum/go-ethereum/ethclient"
	"math/big"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

const (
	IncreaseAllowanceSigCode = "39509351"
	BurnSigCode              = "fe9d9303"
	Bytes32                  = 32
)

var (
	// IncreaseAllowanceSig "increaseAllowance(address,uint256)"
	IncreaseAllowanceSig = []byte(erc20.ERC20MetaData.Sigs[IncreaseAllowanceSigCode])
	// BurnSig "burn(uint256,bytes)"
	BurnSig                = []byte(bridge.BridgeMetaData.Sigs[BurnSigCode])
	DefaultClientIDEncoder = func(id string) []byte {
		return []byte(id)
	}
)

// Description:
// 1. Increase the amount for token
// 2. Call burn using same amount
// 3. Confirm transaction was executed

func InitBridge() {
}

// IncreaseBurnerAllowance FIXME: Is amount in wei?
// IncreaseBurnerAllowance Increases allowance for bridge contract address to transfer
// WZCN tokens on behalf of the token owner to the TokenPool
func IncreaseBurnerAllowance(amountTokens int64) (*types.Transaction, error) {
	// 1. Create etherClient
	etherClient, err := createEthClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create etherClient")
	}

	// 1. Data Parameter (signature)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(IncreaseAllowanceSig)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0x39509351

	// 2. Data Parameter (spender)
	spenderAddress := common.HexToAddress(config.bridgeAddress)
	spenderPaddedAddress := common.LeftPadBytes(spenderAddress.Bytes(), Bytes32)
	fmt.Println(hexutil.Encode(spenderPaddedAddress)) // 0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d

	// 3. Data Parameter (amount)
	amount := new(big.Int)
	amount.SetInt64(amountTokens)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), Bytes32)
	fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000

	var data []byte
	data = append(data, methodID...)
	data = append(data, spenderPaddedAddress...)
	data = append(data, paddedAmount...)

	// To
	tokenAddress := common.HexToAddress(config.wzcnAddress)

	gasLimit, err := etherClient.EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &tokenAddress, // FIXME: From: is required?
		Data: data,
	})
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	// FIXME: proper calculation
	gasLimit = gasLimit + gasLimit/10

	ownerAddress, privKey, err := ownerPrivateKeyAndAddress()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read private key and ownerAddress")
	}

	transactOpts := createSignedTransaction(etherClient, ownerAddress, privKey, gasLimit)

	wzcnTokenInstance, err := erc20.NewERC20(tokenAddress, etherClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize WZCN-ERC20 instance")
	}

	tran, err := wzcnTokenInstance.IncreaseAllowance(transactOpts, spenderAddress, amount)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send `IncreaseAllowance` transaction")
	}

	return tran, nil
}

func TransactionStatus(hash string) int {
	return zcncore.CheckEthHashStatus(hash)
}

// BurnWZCN Burns WZCN tokens on behalf of the 0ZCN client
// amountTokens - ZCN tokens
// clientID - 0ZCN client
func BurnWZCN(amountTokens int64, clientID string) (*types.Transaction, error) {
	if DefaultClientIDEncoder == nil {
		return nil, errors.New("DefaultClientIDEncoder must be setup")
	}

	// 1. Create etherClient
	etherClient, err := createEthClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create etherClient")
	}

	// 1. Data Parameter (signature)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(BurnSig)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0xfe9d9303

	// 2. Data Parameter (amount to burn)
	amount := new(big.Int)
	amount.SetInt64(amountTokens)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), Bytes32)
	fmt.Println(hexutil.Encode(paddedAmount))

	// 3. Data Parameter (clientID string as []byte)
	paddedClientID := common.LeftPadBytes(DefaultClientIDEncoder(clientID), Bytes32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAmount...)
	data = append(data, paddedClientID...)

	// To
	bridgeAddress := common.HexToAddress(config.bridgeAddress)

	ownerAddress, privKey, err := ownerPrivateKeyAndAddress()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read private key and ownerAddress")
	}

	gasLimit, err := etherClient.EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &bridgeAddress, // TODO: From: is required?
		Data: data,
	})
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	// TODO: This needs to fix
	gasLimit = gasLimit + gasLimit/10

	transactOpts := createSignedTransaction(etherClient, ownerAddress, privKey, gasLimit)

	bridgeInstance, err := bridge.NewBridge(bridgeAddress, etherClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bridge instance")
	}

	transaction, err := bridgeInstance.Burn(transactOpts, amount, DefaultClientIDEncoder(clientID))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute Burn transaction to ClientID = %s with amount = %s", clientID, amount)
	}

	return transaction, err
}
