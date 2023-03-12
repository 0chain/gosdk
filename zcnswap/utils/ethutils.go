package utils

import (
	"context"
	"crypto/ecdsa"
	"github.com/0chain/errors"
	l "github.com/0chain/gosdk/zboxcore/logger"
	"github.com/0chain/gosdk/zcnbridge"
	hdwallet "github.com/0chain/gosdk/zcncore/ethhdwallet"
	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	cmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/tyler-smith/go-bip39"
	"go.uber.org/zap"
	"math/big"
	"time"
)

const (
	STATUS_FAIL    = 1
	STATUS_SUCCESS = 0
)

func ConfirmEthereumTransaction(hash string, times int, duration time.Duration, client *ethclient.Client) (int, error) {
	var (
		res = 0
	)

	if hash == "" {
		return -1, errors.New("500", "transaction hash should not be empty")
	}

	l.Logger.Info("Start transaction check", zap.Any("hash", hash))
	for i := 0; i < times; i++ {
		res := CheckEthHashStatus(hash, client)
		if res == STATUS_SUCCESS || res == STATUS_FAIL {
			break
		}
		time.Sleep(duration)
	}
	return res, nil
}

// CheckEthHashStatus - checking the status of ETH transaction
// possible values 0 (fail) or 1 (success)
func CheckEthHashStatus(hash string, client *ethclient.Client) int {
	txHash := cmn.HexToHash(hash)

	tx, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return -1
	}
	return int(tx.Status)
}

func NewSignedTransaction(pack []byte, from, to string, value *big.Int, privateKey *ecdsa.PrivateKey, client *ethclient.Client) (*bind.TransactOpts, error) {
	fromAddress := cmn.HexToAddress(from)
	toAddress := cmn.HexToAddress(to)
	gasLimitUnits, err := client.EstimateGas(context.Background(), eth.CallMsg{
		From: fromAddress,
		To:   &toAddress,
		Data: pack,
	})

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}

	gasPriceWei, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	opts := zcnbridge.CreateSignedTransaction(chainID, client, fromAddress, privateKey, gasLimitUnits)
	valueWei := new(big.Int).Mul(value, big.NewInt(params.Wei))

	opts.Nonce = big.NewInt(int64(nonce))
	opts.Value = valueWei         // in wei
	opts.GasLimit = gasLimitUnits // in units
	opts.GasPrice = gasPriceWei   // wei

	return opts, nil
}

func AddPercents(gasLimitUnits uint64, percents int) *big.Int {
	gasLimitBig := big.NewInt(int64(gasLimitUnits))
	factorBig := big.NewInt(int64(percents))
	deltaBig := gasLimitBig.Div(gasLimitBig, factorBig)

	origin := big.NewInt(int64(gasLimitUnits))
	gasLimitBig = origin.Add(origin, deltaBig)

	return gasLimitBig
}

func CreateHDWallet() (*accounts.Account, string, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, "", err
	}

	walletHd, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, "", err
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := walletHd.Derive(path, false)
	if err != nil {
		return nil, "", err
	}

	return &account, mnemonic, err
}
