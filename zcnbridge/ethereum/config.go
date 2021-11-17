package ethereum

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"

	"github.com/0chain/gosdk/zcnbridge/config"

	"github.com/0chain/gosdk/zcncore"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	etherCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
)

// Required config
// 1. For SC REST we need only miners address, will take it from network - init from config
// 2. For SC method we need to run local chain to run minting and burning
// 3. For Ethereum, we will take params from

//  _allowances[owner][spender] = amount;
// as a spender, ERC20 WZCN token must increase allowance for the bridge to make burn on behalf of WZCN owner

type EthWalletInfo struct {
	ID         string `json:"ID"`
	PrivateKey string `json:"PrivateKey"`
}

func GetEthereumWalletInfo() (*EthWalletInfo, error) {
	ownerWallet, err := zcncore.GetWalletAddrFromEthMnemonic(config.Bridge.Mnemonic)
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

func CreateEthClient() (*ethclient.Client, error) {
	client, err := ethclient.Dial(config.Bridge.EthereumNodeURL)
	if err != nil {
		zcncore.Logger.Error(err)
	}
	return client, err
}

func GetKeysAddress() (common.Address, *ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	ownerWalletInfo, err := GetEthereumWalletInfo()
	if err != nil {
		return [20]byte{}, nil, nil, errors.Wrap(err, "failed to fetch wallet ownerWalletInfo")
	}

	privateKeyECDSA, err := etherCrypto.HexToECDSA(ownerWalletInfo.PrivateKey)
	if err != nil {
		return [20]byte{}, nil, nil, errors.Wrap(err, "failed to read private key")
	}

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		zcncore.Logger.Fatal("error casting public key to ECDSA")
	}

	ownerAddress := etherCrypto.PubkeyToAddress(*publicKeyECDSA)

	return ownerAddress, publicKeyECDSA, privateKeyECDSA, nil
}

func CreateSignedTransaction(
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

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(int64(config.Bridge.ChainID)))
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	valueWei := new(big.Int).Mul(big.NewInt(config.Bridge.Value), big.NewInt(params.Wei))

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = valueWei         // in wei
	auth.GasLimit = gasLimitUnits // in units
	auth.GasPrice = gasPriceWei   // wei

	return auth
}
