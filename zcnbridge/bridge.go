package zcnbridge

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/0chain/gosdk/zcnbridge/ethereum/bancor"
	"math/big"
	"os"
	"time"

	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/zcnbridge/ethereum"
	"github.com/0chain/gosdk/zcnbridge/ethereum/authorizers"
	"github.com/0chain/gosdk/zcnbridge/ethereum/bridge"
	"github.com/0chain/gosdk/zcnbridge/ethereum/erc20"
	"github.com/0chain/gosdk/zcnbridge/log"
	"github.com/0chain/gosdk/zcncore"

	"github.com/0chain/gosdk/zcnbridge/transaction"
	"github.com/0chain/gosdk/zcnbridge/wallet"
	"github.com/0chain/gosdk/zcnbridge/zcnsc"
	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type (
	Wei int64
)

var Logger logger.Logger
var defaultLogLevel = logger.DEBUG

func init() {
	Logger.Init(defaultLogLevel, "zcnbridge-sdk")

	Logger.SetLevel(logger.DEBUG)
	f, err := os.OpenFile("bridge.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	Logger.SetLogFile(f, true)
}

var (
	DefaultClientIDEncoder = func(id string) []byte {
		result, err := hex.DecodeString(id)
		if err != nil {
			Logger.Fatal(err)
		}
		return result
	}
)

func (b *BridgeClient) CreateSignedTransactionFromKeyStore(client EthereumClient, gasLimitUnits uint64) *bind.TransactOpts {
	var (
		signerAddress = common.HexToAddress(b.EthereumAddress)
		password      = b.Password
	)

	signer := accounts.Account{
		Address: signerAddress,
	}

	signerAcc, err := b.keyStore.Find(signer)
	if err != nil {
		Logger.Fatal(errors.Wrapf(err, "signer: %s", signerAddress.Hex()))
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		Logger.Fatal(errors.Wrap(err, "failed to get chain ID"))
	}

	nonce, err := client.PendingNonceAt(context.Background(), signerAddress)
	if err != nil {
		Logger.Fatal(err)
	}

	gasPriceWei, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		Logger.Fatal(err)
	}

	err = b.keyStore.TimedUnlock(signer, password, time.Second*2)
	if err != nil {
		Logger.Fatal(err)
	}

	opts, err := bind.NewKeyStoreTransactorWithChainID(b.keyStore.GetEthereumKeyStore(), signerAcc, chainID)
	if err != nil {
		Logger.Fatal(err)
	}

	opts.Nonce = big.NewInt(int64(nonce))
	opts.GasLimit = gasLimitUnits // in units
	opts.GasPrice = gasPriceWei   // wei

	return opts
}

// AddEthereumAuthorizer Adds authorizer to Ethereum bridge. Only contract deployer can call this method
func (b *BridgeClient) AddEthereumAuthorizer(ctx context.Context, address common.Address) (*types.Transaction, error) {
	instance, transactOpts, err := b.prepareAuthorizers(ctx, "addAuthorizers", address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare bridge")
	}

	tran, err := instance.AddAuthorizers(transactOpts, address)
	if err != nil {
		msg := "failed to execute AddAuthorizers transaction to ClientID = %s with amount = %s"
		return nil, errors.Wrapf(err, msg, zcncore.GetClientWalletID(), address.String())
	}

	return tran, err
}

// RemoveEthereumAuthorizer Removes authorizer from Ethereum bridge. Only contract deployer can call this method
func (b *BridgeClient) RemoveEthereumAuthorizer(ctx context.Context, address common.Address) (*types.Transaction, error) {
	instance, transactOpts, err := b.prepareAuthorizers(ctx, "removeAuthorizers", address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare bridge")
	}

	tran, err := instance.RemoveAuthorizers(transactOpts, address)
	if err != nil {
		msg := "failed to execute RemoveAuthorizers transaction to ClientID = %s with amount = %s"
		return nil, errors.Wrapf(err, msg, zcncore.GetClientWalletID(), address.String())
	}

	return tran, err
}

// IncreaseBurnerAllowance Increases allowance for bridge contract address to transfer
// WZCN tokens on behalf of the token owner to the Burn TokenPool
// During the burn the script transfers amount from token owner to the bridge burn token pool
// Example: owner wants to burn some amount.
// The contract will transfer some amount from owner address to the pool.
// So the owner must call IncreaseAllowance of the WZCN token with 2 parameters:
// spender address which is the bridge contract and amount to be burned (transferred)
// ERC20 signature: "increaseAllowance(address,uint256)"
//
//nolint:funlen
func (b *BridgeClient) IncreaseBurnerAllowance(ctx context.Context, amountWei Wei) (*types.Transaction, error) {
	if amountWei <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	// 1. Data Parameter (spender)
	spenderAddress := common.HexToAddress(b.BridgeAddress)

	// 2. Data Parameter (amount)
	amount := big.NewInt(int64(amountWei))

	tokenAddress := common.HexToAddress(b.TokenAddress)

	wzcnTokenInstance, transactOpts, err := b.prepareERC20(ctx, "increaseAllowance", tokenAddress, spenderAddress, amount)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare wzcn-token")
	}

	Logger.Info(
		"Starting IncreaseAllowance",
		zap.String("token", tokenAddress.String()),
		zap.String("spender", spenderAddress.String()),
		zap.Int64("amount", amount.Int64()),
	)

	tran, err := wzcnTokenInstance.IncreaseAllowance(transactOpts, spenderAddress, amount)
	if err != nil {
		Logger.Error(
			"IncreaseAllowance FAILED",
			zap.String("token", tokenAddress.String()),
			zap.String("spender", spenderAddress.String()),
			zap.Int64("amount", amount.Int64()),
			zap.Error(err))

		return nil, errors.Wrapf(err, "failed to send `IncreaseAllowance` transaction")
	}

	Logger.Info(
		"Posted IncreaseAllowance",
		zap.String("hash", tran.Hash().String()),
		zap.String("token", tokenAddress.String()),
		zap.String("spender", spenderAddress.String()),
		zap.Int64("amount", amount.Int64()),
	)

	return tran, nil
}

// GetTokenBalance returns balance of the current client for the token address
func (b *BridgeClient) GetTokenBalance() (*big.Int, error) {
	// 1. Token address parameter
	of := common.HexToAddress(b.TokenAddress)

	// 2. User's Ethereum wallet address parameter
	from := common.HexToAddress(b.EthereumAddress)

	tokenInstance, err := erc20.NewERC20(of, b.ethereumClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize WZCN-ERC20 instance")
	}

	wei, err := tokenInstance.BalanceOf(&bind.CallOpts{}, from)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to call `BalanceOf` for %s", b.EthereumAddress)
	}

	return wei, nil
}

// VerifyZCNTransaction verifies 0CHain transaction
func (b *BridgeClient) VerifyZCNTransaction(ctx context.Context, hash string) (transaction.Transaction, error) {
	return transaction.Verify(ctx, hash)
}

// SignWithEthereumChain signs the digest with Ethereum chain signer taking key from the current user key storage
func (b *BridgeClient) SignWithEthereumChain(message string) ([]byte, error) {
	hash := crypto.Keccak256Hash([]byte(message))

	signer := accounts.Account{
		Address: common.HexToAddress(b.EthereumAddress),
	}

	signerAcc, err := b.keyStore.Find(signer)
	if err != nil {
		Logger.Fatal(err)
	}

	signature, err := b.keyStore.SignHash(signerAcc, hash.Bytes())
	if err != nil {
		return nil, err
	}

	if err != nil {
		return []byte{}, errors.Wrap(err, "failed to sign the message")
	}

	return signature, nil
}

// GetUserNonceMinted Returns nonce for a specified Ethereum address
func (b *BridgeClient) GetUserNonceMinted(ctx context.Context, rawEthereumAddress string) (*big.Int, error) {
	ethereumAddress := common.HexToAddress(rawEthereumAddress)

	contractAddress := common.HexToAddress(b.BridgeAddress)

	var bridgeInstance *bridge.Bridge
	bridgeInstance, err := bridge.NewBridge(contractAddress, b.ethereumClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bridge instance")
	}

	var nonce *big.Int
	nonce, err = bridgeInstance.GetUserNonceMinted(nil, ethereumAddress)
	if err != nil {
		Logger.Error("GetUserNonceMinted FAILED", zap.Error(err))
		msg := "failed to execute GetUserNonceMinted call, ethereumAddress = %s"
		return nil, errors.Wrapf(err, msg, rawEthereumAddress)
	}
	return nonce, err
}

// MintWZCN Mint ZCN tokens on behalf of the 0ZCN client
// payload: received from authorizers
func (b *BridgeClient) MintWZCN(ctx context.Context, payload *ethereum.MintPayload) (*types.Transaction, error) {
	if DefaultClientIDEncoder == nil {
		return nil, errors.New("DefaultClientIDEncoder must be setup")
	}

	// 1. Burned amount parameter
	amount := new(big.Int)
	amount.SetInt64(payload.Amount) // wei

	// 2. Transaction ID Parameter of burn operation (zcnTxd string as []byte)
	zcnTxd := DefaultClientIDEncoder(payload.ZCNTxnID)

	// 3. Nonce Parameter generated during burn operation
	nonce := new(big.Int)
	nonce.SetInt64(payload.Nonce)

	// 4. Signature
	// For requirements from ERC20 authorizer, the signature length must be 65
	var sigs [][]byte
	for _, signature := range payload.Signatures {
		sigs = append(sigs, signature.Signature)
	}

	// 5. To Ethereum address

	toAddress := common.HexToAddress(payload.To)

	bridgeInstance, transactOpts, err := b.prepareBridge(ctx, payload.To, "mint", toAddress, amount, zcnTxd, nonce, sigs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare bridge")
	}

	Logger.Info(
		"Staring Mint WZCN",
		zap.Int64("amount", amount.Int64()),
		zap.String("zcnTxd", string(zcnTxd)),
		zap.String("nonce", nonce.String()))

	var tran *types.Transaction
	tran, err = bridgeInstance.Mint(transactOpts, toAddress, amount, zcnTxd, nonce, sigs)
	if err != nil {
		Logger.Error("Mint WZCN FAILED", zap.Error(err))
		msg := "failed to execute MintWZCN transaction, amount = %s, ZCN TrxID = %s"
		return nil, errors.Wrapf(err, msg, amount, zcnTxd)
	}

	Logger.Info(
		"Posted Mint WZCN",
		zap.String("hash", tran.Hash().String()),
		zap.Int64("amount", amount.Int64()),
		zap.String("zcnTxd", string(zcnTxd)),
		zap.String("nonce", nonce.String()),
	)

	return tran, err
}

// BurnWZCN Burns WZCN tokens on behalf of the 0ZCN client
// amountTokens - ZCN tokens
// clientID - 0ZCN client
// ERC20 signature: "burn(uint256,bytes)"
func (b *BridgeClient) BurnWZCN(ctx context.Context, amountTokens uint64) (*types.Transaction, error) {
	if DefaultClientIDEncoder == nil {
		return nil, errors.New("DefaultClientIDEncoder must be setup")
	}

	// 1. Data Parameter (amount to burn)
	clientID := DefaultClientIDEncoder(zcncore.GetClientWalletID())

	// 2. Data Parameter (signature)
	amount := new(big.Int)
	amount.SetInt64(int64(amountTokens))

	bridgeInstance, transactOpts, err := b.prepareBridge(ctx, b.EthereumAddress, "burn", amount, clientID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare bridge")
	}

	Logger.Info(
		"Staring Burn WZCN",
		zap.Int64("amount", amount.Int64()),
	)

	tran, err := bridgeInstance.Burn(transactOpts, amount, clientID)
	if err != nil {
		msg := "failed to execute Burn WZCN transaction to ClientID = %s with amount = %s"
		return nil, errors.Wrapf(err, msg, zcncore.GetClientWalletID(), amount)
	}

	Logger.Info(
		"Posted Burn WZCN",
		zap.String("clientID", zcncore.GetClientWalletID()),
		zap.Int64("amount", amount.Int64()),
	)

	return tran, err
}

// MintZCN mints ZCN tokens after receiving proof-of-burn of WZCN tokens
func (b *BridgeClient) MintZCN(ctx context.Context, payload *zcnsc.MintPayload) (string, error) {
	trx, err := b.transactionProvider.NewTransactionEntity(0)
	if err != nil {
		log.Logger.Fatal("failed to create new transaction", zap.Error(err))
	}

	Logger.Info(
		"Starting MINT smart contract",
		zap.String("sc address", wallet.ZCNSCSmartContractAddress),
		zap.String("function", wallet.MintFunc),
		zap.Int64("mint amount", int64(payload.Amount)))

	hash, err := trx.ExecuteSmartContract(
		ctx,
		wallet.ZCNSCSmartContractAddress,
		wallet.MintFunc,
		payload,
		0)

	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to execute smart contract, hash = %s", hash))
	}

	Logger.Info(
		"Mint ZCN transaction",
		zap.String("hash", hash),
		zap.Int64("mint amount", int64(payload.Amount)))

	return hash, nil
}

// BurnZCN burns ZCN tokens before conversion from ZCN to WZCN as a first step
func (b *BridgeClient) BurnZCN(ctx context.Context, amount, txnfee uint64) (transaction.Transaction, error) {
	payload := zcnsc.BurnPayload{
		EthereumAddress: b.EthereumAddress,
	}

	trx, err := b.transactionProvider.NewTransactionEntity(txnfee)
	if err != nil {
		log.Logger.Fatal("failed to create new transaction", zap.Error(err))
	}

	Logger.Info(
		"Starting BURN smart contract",
		zap.String("sc address", wallet.ZCNSCSmartContractAddress),
		zap.String("function", wallet.BurnFunc),
		zap.Uint64("burn amount", amount),
	)

	var hash string
	hash, err = trx.ExecuteSmartContract(
		ctx,
		wallet.ZCNSCSmartContractAddress,
		wallet.BurnFunc,
		payload,
		amount,
	)

	if err != nil {
		Logger.Error("Burn ZCN transaction FAILED", zap.Error(err))
		return trx, errors.Wrap(err, fmt.Sprintf("failed to execute smart contract, hash = %s", hash))
	}

	err = trx.Verify(context.Background())
	if err != nil {
		return trx, errors.Wrap(err, fmt.Sprintf("failed to verify smart contract, hash = %s", hash))
	}

	Logger.Info(
		"Burn ZCN transaction",
		zap.String("hash", hash),
		zap.Uint64("burn amount", amount),
		zap.Uint64("amount", amount),
	)

	return trx, nil
}

// Swap is used for token swap operation.
func (b *BridgeClient) Swap(ctx context.Context, amountSwap uint64) (*types.Transaction, error) {
	// 1. Swap amount parameter.
	amount := big.NewInt(int64(amountSwap))

	// 2. User's Ethereum wallet address.
	beneficiary := common.HexToAddress(b.EthereumAddress)

	// 3. Trade transaction deadline.
	deadline := big.NewInt(time.Now().Add(time.Minute).Unix())

	// 4. Source token address parameter
	from := common.HexToAddress(b.SourceTokenAddress)

	// 5. Target token address parameter
	to := common.HexToAddress(b.TokenAddress)

	bancorInstance, transactOpts, err := b.prepareBancor(ctx, "tradeByTargetAmount", from, to, amount, big.NewInt(100), deadline, beneficiary)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare bancor")
	}

	Logger.Info(
		"Starting Swap",
		zap.Int64("amount", amount.Int64()),
		zap.String("sourceToken", b.SourceTokenAddress),
	)

	tran, err := bancorInstance.TradeByTargetAmount(transactOpts, from, to, amount, big.NewInt(100), deadline, beneficiary)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute ConvertByPath transaction")
	}

	return tran, nil
}

func (b *BridgeClient) prepareBancor(ctx context.Context, method string, params ...interface{}) (*bancor.IBancorNetwork, *bind.TransactOpts, error) {
	// To (contract)
	contractAddress := common.HexToAddress(BancorNetworkAddress)
	//
	//abi, err := bancor.BancorNetworkMetaData.GetAbi()
	//if err != nil {
	//	return nil, nil, errors.Wrap(err, "failed to get bancor abi")
	//}

	//pack, err := abi.Pack(method, params...)
	//if err != nil {
	//	return nil, nil, errors.Wrap(err, "failed to pack arguments")
	//}
	//
	//from := common.HexToAddress(b.EthereumAddress)

	//gasLimitUnits, err := b.ethereumClient.EstimateGas(ctx, eth.CallMsg{
	//	To:   &contractAddress,
	//	From: from,
	//	Data: pack,
	//})
	//if err != nil {
	//	return nil, nil, errors.Wrap(err, "failed to estimate gas limit")
	//}
	//
	//gasLimitUnits = addPercents(gasLimitUnits, 10).Uint64()

	transactOpts := b.CreateSignedTransactionFromKeyStore(b.ethereumClient, 400000)

	bancorInstance, err := bancor.NewIBancorNetwork(contractAddress, b.ethereumClient)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to initialize bancor instance")
	}

	return bancorInstance, transactOpts, nil
}

func (b *BridgeClient) prepareERC20(ctx context.Context, method string, tokenAddress common.Address, params ...interface{}) (*erc20.ERC20, *bind.TransactOpts, error) {
	abi, err := erc20.ERC20MetaData.GetAbi()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get erc20 abi")
	}

	pack, err := abi.Pack(method, params...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to pack arguments")
	}

	from := common.HexToAddress(b.EthereumAddress)

	gasLimitUnits, err := b.ethereumClient.EstimateGas(ctx, eth.CallMsg{
		To:   &tokenAddress,
		From: from,
		Data: pack,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to estimate gas limit")
	}

	gasLimitUnits = addPercents(gasLimitUnits, 10).Uint64()

	transactOpts := b.CreateSignedTransactionFromKeyStore(b.ethereumClient, gasLimitUnits)

	tokenInstance, err := erc20.NewERC20(tokenAddress, b.ethereumClient)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to initialize erc20 instance")
	}

	return tokenInstance, transactOpts, nil
}

func (b *BridgeClient) prepareAuthorizers(ctx context.Context, method string, params ...interface{}) (*authorizers.Authorizers, *bind.TransactOpts, error) {
	// To (contract)
	contractAddress := common.HexToAddress(b.AuthorizersAddress)

	// Get ABI of the contract
	abi, err := authorizers.AuthorizersMetaData.GetAbi()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get ABI")
	}

	// Pack the method argument
	pack, err := abi.Pack(method, params...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to pack arguments")
	}

	from := common.HexToAddress(b.EthereumAddress)

	// Gas limits in units
	gasLimitUnits, err := b.ethereumClient.EstimateGas(ctx, eth.CallMsg{
		To:   &contractAddress,
		From: from,
		Data: pack,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to estimate gas")
	}

	// Update gas limits + 10%
	gasLimitUnits = addPercents(gasLimitUnits, 10).Uint64()

	transactOpts := b.CreateSignedTransactionFromKeyStore(b.ethereumClient, gasLimitUnits)

	// Authorizers instance
	authorizersInstance, err := authorizers.NewAuthorizers(contractAddress, b.ethereumClient)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create authorizers instance")
	}

	return authorizersInstance, transactOpts, nil
}

func (b *BridgeClient) prepareBridge(ctx context.Context, ethereumAddress, method string, params ...interface{}) (*bridge.Bridge, *bind.TransactOpts, error) {
	// To (contract)
	contractAddress := common.HexToAddress(b.BridgeAddress)

	//Get ABI of the contract
	abi, err := bridge.BridgeMetaData.GetAbi()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get ABI")
	}

	//Pack the method argument
	pack, err := abi.Pack(method, params...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to pack arguments")
	}

	//Gas limits in units
	fromAddress := common.HexToAddress(ethereumAddress)

	gasLimitUnits, err := b.ethereumClient.EstimateGas(ctx, eth.CallMsg{
		To:   &contractAddress,
		From: fromAddress,
		Data: pack,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to estimate gas")
	}

	//Update gas limits + 10%
	gasLimitUnits = addPercents(gasLimitUnits, 10).Uint64()

	transactOpts := b.CreateSignedTransactionFromKeyStore(b.ethereumClient, gasLimitUnits)

	// BridgeClient instance
	bridgeInstance, err := bridge.NewBridge(contractAddress, b.ethereumClient)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create bridge instance")
	}

	return bridgeInstance, transactOpts, nil
}
