package zcnbridge

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"path"
	"time"

	"github.com/ethereum/go-ethereum/accounts"

	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/0chain/gosdk/zcncore"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
)

func (b *EthereumConfig) CreateEthClient() (*ethclient.Client, error) {
	client, err := ethclient.Dial(b.EthereumNodeURL)
	if err != nil {
		zcncore.Logger.Error(err)
	}
	return client, err
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
	client *ethclient.Client,
	signerAddress common.Address,
	gasLimitUnits uint64,
	password string,
	value int64,
) *bind.TransactOpts {
	keyDir := path.Join(GetConfigDir(), "wallets")
	ks := keystore.NewKeyStore(keyDir, keystore.StandardScryptN, keystore.StandardScryptP)
	signer := accounts.Account{
		Address: signerAddress,
	}
	signerAcc, err := ks.Find(signer)
	if err != nil {
		zcncore.Logger.Fatal(errors.Wrapf(err, "signer: %s", signerAddress.Hex()))
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		zcncore.Logger.Fatal(errors.Wrap(err, "failed to get chain ID"))
	}

	nonce, err := client.PendingNonceAt(context.Background(), signerAddress)
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	gasPriceWei, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	err = ks.TimedUnlock(signer, password, time.Second*2)
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	opts, err := bind.NewKeyStoreTransactorWithChainID(ks, signerAcc, chainID)
	if err != nil {
		zcncore.Logger.Fatal(err)
	}

	valueWei := new(big.Int).Mul(big.NewInt(value), big.NewInt(params.Wei))

	opts.Nonce = big.NewInt(int64(nonce))
	opts.Value = valueWei         // in wei
	opts.GasLimit = gasLimitUnits // in units
	opts.GasPrice = gasPriceWei   // wei

	return opts
}
