package zcnswap

import (
	"crypto/ecdsa"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/zcncore"
	hdwallet "github.com/0chain/gosdk/zcncore/ethhdwallet"
	"github.com/0chain/gosdk/zcnswap/config"
	"github.com/0chain/gosdk/zcnswap/swapfactory/bancor"
	"github.com/0chain/gosdk/zcnswap/swapfactory/erc20"
	"github.com/ethereum/go-ethereum/accounts"
	cmn "github.com/ethereum/go-ethereum/common"
	"math/big"
)

type ethAccount struct {
	PrivateKey *ecdsa.PrivateKey
	SourceAddr *cmn.Address
	Account    *accounts.Account
}

func Swap(swapAmount int64, tokenSource string) (string, error) {
	client, _ := zcncore.GetEthClient()

	ethAccount, err := getWallet()
	if err != nil {
		return "", err
	}

	targetTokenAddress := config.Configuration.ZcnTokenAddress
	sourceTokenAddress := tokenSource

	amount := new(big.Int).SetInt64(swapAmount)

	// checking for available funds
	balance, err := erc20.TokenBalance(*ethAccount.SourceAddr,
		cmn.HexToAddress(sourceTokenAddress),
		client)
	if err != nil {
		return "", err
	}
	if balance.Cmp(amount) == -1 {
		return "", errors.New("500", "Not enough balance")
	}

	bancorService := bancor.NewSwapService(client, ethAccount.PrivateKey)
	pair, err := bancorService.EstimateRate(sourceTokenAddress, targetTokenAddress, amount)
	if err != nil {
		return "", err
	}

	signedTx, err := bancorService.SwapWithConversionPath(pair,
		ethAccount.SourceAddr.Hex())
	if err != nil {
		return "", err
	}
	return signedTx.Hash().String(), nil
}

func getWallet() (acc *ethAccount, err error) {
	walletHd, err := hdwallet.NewFromMnemonic(config.Configuration.WalletMnemonic)
	if err != nil {
		return nil, err
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := walletHd.Derive(path, false)
	if err != nil {
		return nil, err
	}

	privateKey, err := walletHd.PrivateKey(account)
	if err != nil {
		return nil, err
	}
	sourceAddrHex, _ := walletHd.AddressHex(account)
	address := cmn.HexToAddress(sourceAddrHex)

	return &ethAccount{
		PrivateKey: privateKey,
		SourceAddr: &address,
		Account:    &account,
	}, err
}
