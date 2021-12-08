package zcnbridge

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"

	"github.com/0chain/gosdk/zcncore"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
)

type EthWalletInfo struct {
	ID         string `json:"ID"`
	PrivateKey string `json:"PrivateKey"`
}

type EthereumWallet struct {
	PublicKey  *ecdsa.PublicKey
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address
}

func (b *Bridge) CreateEthereumWalletFromMnemonic(mnemonic string) (*EthereumWallet, error) {
	address, publicKey, privateKey, err := b.GetKeysAndAddressFromMnemonic(mnemonic)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize ethereum wallet")
	}

	return &EthereumWallet{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Address:    address,
	}, nil
}

func (b *Bridge) CreateEthereumWallet() (*EthereumWallet, error) {
	address, publicKey, privateKey, err := b.GetKeysAndAddress()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize ethereum wallet")
	}

	return &EthereumWallet{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Address:    address,
	}, nil
}

func (b *Bridge) GetEthereumWalletInfoFromMnemonic(mnemonic string) (*EthWalletInfo, error) {
	ownerWallet, err := zcncore.GetWalletAddrFromEthMnemonic(mnemonic)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize wallet from mnemonic")
	}

	wallet := &EthWalletInfo{}
	err = json.Unmarshal([]byte(ownerWallet), wallet)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal wallet info")
	}

	return wallet, err
}

func (b *Bridge) GetEthereumWalletInfo() (*EthWalletInfo, error) {
	return b.GetEthereumWalletInfoFromMnemonic(b.ClientEthereumMnemonic)
}

func (b *Bridge) CreateEthClient() (*ethclient.Client, error) {
	client, err := ethclient.Dial(b.EthereumNodeURL)
	if err != nil {
		zcncore.Logger.Error(err)
	}
	return client, err
}

func (b *Bridge) GetKeysAndAddressFromMnemonic(mnemonic string) (common.Address, *ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	ownerWalletInfo, err := b.GetEthereumWalletInfoFromMnemonic(mnemonic)
	if err != nil {
		return [20]byte{}, nil, nil, errors.Wrap(err, "failed to fetch wallet ownerWalletInfo")
	}

	return b.GetKeysAndAddressFromPrivateKey(ownerWalletInfo.PrivateKey)
}

func (b *Bridge) GetKeysAndAddress() (common.Address, *ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	ownerWalletInfo, err := b.GetEthereumWalletInfo()
	if err != nil {
		return [20]byte{}, nil, nil, errors.Wrap(err, "failed to fetch wallet ownerWalletInfo")
	}

	return b.GetKeysAndAddressFromPrivateKey(ownerWalletInfo.PrivateKey)
}

func (b *Bridge) GetKeysAndAddressFromPrivateKey(privateKey string) (common.Address, *ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return [20]byte{}, nil, nil, errors.Wrap(err, "failed to read private key")
	}

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		zcncore.Logger.Fatal("error casting public key to ECDSA")
	}

	ownerAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	return ownerAddress, publicKeyECDSA, privateKeyECDSA, nil
}

// Required config
// 1. For SC REST we need only miners address, will take it from network - init from config
// 2. For SC method we need to run local chain to run minting and burning
// 3. For Ethereum, we will take params from

//  _allowances[owner][spender] = amount;
// as a spender, ERC20 WZCN token must increase allowance for the bridge to make burn on behalf of WZCN owner

func (b *Bridge) CreateSignedTransaction(
	chainID *big.Int,
	client *ethclient.Client,
	fromAddress common.Address,
	privateKey *ecdsa.PrivateKey,
	gasLimitUnits uint64,
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

	gasPriceWei, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	valueWei := new(big.Int).Mul(big.NewInt(b.Value), big.NewInt(params.Wei))

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = valueWei         // in wei
	auth.GasLimit = gasLimitUnits // in units
	auth.GasPrice = gasPriceWei   // wei

	return auth
}
