package zcnbridge

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"github.com/0chain/gosdk/zcncore"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	etherCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
	"math/big"
)

//  _allowances[owner][spender] = amount;
// as a spender, ERC20 WZCN token must increase allowance for the bridge to make burn on behalf of WZCN owner

type bridgeConfig struct {
	mnemonic string // owner mnemonic
	//ownerAddress  common.Address
	//publicKey     crypto.PublicKey
	bridgeAddress string
	wzcnAddress   string
	nodeURL       string
	chainID       int
	gasLimit      int
	value         int
}

type ethWalletInfo struct {
	ID         string `json:"ID"`
	PrivateKey string `json:"PrivateKey"`
}

var (
	config bridgeConfig
)

func getOwnerWalletInfo() (*ethWalletInfo, error) {
	ownerWallet, err := zcncore.GetWalletAddrFromEthMnemonic(config.mnemonic)
	if err != nil {
		err = errors.Wrap(err, "failed to initialize wallet from mnemonic")
	}

	wallet := &ethWalletInfo{}
	err = json.Unmarshal([]byte(ownerWallet), wallet)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal wallet info")
	}

	return wallet, err
}

func createClient() (*ethclient.Client, error) {
	client, err := ethclient.Dial(config.nodeURL)
	if err != nil {
		zcncore.Logger.Error(err)
	}
	return client, err
}

func ownerPrivateKeyAndAddress() (common.Address, *ecdsa.PrivateKey, error) {
	ownerWalletInfo, err := getOwnerWalletInfo()
	if err != nil {
		return [20]byte{}, nil, errors.Wrap(err, "failed to fetch wallet ownerWalletInfo")
	}

	privateKey, err := etherCrypto.HexToECDSA(ownerWalletInfo.PrivateKey)
	if err != nil {
		return [20]byte{}, nil, errors.Wrap(err, "failed to read private key")
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		zcncore.Logger.Fatal("error casting public key to ECDSA")
	}

	ownerAddress := etherCrypto.PubkeyToAddress(*publicKeyECDSA)

	return ownerAddress, privateKey, nil
}

func createSignedTransaction(
	client *ethclient.Client,
	fromAddress common.Address,
	privateKey *ecdsa.PrivateKey,
	gasLimit uint64,
) *bind.TransactOpts {
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	// eth_estimateGas
	// EstimateGas tries to estimate the gas needed to execute a specific transaction based on
	// the current pending state of the backend blockchain. There is no guarantee that this is
	// the true gas limit requirement as other transactions may be added or removed by miners,
	// but it should provide a basis for setting a reasonable default.

	// eth_gasPrice
	// retrieves the currently suggested gas price to allow a timely
	// execution of a transaction

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(int64(config.chainID)))
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	value := new(big.Int).Mul(big.NewInt(int64(config.value)), big.NewInt(params.Wei))

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = value       // in wei
	auth.GasLimit = gasLimit // in units
	auth.GasPrice = gasPrice

	return auth
}
