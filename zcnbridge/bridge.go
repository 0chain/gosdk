package zcnbridge

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/0chain/gosdk/zcnbridge/zcnsc"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/0chain/gosdk/zcnbridge/config"
	"github.com/0chain/gosdk/zcnbridge/ethereum"
	"github.com/0chain/gosdk/zcnbridge/ethereum/bridge"
	"github.com/0chain/gosdk/zcnbridge/ethereum/erc20"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/zcnbridge/node"
	"github.com/0chain/gosdk/zcnbridge/transaction"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/0chain/gosdk/zcncore"
	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/sha3"
)

const (
	IncreaseAllowanceSigCode = "39509351"
	BurnSigCode              = "fe9d9303"
	MintSigCode              = "4d02be9f"
	Bytes32                  = 32
)

type (
	wei int64
)

var (
	// IncreaseAllowanceSig "increaseAllowance(address,uint256)"
	IncreaseAllowanceSig = []byte(erc20.ERC20MetaData.Sigs[IncreaseAllowanceSigCode])
	// BurnSig "burn(uint256,bytes)"
	BurnSig                = []byte(bridge.BridgeMetaData.Sigs[BurnSigCode])
	MintSig                = []byte(bridge.BridgeMetaData.Sigs[MintSigCode])
	DefaultClientIDEncoder = func(id string) []byte {
		return []byte(id)
	}
)

// InitBridge Sets up the wallet and node
// Wallet setup reads keys from keyfile and registers in the 0chain
func InitBridge() {
	client, err := wallet.Setup()
	if err != nil {
		log.Logger.Fatal("failed to setup wallet", zap.Error(err))
	}

	node.Start(client)
}

// IncreaseBurnerAllowance Increases allowance for bridge contract address to transfer
// WZCN tokens on behalf of the token owner to the TokenPool
func IncreaseBurnerAllowance(amountWei wei) (*types.Transaction, error) {
	// 1. Create etherClient
	etherClient, err := ethereum.CreateEthClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create etherClient")
	}

	// 1. Data Parameter (signature)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(IncreaseAllowanceSig)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0x39509351

	// 2. Data Parameter (spender)
	spenderAddress := common.HexToAddress(config.Bridge.BridgeAddress)
	spenderPaddedAddress := common.LeftPadBytes(spenderAddress.Bytes(), Bytes32)
	fmt.Println(hexutil.Encode(spenderPaddedAddress)) // 0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d

	// 3. Data Parameter (amount)
	amount := big.NewInt(int64(amountWei))
	paddedAmount := common.LeftPadBytes(amount.Bytes(), Bytes32)
	fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000

	var data []byte
	data = append(data, methodID...)
	data = append(data, spenderPaddedAddress...)
	data = append(data, paddedAmount...)

	// To
	tokenAddress := common.HexToAddress(config.Bridge.WzcnAddress)

	gasLimitUnits, err := etherClient.EstimateGas(context.Background(), eth.CallMsg{
		To:   &tokenAddress, // FIXME: From: is required?
		Data: data,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to estimate gas limit")
	}

	gasLimitUnits = AddPercents(gasLimitUnits, 10).Uint64()

	ownerAddress, privKey, err := ethereum.PrivateKeyAndAddress()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read private key and ownerAddress")
	}

	transactOpts := ethereum.CreateSignedTransaction(etherClient, ownerAddress, privKey, gasLimitUnits)

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

func GetTransactionStatus(hash string) int {
	return zcncore.CheckEthHashStatus(hash)
}

func ConfirmEthereumTransactionStatus(hash string, times int, duration time.Duration) int {
	var res = 0
	for i := 0; i < times; i++ {
		res = GetTransactionStatus(hash)
		if res == 1 {
			break
		}
		log.Logger.Info(fmt.Sprintf("try # %d", i))
		time.Sleep(time.Second * duration)
	}
	return res
}

func MintWZCN(amountTokens wei, payload *ethereum.MintPayload) (*types.Transaction, error) {
	if DefaultClientIDEncoder == nil {
		return nil, errors.New("DefaultClientIDEncoder must be setup")
	}

	zcnTxd := DefaultClientIDEncoder(payload.ZCNTxnID)

	// 1. Data Parameter (signature)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(MintSig)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID))

	// 2. Data Parameter (amount to burn)
	amount := new(big.Int)
	amount.SetInt64(int64(amountTokens)) // wei
	paddedAmount := common.LeftPadBytes(amount.Bytes(), Bytes32)

	// 3. Data Parameter (zcnTxd string as []byte)
	paddedZCNTxd := common.LeftPadBytes(zcnTxd, Bytes32)

	// 4. Nonce Parameter
	nonce := new(big.Int)
	nonce.SetInt64(payload.Nonce)
	paddedNonce := common.LeftPadBytes(amount.Bytes(), Bytes32)

	// Signature
	// For requirements from ERC20 authorizer, the signature length must be 65
	var sb strings.Builder
	for _, signature := range payload.Signatures {
		sb.WriteString(signature.Signature)
	}
	sigs := []byte(sb.String())

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAmount...)
	data = append(data, paddedZCNTxd...)
	data = append(data, paddedNonce...)
	data = append(data, sigs...)

	bridgeInstance, transactOpts, err := prepareBridge(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare bridge")
	}

	tran, err := bridgeInstance.Mint(transactOpts, amount, zcnTxd, nonce, sigs)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to execute MintWZCN transaction, ClientID = %s, amount = %s, ZCN TrxID = %s",
			node.ID(),
			amount,
			zcnTxd,
		)
	}

	return tran, err
}

// BurnWZCN Burns WZCN tokens on behalf of the 0ZCN client
// amountTokens - ZCN tokens
// clientID - 0ZCN client
func BurnWZCN(amountTokens int64) (*types.Transaction, error) {
	if DefaultClientIDEncoder == nil {
		return nil, errors.New("DefaultClientIDEncoder must be setup")
	}

	clientID := DefaultClientIDEncoder(node.ID())

	// 1. Data Parameter (signature)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(BurnSig)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) // 0xfe9d9303

	// 2. Data Parameter (amount to burn)
	amount := new(big.Int)
	amount.SetInt64(amountTokens)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), Bytes32)

	// 3. Data Parameter (clientID string as []byte)
	paddedClientID := common.LeftPadBytes(clientID, Bytes32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAmount...)
	data = append(data, paddedClientID...)

	bridgeInstance, transactOpts, err := prepareBridge(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare bridge")
	}

	tran, err := bridgeInstance.Burn(transactOpts, amount, clientID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute BurnZCN transaction to ClientID = %s with amount = %s", node.ID(), amount)
	}

	return tran, err
}

func prepareBridge(data []byte) (*bridge.Bridge, *bind.TransactOpts, error) {
	etherClient, err := ethereum.CreateEthClient()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create etherClient")
	}

	// To
	bridgeAddress := common.HexToAddress(config.Bridge.BridgeAddress)

	ownerAddress, privKey, err := ethereum.PrivateKeyAndAddress()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to read private key and ownerAddress")
	}

	// Gas limits in units
	gasLimitUnits, err := etherClient.EstimateGas(context.Background(), eth.CallMsg{
		To:   &bridgeAddress, // TODO: From: is required?
		Data: data,
	})

	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to estimate gas")
	}

	gasLimitUnits = AddPercents(gasLimitUnits, 10).Uint64()

	transactOpts := ethereum.CreateSignedTransaction(etherClient, ownerAddress, privKey, gasLimitUnits)

	bridgeInstance, err := bridge.NewBridge(bridgeAddress, etherClient)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create bridge")
	}

	return bridgeInstance, transactOpts, nil
}

func MintZCN(ctx context.Context, payload *zcnsc.MintPayload) (*transaction.Transaction, error) {
	trx, err := transaction.NewTransactionEntity()
	if err != nil {
		log.Logger.Fatal("failed to create new transaction", zap.Error(err))
	}

	hash, err := trx.ExecuteSmartContract(
		ctx,
		wallet.ZCNSCSmartContractAddress,
		wallet.MintFunc,
		string(payload.Encode()),
		0,
	)
	if err != nil {
		return trx, errors.Wrap(err, fmt.Sprintf("failed to execute smart contract, hash = %s", hash))
	}

	err = trx.Verify(ctx)
	if err != nil {
		return trx, errors.Wrap(err, fmt.Sprintf("failed to verify smart contract transaction, hash = %s", hash))
	}

	return trx, nil
}

func BurnZCN(ctx context.Context, value int64) (*transaction.Transaction, error) {
	payload := zcnsc.BurnPayload{
		Nonce:           node.IncrementNonce(),
		EthereumAddress: config.Bridge.EthereumAddress,
	}

	trx, err := transaction.NewTransactionEntity()
	if err != nil {
		log.Logger.Fatal("failed to create new transaction", zap.Error(err))
	}

	hash, err := trx.ExecuteSmartContract(
		ctx,
		wallet.ZCNSCSmartContractAddress,
		wallet.BurnFunc,
		string(payload.Encode()),
		value,
	)
	if err != nil {
		return trx, errors.Wrap(err, fmt.Sprintf("failed to execute smart contract, hash = %s", hash))
	}

	err = trx.Verify(ctx)
	if err != nil {
		return trx, errors.Wrap(err, fmt.Sprintf("failed to verify smart contract transaction, hash = %s", hash))
	}

	return trx, nil
}

func AddPercents(gasLimitUnits uint64, percents int) *big.Int {
	gasLimitBig := big.NewInt(int64(gasLimitUnits))
	factorBig := big.NewInt(int64(percents))
	deltaBig := gasLimitBig.Div(gasLimitBig, factorBig)
	gasLimitBig = gasLimitBig.Add(gasLimitBig, deltaBig)
	return gasLimitBig
}
