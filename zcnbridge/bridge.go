package zcnbridge

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/0chain/gosdk/zcnbridge/ethereum/uniswapnetwork"
	"github.com/0chain/gosdk/zcnbridge/ethereum/uniswaprouter"

	"github.com/ybbus/jsonrpc/v3"

	"github.com/0chain/gosdk/zcnbridge/ethereum/zcntoken"
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

// CreateSignedTransactionFromKeyStore creates signed transaction from key store
// - client - Ethereum client
// - gasLimitUnits - gas limit in units
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
//   - ctx go context instance to run the transaction
//   - address Ethereum address of the authorizer
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
//   - ctx go context instance to run the transaction
//   - address Ethereum address of the authorizer
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

// AddEthereumAuthorizers add bridge authorizers to the Ethereum authorizers contract
// 		- configDir - configuration directory
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
//   - key key for the plan
//   - param plan id
func EncodePackInt64(key string, param int64) common.Hash {
	return crypto.Keccak256Hash(
		[]byte(key),
		common.LeftPadBytes(big.NewInt(param).Bytes(), 32),
	)
}

// NFTConfigSetUint256  sets a uint256 field in the NFT config, given the key as a string
//   - ctx go context instance to run the transaction
//   - key key for this field
//   - value value to set
func (b *BridgeClient) NFTConfigSetUint256(ctx context.Context, key string, value int64) (*types.Transaction, error) {
	kkey := crypto.Keccak256Hash([]byte(key))
	return b.NFTConfigSetUint256Raw(ctx, kkey, value)
}

// NFTConfigSetUint256Raw sets a uint256 field in the NFT config, given the key as a Keccak256 hash
//   - ctx go context instance to run the transaction
//   - key key for this field (hased)
//   - value value to set
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

// NFTConfigGetUint256 retrieves a uint256 field in the NFT config, given the key as a string
//   - ctx go context instance to run the transaction
//   - key key for this field
//   - keyParam additional key parameter, only the first item is used
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

// NFTConfigSetAddress sets an address field in the NFT config, given the key as a string
//   - ctx go context instance to run the transaction
//   - key key for this field
//   - address address to set
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

// NFTConfigGetAddress retrieves an address field in the NFT config, given the key as a string
//   - ctx go context instance to run the transaction
//   - key key for this field
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
//   - ctx go context instance to run the transaction
//   - allowanceAmount amount to increase
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
//   - ctx go context instance to run the transaction
//   - hash transaction hash
func (b *BridgeClient) VerifyZCNTransaction(ctx context.Context, hash string) (transaction.Transaction, error) {
	return transaction.Verify(ctx, hash)
}

// SignWithEthereumChain signs the digest with Ethereum chain signer taking key from the current user key storage
//   - message message to sign
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
//   - ctx go context instance to run the transaction
//   - rawEthereumAddress Ethereum address
func (b *BridgeClient) GetUserNonceMinted(ctx context.Context, rawEthereumAddress string) (*big.Int, error) {
	ethereumAddress := common.HexToAddress(rawEthereumAddress)

	contractAddress := common.HexToAddress(b.BridgeAddress)

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
//   - ctx go context instance to run the transaction
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
//   - ctx go context instance to run the transaction
//   - payload received from authorizers
//
// ERC20 signature: "mint(address,uint256,bytes,uint256,bytes[])"
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
//   - ctx go context instance to run the transaction
//   - amountTokens amount of tokens to burn
//
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
//   - ctx go context instance to run the transaction
//   - payload received from authorizers
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
//   - ctx go context instance to run the transaction
//   - amount amount of tokens to burn
//   - txnfee transaction fee
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

// ApproveUSDCSwap provides opportunity to approve swap operation for ERC20 tokens
//   - ctx go context instance to run the transaction
//   - source source amount
func (b *BridgeClient) ApproveUSDCSwap(ctx context.Context, source uint64) (*types.Transaction, error) {
	// 1. USDC token smart contract address
	tokenAddress := common.HexToAddress(UsdcTokenAddress)

	// 2. Swap source amount parameter.
	sourceInt := big.NewInt(int64(source))

	// 3. User's Ethereum wallet address parameter
	spenderAddress := common.HexToAddress(b.UniswapAddress)

	tokenInstance, transactOpts, err := b.prepareToken(ctx, "approve", tokenAddress, spenderAddress, sourceInt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare usdctoken")
	}

	Logger.Info(
		"Starting ApproveUSDCSwap",
		zap.String("usdctoken", tokenAddress.String()),
		zap.String("spender", spenderAddress.String()),
		zap.Int64("source", sourceInt.Int64()),
	)

	var tran *types.Transaction

	tran, err = tokenInstance.Approve(transactOpts, spenderAddress, sourceInt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute approve transaction")
	}

	return tran, nil
}

// GetETHSwapAmount retrieves ETH swap amount from the given source.
//   - ctx go context instance to run the transaction
//   - source source amount
func (b *BridgeClient) GetETHSwapAmount(ctx context.Context, source uint64) (*big.Int, error) {
	// 1. Uniswap smart contract address
	contractAddress := common.HexToAddress(UniswapRouterAddress)

	// 2. User's Ethereum wallet address parameter
	from := common.HexToAddress(b.EthereumAddress)

	// 3. Swap source amount parameter.
	sourceInt := big.NewInt(int64(source))

	// 3. Swap path parameter.
	path := []common.Address{
		common.HexToAddress(WethTokenAddress),
		common.HexToAddress(b.TokenAddress)}

	uniswapRouterInstance, err := uniswaprouter.NewUniswaprouter(contractAddress, b.ethereumClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize uniswaprouter instance")
	}

	Logger.Info(
		"Starting GetETHSwapAmount",
		zap.Uint64("source", source))

	var result []*big.Int

	result, err = uniswapRouterInstance.GetAmountsIn(&bind.CallOpts{From: from}, sourceInt, path)
	if err != nil {
		Logger.Error("GetAmountsIn FAILED", zap.Error(err))
		msg := "failed to execute GetAmountsIn call, ethereumAddress = %s"

		return nil, errors.Wrapf(err, msg, from)
	}

	return result[0], nil
}

// SwapETH provides opportunity to perform zcn token swap operation using ETH as source token.
//   - ctx go context instance to run the transaction
//   - source source amount
//   - target target amount
func (b *BridgeClient) SwapETH(ctx context.Context, source uint64, target uint64) (*types.Transaction, error) {
	// 1. Swap source amount parameter.
	sourceInt := big.NewInt(int64(source))

	// 2. Swap target amount parameter.
	targetInt := big.NewInt(int64(target))

	uniswapNetworkInstance, transactOpts, err := b.prepareUniswapNetwork(
		ctx, sourceInt, "swapETHForZCNExactAmountOut", targetInt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare uniswapnetwork")
	}

	Logger.Info(
		"Starting SwapETH",
		zap.Uint64("source", source),
		zap.Uint64("target", target))

	var tran *types.Transaction

	tran, err = uniswapNetworkInstance.SwapETHForZCNExactAmountOut(transactOpts, targetInt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute swapETHForZCNExactAmountOut transaction")
	}

	return tran, nil
}

// SwapUSDC provides opportunity to perform zcn token swap operation using USDC as source token.
//   - ctx go context instance to run the transaction
//   - source source amount
//   - target target amount
func (b *BridgeClient) SwapUSDC(ctx context.Context, source uint64, target uint64) (*types.Transaction, error) {
	// 1. Swap target amount parameter.
	sourceInt := big.NewInt(int64(source))

	// 2. Swap source amount parameter.
	targetInt := big.NewInt(int64(target))

	uniswapNetworkInstance, transactOpts, err := b.prepareUniswapNetwork(
		ctx, big.NewInt(0), "swapUSDCForZCNExactAmountOut", targetInt, sourceInt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare uniswapnetwork")
	}

	Logger.Info(
		"Starting SwapUSDC",
		zap.Uint64("source", source),
		zap.Uint64("target", target))

	var tran *types.Transaction

	tran, err = uniswapNetworkInstance.SwapUSDCForZCNExactAmountOut(transactOpts, targetInt, sourceInt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute swapUSDCForZCNExactAmountOut transaction")
	}

	return tran, nil
}

// prepareUniswapNetwork performs uniswap network smart contract preparation actions.
func (b *BridgeClient) prepareUniswapNetwork(ctx context.Context, value *big.Int, method string, params ...interface{}) (*uniswapnetwork.Uniswap, *bind.TransactOpts, error) {
	// 1. Uniswap smart contract address
	contractAddress := common.HexToAddress(b.UniswapAddress)

	// 2. To address parameter.
	to := common.HexToAddress(b.TokenAddress)

	// 3. From address parameter.
	from := common.HexToAddress(b.EthereumAddress)

	abi, err := uniswapnetwork.UniswapMetaData.GetAbi()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get uniswaprouter abi")
	}

	var pack []byte

	pack, err = abi.Pack(method, params...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to pack arguments")
	}

	opts := eth.CallMsg{
		To:   &to,
		From: from,
		Data: pack,
	}

	if value.Int64() != 0 {
		opts.Value = value
	}

	transactOpts := b.CreateSignedTransactionFromKeyStore(b.ethereumClient, 0)
	if value.Int64() != 0 {
		transactOpts.Value = value
	}

	var uniswapNetworkInstance *uniswapnetwork.Uniswap

	uniswapNetworkInstance, err = uniswapnetwork.NewUniswap(contractAddress, b.ethereumClient)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to initialize uniswapnetwork instance")
	}

	return uniswapNetworkInstance, transactOpts, nil
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

	var tokenInstance *zcntoken.Token

	tokenInstance, err = zcntoken.NewToken(tokenAddress, b.ethereumClient)
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
func (b *BridgeClient) estimateAlchemyGasAmount(ctx context.Context, from, to, data string, value int64) (float64, error) {
	client := jsonrpc.NewClient(b.EthereumNodeURL)

	valueHex := ConvertIntToHex(value)

	resp, err := client.Call(ctx, "eth_estimateGas", &AlchemyGasEstimationRequest{
		From:  from,
		To:    to,
		Value: valueHex,
		Data:  data})
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

// EstimateBurnWZCNGasAmount performs gas amount estimation for the given wzcn burn transaction.
//   - ctx go context instance to run the transaction
//   - from source address
//   - to target address
//   - amountTokens amount of tokens to burn
func (b *BridgeClient) EstimateBurnWZCNGasAmount(ctx context.Context, from, to, amountTokens string) (float64, error) {
	switch b.getProviderType() {
	case AlchemyProvider:
		abi, err := bridge.BridgeMetaData.GetAbi()
		if err != nil {
			return 0, errors.Wrap(err, "failed to get ABI")
		}

		clientID := DefaultClientIDEncoder(zcncore.GetClientWalletID())

		amount := new(big.Int)
		amount.SetString(amountTokens, 10)

		var packRaw []byte
		packRaw, err = abi.Pack("burn", amount, clientID)
		if err != nil {
			return 0, errors.Wrap(err, "failed to pack arguments")
		}

		pack := "0x" + hex.EncodeToString(packRaw)

		return b.estimateAlchemyGasAmount(ctx, from, to, pack, 0)
	case TenderlyProvider:
		return b.estimateTenderlyGasAmount(ctx, from, to, 0)
	}

	return 0, errors.New("used json-rpc does not allow to estimate gas amount")
}

// EstimateMintWZCNGasAmount performs gas amount estimation for the given wzcn mint transaction.
//   - ctx go context instance to run the transaction
//   - from source address
//   - to target address
//   - zcnTransactionRaw zcn transaction (hashed)
//   - amountToken amount of tokens to mint
//   - nonceRaw nonce
//   - signaturesRaw authorizer signatures
func (b *BridgeClient) EstimateMintWZCNGasAmount(
	ctx context.Context, from, to, zcnTransactionRaw, amountToken string, nonceRaw int64, signaturesRaw [][]byte) (float64, error) {
	switch b.getProviderType() {
	case AlchemyProvider:
		amount := new(big.Int)
		amount.SetString(amountToken, 10)

		zcnTransaction := DefaultClientIDEncoder(zcnTransactionRaw)

		nonce := new(big.Int)
		nonce.SetInt64(nonceRaw)

		fromRaw := common.HexToAddress(from)

		abi, err := bridge.BridgeMetaData.GetAbi()
		if err != nil {
			return 0, errors.Wrap(err, "failed to get ABI")
		}

		var packRaw []byte
		packRaw, err = abi.Pack("mint", fromRaw, amount, zcnTransaction, nonce, signaturesRaw)
		if err != nil {
			return 0, errors.Wrap(err, "failed to pack arguments")
		}

		pack := "0x" + hex.EncodeToString(packRaw)

		return b.estimateAlchemyGasAmount(ctx, from, to, pack, 0)
	case TenderlyProvider:
		return b.estimateTenderlyGasAmount(ctx, from, to, 0)
	}

	return 0, errors.New("used json-rpc does not allow to estimate gas amount")
}

// estimateTenderlyGasPrice performs gas estimation for the given transaction using Tenderly API.
func (b *BridgeClient) estimateTenderlyGasPrice(ctx context.Context) (float64, error) {
	return 1, nil
}

// estimateAlchemyGasPrice performs gas estimation for the given transaction using Alchemy enhanced API returning
// approximate final gas fee.
func (b *BridgeClient) estimateAlchemyGasPrice(ctx context.Context) (float64, error) {
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

// EstimateGasPrice performs gas estimation for the given transaction.
//   - ctx go context instance to run the transaction
func (b *BridgeClient) EstimateGasPrice(ctx context.Context) (float64, error) {
	switch b.getProviderType() {
	case AlchemyProvider:
		return b.estimateAlchemyGasPrice(ctx)
	case TenderlyProvider:
		return b.estimateTenderlyGasPrice(ctx)
	}

	return 0, errors.New("used json-rpc does not allow to estimate gas price")
}
