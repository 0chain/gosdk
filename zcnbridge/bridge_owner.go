package zcnbridge

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/0chain/gosdk/zcnbridge/ethereum/authorizers"

	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

func (b *BridgeOwner) prepareAuthorizers(ctx context.Context, method string, params ...interface{}) (*authorizers.Authorizers, *bind.TransactOpts, error) {
	etherClient, err := b.CreateEthClient()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create etherClient")
	}

	// To (contract)
	contractAddress := common.HexToAddress(b.AuthorizersAddress)

	// BridgeClient Ethereum Wallet
	ethereumWallet := b.GetEthereumWallet()
	if ethereumWallet == nil {
		return nil, nil, errors.New("BridgeClient Ethereum zcnWallet is not initialized")
	}

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

	// Gas limits in units
	gasLimitUnits, err := etherClient.EstimateGas(ctx, eth.CallMsg{
		To:   &contractAddress,
		From: ethereumWallet.Address,
		Data: pack,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to estimate gas")
	}

	// Update gas limits + 10%
	gasLimitUnits = addPercents(gasLimitUnits, 10).Uint64()
	chainID, err := etherClient.ChainID(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get chain ID")
	}

	// Create options
	transactOpts := CreateSignedTransaction(
		chainID,
		etherClient,
		ethereumWallet.Address,
		ethereumWallet.PrivateKey,
		gasLimitUnits,
	)

	// Authorizers instance
	authorizersInstance, err := authorizers.NewAuthorizers(contractAddress, etherClient)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create authorizers instance")
	}

	return authorizersInstance, transactOpts, nil
}

// AddEthereumAuthorizer Adds authorizer to Ethereum bridge. Only contract deployer can call this method
func (b *BridgeOwner) AddEthereumAuthorizer(ctx context.Context, address common.Address) (*types.Transaction, error) {
	instance, transactOpts, err := b.prepareAuthorizers(ctx, "addAuthorizers", address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare bridge")
	}

	tran, err := instance.AddAuthorizers(transactOpts, address)
	if err != nil {
		msg := "failed to execute BurnZCN transaction to ClientID = %s with amount = %s"
		return nil, errors.Wrapf(err, msg, b.ID(), address.String())
	}

	return tran, err
}

func (b *BridgeOwner) AddEthereumAuthorizers(configDir string) {
	cfg := viper.New()
	cfg.AddConfigPath(configDir)
	cfg.SetConfigName("cfg")
	if err := cfg.ReadInConfig(); err != nil {
		fmt.Println(err)
		return
	}

	mnemonics := cfg.GetStringSlice("cfg")

	for _, mnemonic := range mnemonics {
		wallet, err := CreateEthereumWalletFromMnemonic(mnemonic)
		if err != nil {
			fmt.Println(err)
			continue
		}

		transaction, err := b.AddEthereumAuthorizer(context.TODO(), wallet.Address)
		if err != nil || transaction == nil {
			fmt.Println(err)
			continue
		}

		status, err := ConfirmEthereumTransaction(transaction.Hash().String(), 5, time.Second*5)
		if err != nil {
			fmt.Println(err)
		}

		if status == 1 {
			fmt.Printf("Authorizer has been added: %s\n", wallet.Address.String())
		} else {
			fmt.Printf("Authorizer has failed to be added: %s\n", wallet.Address.String())
		}
	}
}
