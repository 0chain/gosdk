package zcnbridge

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"path"

	"github.com/ethereum/go-ethereum/accounts"

	"github.com/ethereum/go-ethereum/accounts/keystore"

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

func (b *BridgeOwner) CreateEthereumWallet() (*EthereumWallet, error) {
	address, publicKey, privateKey, err := GetKeysAndAddressFromMnemonic(b.EthereumMnemonic)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize owner ethereum zcnWallet")
	}

	return &EthereumWallet{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Address:    address,
	}, nil
}

func (b *BridgeClient) CreateEthereumWallet() (*EthereumWallet, error) {
	address, publicKey, privateKey, err := GetKeysAndAddressFromMnemonic(b.EthereumMnemonic)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize client ethereum zcnWallet")
	}

	return &EthereumWallet{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Address:    address,
	}, nil
}

func (b *BridgeClient) GetEthereumWalletInfo() (*EthWalletInfo, error) {
	return GetEthereumWalletInfoFromMnemonic(b.EthereumMnemonic)
}

func (b *EthereumConfig) CreateEthClient() (*ethclient.Client, error) {
	client, err := ethclient.Dial(b.EthereumNodeURL)
	if err != nil {
		zcncore.Logger.Error(err)
	}
	return client, err
}

func (b *BridgeClient) GetKeysAndAddress() (common.Address, *ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	ownerWalletInfo, err := b.GetEthereumWalletInfo()
	if err != nil {
		return [20]byte{}, nil, nil, errors.Wrap(err, "failed to fetch zcnWallet ownerWalletInfo")
	}

	return GetKeysAndAddressFromPrivateKey(ownerWalletInfo.PrivateKey)
}

func CreateEthereumWalletFromMnemonic(mnemonic string) (*EthereumWallet, error) {
	address, publicKey, privateKey, err := GetKeysAndAddressFromMnemonic(mnemonic)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize ethereum zcnWallet")
	}

	return &EthereumWallet{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Address:    address,
	}, nil
}

func GetEthereumWalletInfoFromMnemonic(mnemonic string) (*EthWalletInfo, error) {
	ownerWallet, err := zcncore.GetWalletAddrFromEthMnemonic(mnemonic)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize zcnWallet from mnemonic")
	}

	wallet := &EthWalletInfo{}
	err = json.Unmarshal([]byte(ownerWallet), wallet)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal zcnWallet info")
	}

	return wallet, err
}

func GetKeysAndAddressFromPrivateKey(privateKey string) (common.Address, *ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
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

func GetKeysAndAddressFromMnemonic(mnemonic string) (common.Address, *ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	ownerWalletInfo, err := GetEthereumWalletInfoFromMnemonic(mnemonic)
	if err != nil {
		return [20]byte{}, nil, nil, errors.Wrap(err, "failed to fetch zcnWallet ownerWalletInfo")
	}

	return GetKeysAndAddressFromPrivateKey(ownerWalletInfo.PrivateKey)
}

//  _allowances[owner][spender] = amount;
// as a spender, ERC20 WZCN token must increase allowance for the bridge to make burn on behalf of WZCN owner

func CreateSignedTransaction(
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

	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	valueWei := new(big.Int).Mul(big.NewInt(0), big.NewInt(params.Wei))

	opts.Nonce = big.NewInt(int64(nonce))
	opts.Value = valueWei         // in wei
	opts.GasLimit = gasLimitUnits // in units
	opts.GasPrice = gasPriceWei   // wei

	return opts
}

func CreateSignedTransactionFromKeyStore(
	chainID *big.Int,
	client *ethclient.Client,
	fromAddress common.Address,
	signerAddress string,
	gasLimitUnits uint64,
) *bind.TransactOpts {
	keyDir := path.Join(GetConfigDir(), "wallets")
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	signer := accounts.Account{
		Address: common.HexToAddress(signerAddress),
	}

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	gasPriceWei, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	opts, err := bind.NewKeyStoreTransactorWithChainID(ks, signer, chainID)
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	valueWei := new(big.Int).Mul(big.NewInt(0), big.NewInt(params.Wei))

	opts.Nonce = big.NewInt(int64(nonce))
	opts.Value = valueWei         // in wei
	opts.GasLimit = gasLimitUnits // in units
	opts.GasPrice = gasPriceWei   // wei

	return opts
}
