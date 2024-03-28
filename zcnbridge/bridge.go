package zcnbridge

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ybbus/jsonrpc/v3"

	"github.com/0chain/gosdk/zcnbridge/ethereum/bancortoken"

	"github.com/0chain/common/core/currency"
	"github.com/0chain/gosdk/zcnbridge/ethereum/bancornetwork"
	"github.com/0chain/gosdk/zcnbridge/ethereum/zcntoken"
	h "github.com/0chain/gosdk/zcnbridge/http"
	hdw "github.com/0chain/gosdk/zcncore/ethhdwallet"
	"github.com/spf13/viper"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/0chain/gosdk/core/logger"
	"github.com/0chain/gosdk/zcnbridge/ethereum"
	"github.com/0chain/gosdk/zcnbridge/ethereum/authorizers"
	"github.com/0chain/gosdk/zcnbridge/ethereum/bridge"
	"github.com/0chain/gosdk/zcnbridge/ethereum/nftconfig"
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

var Logger logger.Logger
var defaultLogLevel = logger.DEBUG

func init() {
	Logger.Init(defaultLogLevel, "zcnbridge-sdk")

	Logger.SetLevel(logger.DEBUG)
	ioWriter := &lumberjack.Logger{
		Filename:   "bridge.log",
		MaxSize:    100, // MB
		MaxBackups: 5,   // number of backups
		MaxAge:     28,  //days
		LocalTime:  false,
		Compress:   false, // disabled by default
	}
	Logger.SetLogFile(ioWriter, true)
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

func (b *BridgeClient) AddEthereumAuthorizers(configDir string) {
	cfg := viper.New()
	cfg.AddConfigPath(configDir)
	cfg.SetConfigName("authorizers")
	if err := cfg.ReadInConfig(); err != nil {
		fmt.Println(err)
		return
	}

	mnemonics := cfg.GetStringSlice("authorizers")

	for _, mnemonic := range mnemonics {
		wallet, err := hdw.NewFromMnemonic(mnemonic)
		if err != nil {
			fmt.Printf("failed to read mnemonic: %v", err)
			continue
		}

		pathD := hdw.MustParseDerivationPath("m/44'/60'/0'/0/0")
		account, err := wallet.Derive(pathD, true)
		if err != nil {
			fmt.Println(err)
			continue
		}

		transaction, err := b.AddEthereumAuthorizer(context.TODO(), account.Address)
		if err != nil || transaction == nil {
			fmt.Printf("AddAuthorizer error: %v, Address: %s", err, account.Address.Hex())
			continue
		}

		status, err := ConfirmEthereumTransaction(transaction.Hash().String(), 100, time.Second*10)
		if err != nil {
			fmt.Println(err)
		}

		if status == 1 {
			fmt.Printf("Authorizer has been added: %s\n", mnemonic)
		} else {
			fmt.Printf("Authorizer has failed to be added: %s\n", mnemonic)
		}
	}
}

func (b *BridgeClient) prepareNFTConfig(ctx context.Context, method string, params ...interface{}) (*nftconfig.NFTConfig, *bind.TransactOpts, error) {
	// To (contract)
	contractAddress := common.HexToAddress(b.NFTConfigAddress)

	// Get ABI of the contract
	abi, err := nftconfig.NFTConfigMetaData.GetAbi()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get nftconfig ABI")
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

	// NFTConfig instance
	cfg, err := nftconfig.NewNFTConfig(contractAddress, b.ethereumClient)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create nftconfig instance")
	}

	return cfg, transactOpts, nil
}

// EncodePackInt do abi.encodedPack(string, int), it is used for setting plan id for royalty
func EncodePackInt64(key string, param int64) common.Hash {
	return crypto.Keccak256Hash(
		[]byte(key),
		common.LeftPadBytes(big.NewInt(param).Bytes(), 32),
	)
}

// NFTConfigSetUint256 call setUint256 method of NFTConfig contract
func (b *BridgeClient) NFTConfigSetUint256(ctx context.Context, key string, value int64) (*types.Transaction, error) {
	kkey := crypto.Keccak256Hash([]byte(key))
	return b.NFTConfigSetUint256Raw(ctx, kkey, value)
}

func (b *BridgeClient) NFTConfigSetUint256Raw(ctx context.Context, key common.Hash, value int64) (*types.Transaction, error) {
	if value < 0 {
		return nil, errors.New("value must be greater than zero")
	}

	v := big.NewInt(value)
	Logger.Debug("NFT config setUint256", zap.String("key", key.String()), zap.Any("value", v))
	instance, transactOpts, err := b.prepareNFTConfig(ctx, "setUint256", key, v)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare bridge")
	}

	tran, err := instance.SetUint256(transactOpts, key, v)
	if err != nil {
		msg := "failed to execute setUint256 transaction to ClientID = %s with key = %s, value = %v"
		return nil, errors.Wrapf(err, msg, zcncore.GetClientWalletID(), key, v)
	}

	return tran, err
}

func (b *BridgeClient) NFTConfigGetUint256(ctx context.Context, key string, keyParam ...int64) (string, int64, error) {
	kkey := crypto.Keccak256Hash([]byte(key))
	if len(keyParam) > 0 {
		kkey = EncodePackInt64(key, keyParam[0])
	}

	contractAddress := common.HexToAddress(b.NFTConfigAddress)

	cfg, err := nftconfig.NewNFTConfig(contractAddress, b.ethereumClient)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to create NFT config instance")
	}

	v, err := cfg.GetUint256(nil, kkey)
	if err != nil {
		Logger.Error("NFTConfig GetUint256 FAILED", zap.Error(err))
		msg := "failed to execute getUint256 call, key = %s"
		return "", 0, errors.Wrapf(err, msg, kkey)
	}
	return kkey.String(), v.Int64(), err
}

func (b *BridgeClient) NFTConfigSetAddress(ctx context.Context, key, address string) (*types.Transaction, error) {
	kkey := crypto.Keccak256Hash([]byte(key))
	// return b.NFTConfigSetAddress(ctx, kkey, address)

	Logger.Debug("NFT config setAddress",
		zap.String("key", kkey.String()),
		zap.String("address", address))

	addr := common.HexToAddress(address)
	instance, transactOpts, err := b.prepareNFTConfig(ctx, "setAddress", kkey, addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare bridge")
	}

	tran, err := instance.SetAddress(transactOpts, kkey, addr)
	if err != nil {
		msg := "failed to execute setAddress transaction to ClientID = %s with key = %s, value = %v"
		return nil, errors.Wrapf(err, msg, zcncore.GetClientWalletID(), key, address)
	}

	return tran, err
}

func (b *BridgeClient) NFTConfigGetAddress(ctx context.Context, key string) (string, string, error) {
	kkey := crypto.Keccak256Hash([]byte(key))

	contractAddress := common.HexToAddress(b.NFTConfigAddress)

	cfg, err := nftconfig.NewNFTConfig(contractAddress, b.ethereumClient)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create NFT config instance")
	}

	v, err := cfg.GetAddress(nil, kkey)
	if err != nil {
		Logger.Error("NFTConfig GetAddress FAILED", zap.Error(err))
		msg := "failed to execute getAddress call, key = %s"
		return "", "", errors.Wrapf(err, msg, kkey)
	}
	return kkey.String(), v.String(), err
}

// IncreaseBurnerAllowance Increases allowance for bridge contract address to transfer
// ERC-20 tokens on behalf of the zcntoken owner to the Burn TokenPool
// During the burn the script transfers amount from zcntoken owner to the bridge burn zcntoken pool
// Example: owner wants to burn some amount.
// The contract will transfer some amount from owner address to the pool.
// So the owner must call IncreaseAllowance of the WZCN zcntoken with 2 parameters:
// spender address which is the bridge contract and amount to be burned (transferred)
// Token signature: "increaseApproval(address,uint256)"
//
//nolint:funlen
func (b *BridgeClient) IncreaseBurnerAllowance(ctx context.Context, allowanceAmount uint64) (*types.Transaction, error) {
	if allowanceAmount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	// 1. Data Parameter (spender)
	spenderAddress := common.HexToAddress(b.BridgeAddress)

	// 2. Data Parameter (amount)
	amount := big.NewInt(int64(allowanceAmount))

	tokenAddress := common.HexToAddress(b.TokenAddress)

	tokenInstance, transactOpts, err := b.prepareToken(ctx, "increaseApproval", tokenAddress, spenderAddress, amount)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare zcntoken")
	}

	Logger.Info(
		"Starting IncreaseApproval",
		zap.String("zcntoken", tokenAddress.String()),
		zap.String("spender", spenderAddress.String()),
		zap.Int64("amount", amount.Int64()),
	)

	tran, err := tokenInstance.IncreaseApproval(transactOpts, spenderAddress, amount)
	if err != nil {
		Logger.Error(
			"IncreaseApproval FAILED",
			zap.String("zcntoken", tokenAddress.String()),
			zap.String("spender", spenderAddress.String()),
			zap.Int64("amount", amount.Int64()),
			zap.Error(err))

		return nil, errors.Wrapf(err, "failed to send `IncreaseApproval` transaction")
	}

	Logger.Info(
		"Posted IncreaseApproval",
		zap.String("hash", tran.Hash().String()),
		zap.String("zcntoken", tokenAddress.String()),
		zap.String("spender", spenderAddress.String()),
		zap.Int64("amount", amount.Int64()),
	)

	return tran, nil
}

// GetTokenBalance returns balance of the current client for the zcntoken address
func (b *BridgeClient) GetTokenBalance() (*big.Int, error) {
	// 1. Token address parameter
	of := common.HexToAddress(b.TokenAddress)

	// 2. User's Ethereum wallet address parameter
	from := common.HexToAddress(b.EthereumAddress)

	tokenInstance, err := zcntoken.NewToken(of, b.ethereumClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize zcntoken instance")
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

// ResetUserNonceMinted Resets nonce for a specified Ethereum address
func (b *BridgeClient) ResetUserNonceMinted(ctx context.Context) (*types.Transaction, error) {
	bridgeInstance, transactOpts, err := b.prepareBridge(ctx, b.EthereumAddress, "resetUserNonceMinted")
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare bridge")
	}

	tran, err := bridgeInstance.ResetUserNonceMinted(transactOpts)
	if err != nil {
		Logger.Error("ResetUserNonceMinted FAILED", zap.Error(err))
		msg := "failed to execute ResetUserNonceMinted call, ethereumAddress = %s"
		return nil, errors.Wrapf(err, msg, b.EthereumAddress)
	}

	Logger.Info(
		"Posted ResetUserMintedNonce",
		zap.String("ethereumWallet", b.EthereumAddress),
	)

	return tran, err
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

// FetchZCNToETHRate retrieves latest ZCN to ETH rate using Bancor API
func (b *BridgeClient) FetchZCNToSourceTokenRate(sourceTokenAddress string) (*big.Float, error) {
	client = h.CleanClient()

	resp, err := client.Get(fmt.Sprintf("%s/tokens?dlt_id=%s", b.BancorAPIURL, b.TokenAddress))
	if err != nil {
		return nil, err
	}

	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var bancorTokenDetails *BancorTokenDetails
	err = json.Unmarshal(body, &bancorTokenDetails)
	if err != nil {
		return nil, err
	}

	var zcnSourceTokenRateFloat float64

	switch sourceTokenAddress {
	case SourceTokenETHAddress:
		zcnSourceTokenRateFloat, err = strconv.ParseFloat(bancorTokenDetails.Data.Rate.ETH, 64)
	case SourceTokenBNTAddress:
		zcnSourceTokenRateFloat, err = strconv.ParseFloat(bancorTokenDetails.Data.Rate.BNT, 64)
	case SourceTokenUSDCAddress:
		zcnSourceTokenRateFloat, err = strconv.ParseFloat(bancorTokenDetails.Data.Rate.USDC, 64)
	case SourceTokenEURCAddress:
		zcnSourceTokenRateFloat, err = strconv.ParseFloat(bancorTokenDetails.Data.Rate.EURC, 64)
	}

	if err != nil {
		return nil, err
	}

	return big.NewFloat(zcnSourceTokenRateFloat), nil
}

// GetMaxBancorTargetAmount retrieves max amount of a given source token for Bancor swap
func (b *BridgeClient) GetMaxBancorTargetAmount(sourceTokenAddress string, amountSwap uint64) (*big.Int, error) {
	amountSwapZCN, err := currency.Coin(amountSwap).ToZCN()
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert current zcntoken balance to ZCN")
	}

	var zcnEthRate *big.Float
	zcnEthRate, err = b.FetchZCNToSourceTokenRate(sourceTokenAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve ZCN to source zcntoken rate using Bancor API")
	}

	zcnEthRateFloat, _ := zcnEthRate.Float64()

	return big.NewInt(int64(amountSwapZCN * zcnEthRateFloat * 1.5 * 1e18)), nil
}

// ApproveSwap provides opportunity to approve swap operation for ERC20 tokens
func (b *BridgeClient) ApproveSwap(ctx context.Context, sourceTokenAddress string, maxAmountSwap *big.Int) (*types.Transaction, error) {
	// 1. Token source token address parameter
	tokenAddress := common.HexToAddress(sourceTokenAddress)

	// 2. Spender source token address parameter
	spender := common.HexToAddress(BancorNetworkAddress)

	bancorTokenInstance, transactOpts, err := b.prepareBancorToken(ctx, "approve", tokenAddress, spender, maxAmountSwap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare bancor token")
	}

	Logger.Info(
		"Starting ApproveSwap",
		zap.Int64("amount", maxAmountSwap.Int64()),
		zap.String("spender", spender.String()),
	)

	tran, err := bancorTokenInstance.Approve(transactOpts, spender, maxAmountSwap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute Approve transaction")
	}

	return tran, nil
}

// Swap provides opportunity to perform zcntoken swap operation.
func (b *BridgeClient) Swap(ctx context.Context, sourceTokenAddress string, amountSwap uint64, maxAmountSwap *big.Int, deadlinePeriod time.Time) (*types.Transaction, error) {
	// 1. Swap amount parameter.
	amount := big.NewInt(int64(amountSwap))

	// 2. User's Ethereum wallet address.
	beneficiary := common.HexToAddress(b.EthereumAddress)

	// 3. Trade deadline
	deadline := big.NewInt(deadlinePeriod.Unix())

	// 4. Value of the Ethereum transaction
	var value *big.Int

	if sourceTokenAddress == SourceTokenETHAddress {
		value = maxAmountSwap
	} else {
		value = big.NewInt(0)
	}

	// 6. Source zcntoken address parameter
	from := common.HexToAddress(sourceTokenAddress)

	// 7. Target zcntoken address parameter
	to := common.HexToAddress(b.TokenAddress)

	bancorInstance, transactOpts, err := b.prepareBancor(ctx, value, "tradeByTargetAmount", from, to, amount, maxAmountSwap, deadline, beneficiary)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare bancornetwork")
	}

	Logger.Info(
		"Starting Swap",
		zap.Int64("amount", amount.Int64()),
		zap.String("sourceToken", sourceTokenAddress),
	)

	tran, err := bancorInstance.TradeByTargetAmount(transactOpts, from, to, amount, maxAmountSwap, deadline, beneficiary)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute TradeByTargetAmount transaction")
	}

	return tran, nil
}

func (b *BridgeClient) prepareBancor(ctx context.Context, value *big.Int, method string, params ...interface{}) (*bancornetwork.Bancor, *bind.TransactOpts, error) {
	// 1. Bancor network smart contract address
	contractAddress := common.HexToAddress(BancorNetworkAddress)

	abi, err := bancornetwork.BancorMetaData.GetAbi()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get bancornetwork abi")
	}

	pack, err := abi.Pack(method, params...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to pack arguments")
	}

	from := common.HexToAddress(b.EthereumAddress)

	opts := eth.CallMsg{
		To:   &contractAddress,
		From: from,
		Data: pack,
	}

	if value.Int64() != 0 {
		opts.Value = value
	}

	gasLimitUnits, err := b.ethereumClient.EstimateGas(ctx, opts)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to estimate gas limit")
	}

	gasLimitUnits = addPercents(gasLimitUnits, 10).Uint64()

	transactOpts := b.CreateSignedTransactionFromKeyStore(b.ethereumClient, gasLimitUnits)
	if value.Int64() != 0 {
		transactOpts.Value = value
	}

	bancorInstance, err := bancornetwork.NewBancor(contractAddress, b.ethereumClient)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to initialize bancornetwork instance")
	}

	return bancorInstance, transactOpts, nil
}

func (b *BridgeClient) prepareBancorToken(ctx context.Context, method string, tokenAddress common.Address, params ...interface{}) (*bancortoken.Bancortoken, *bind.TransactOpts, error) {
	abi, err := zcntoken.TokenMetaData.GetAbi()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get zcntoken abi")
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

	bancorTokenInstance, err := bancortoken.NewBancortoken(tokenAddress, b.ethereumClient)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to initialize zcntoken instance")
	}

	return bancorTokenInstance, transactOpts, nil
}

func (b *BridgeClient) prepareToken(ctx context.Context, method string, tokenAddress common.Address, params ...interface{}) (*zcntoken.Token, *bind.TransactOpts, error) {
	abi, err := zcntoken.TokenMetaData.GetAbi()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get zcntoken abi")
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

	tokenInstance, err := zcntoken.NewToken(tokenAddress, b.ethereumClient)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to initialize zcntoken instance")
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

// getProviderType validates the provider url and exposes pre-defined type definition.
func (b *BridgeClient) getProviderType() int {
	if strings.Contains(b.EthereumNodeURL, "g.alchemy.com") {
		return AlchemyProvider
	} else if strings.Contains(b.EthereumNodeURL, "rpc.tenderly.co") {
		return TenderlyProvider
	} else {
		return UnknownProvider
	}
}

// estimateTenderlyGasAmount performs gas amount estimation for the given transaction using Tenderly provider.
func (b *BridgeClient) estimateTenderlyGasAmount(ctx context.Context, from, to string, value int64) (float64, error) {
	return 8000000, nil
}

// estimateAlchemyGasAmount performs gas amount estimation for the given transaction using Alchemy provider
func (b *BridgeClient) estimateAlchemyGasAmount(ctx context.Context, from, to string, value int64) (float64, error) {
	client := jsonrpc.NewClient(b.EthereumNodeURL)

	valueHex := ConvertIntToHex(value)

	resp, err := client.Call(ctx, "eth_estimateGas", &AlchemyGasEstimationRequest{
		From: from, To: to, Value: valueHex})
	if err != nil {
		return 0, errors.Wrap(err, "gas price estimation failed")
	}

	if resp.Error != nil {
		return 0, errors.Wrap(errors.New(resp.Error.Error()), "gas price estimation failed")
	}

	gasAmountRaw, ok := resp.Result.(string)
	if !ok {
		return 0, errors.New("failed to parse gas amount")
	}

	gasAmountInt := new(big.Float)
	gasAmountInt.SetString(gasAmountRaw)

	gasAmountFloat, _ := gasAmountInt.Float64()

	return gasAmountFloat, nil
}

// EstimateGasAmount performs gas amount estimation for the given transaction.
func (b *BridgeClient) EstimateGasAmount(ctx context.Context, from, to string, value int64) (float64, error) {
	switch b.getProviderType() {
	case AlchemyProvider:
		return b.estimateAlchemyGasAmount(ctx, from, to, value)
	case TenderlyProvider:
		return b.estimateTenderlyGasAmount(ctx, from, to, value)
	}

	return 0, errors.New("used json-rpc does not allow to estimate gas amount")
}

// estimateTenderlyGasPrice performs gas estimation for the given transaction using Tenderly API.
func (b *BridgeClient) estimateTenderlyGasPrice(ctx context.Context, from, to string, value int64) (float64, error) {
	return 0, nil
}

// estimateAlchemyGasPrice performs gas estimation for the given transaction using Alchemy enhanced API returning
// approximate final gas fee.
func (b *BridgeClient) estimateAlchemyGasPrice(ctx context.Context, from, to string, value int64) (float64, error) {
	client := jsonrpc.NewClient(b.EthereumNodeURL)

	resp, err := client.Call(ctx, "eth_gasPrice")
	if err != nil {
		return 0, errors.Wrap(err, "gas price estimation failed")
	}

	if resp.Error != nil {
		return 0, errors.Wrap(errors.New(resp.Error.Error()), "gas price estimation failed")
	}

	gasPriceRaw, ok := resp.Result.(string)
	if !ok {
		return 0, errors.New("failed to parse gas price")
	}

	gasPriceInt := new(big.Float)
	gasPriceInt.SetString(gasPriceRaw)

	gasPriceFloat, _ := gasPriceInt.Float64()

	return gasPriceFloat, nil
}

// EstimateGasPrice performs gas estimation for the given transaction using Alchemy enhanced API returning
// approximate final gas fee.
func (b *BridgeClient) EstimateGasPrice(ctx context.Context, from, to string, value int64) (float64, error) {
	switch b.getProviderType() {
	case AlchemyProvider:
		return b.estimateAlchemyGasPrice(ctx, from, to, value)
	case TenderlyProvider:
		return b.estimateTenderlyGasPrice(ctx, from, to, value)
	}

	return 0, errors.New("used json-rpc does not allow to estimate gas price")
}
