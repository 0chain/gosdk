package zcncore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/0chain/gosdk/core/zcncrypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"math"
	"math/big"
	"regexp"
)

func TokensToEth(tokens int64) float64 {
	fbalance := new(big.Float)
	fbalance.SetString(string(tokens))
	ethValue, _ := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18))).Float64()
	return ethValue
}

func GetWalletAddrFromEthMnemonic(mnemonic string) (string, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return "", err
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, false)
	if err != nil {
		return "", err
	}

	privKey, err := wallet.PrivateKeyHex(account)
	if err != nil {
		return "", err
	}

	type ethWalletinfo struct {
		ID         string `json:"ID"`
		PrivateKey string `json:"PrivateKey"`
	}

	res, err := json.Marshal(ethWalletinfo{ID: account.Address.Hex(), PrivateKey: privKey})
	return string(res), err
}

func GetEthBalance(ethAddr string, cb GetBalanceCallback) error {
	go func() {
		value, err := getBalanceFromEthNode(ethAddr)
		if err != nil {
			Logger.Error(err)
			cb.OnBalanceAvailable(StatusError, 0, "")
			return
		}
		cb.OnBalanceAvailable(StatusSuccess, value, "")
	}()
	return nil
}

func IsValidEthAddress(ethAddr string) (bool, error) {
	if len(_config.chain.EthNode) == 0 {
		return false, fmt.Errorf("Eth node SDK not initialized.")
	}

	Logger.Info("requesting from", _config.chain.EthNode)
	client, err := ethclient.Dial(_config.chain.EthNode)
	if err != nil {
		return false, err
	}

	return isValidEthAddress(ethAddr, client)
}

func isValidEthAddress(ethAddr string, client *ethclient.Client) (bool, error) {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	if !re.MatchString(ethAddr) {
		return false, fmt.Errorf("regex error")
	}

	address := common.HexToAddress(ethAddr)
	bytecode, err := client.CodeAt(context.Background(), address, nil) // nil is latest block
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}
	isContract := len(bytecode) > 0
	return isContract, nil
}

func getBalanceFromEthNode(ethAddr string) (int64, error) {
	if client, err := getEthClient(); err == nil {
		res, err := isValidEthAddress(ethAddr, client)
		if !res {
			return 0, err
		}

		account := common.HexToAddress(ethAddr)
		Logger.Info("for eth address", account)
		balance, err := client.BalanceAt(context.Background(), account, nil)
		if err != nil {
			return 0, err
		}

		Logger.Info("balance", balance.String())

		return balance.Int64(), nil
	} else {
		return 0, err
	}
}

func getEthClient() (*ethclient.Client, error) {
	if len(_config.chain.EthNode) == 0 {
		return nil, fmt.Errorf("Eth node SDK not initialized.")
	}

	Logger.Info("requesting from", _config.chain.EthNode)
	client, err := ethclient.Dial(_config.chain.EthNode)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func CreateWalletFromEthMnemonic(mnemonic, password string, statusCb WalletCallback) error {
	if len(_config.chain.Miners) < 1 || len(_config.chain.Sharders) < 1 {
		return fmt.Errorf("SDK not initialized")
	}
	go func() {
		sigScheme := zcncrypto.NewSignatureScheme(_config.chain.SignatureScheme)
		wallet, err := sigScheme.GenerateKeysWithEth(mnemonic, password)
		if err != nil {
			statusCb.OnWalletCreateComplete(StatusError, "", fmt.Sprintf("%s", err.Error()))
			return
		}
		err = RegisterToMiners(wallet, statusCb)
		if err != nil {
			statusCb.OnWalletCreateComplete(StatusError, "", fmt.Sprintf("%s", err.Error()))
			return
		}
	}()
	return nil
}

func CheckEthHashStatus(hash string) int {
	txHash := common.HexToHash(hash)

	var client *ethclient.Client
	var err error
	if client, err = getEthClient(); err != nil {
		return  -1
	}

	tx, err := client.TransactionReceipt(context.Background(), txHash)
	if err !=nil {
		return -1
	}
	return int(tx.Status)
}
