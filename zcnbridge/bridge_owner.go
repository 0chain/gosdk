package zcnbridge

import (
	"context"
	"fmt"
	eth "github.com/ethereum/go-ethereum"
	"time"

	"github.com/0chain/gosdk/zcnbridge/ethereum/authorizers"
	hdw "github.com/0chain/gosdk/zcncore/ethhdwallet"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func (b *BridgeOwner) prepareAuthorizers(ctx context.Context, method string, params ...interface{}) (*authorizers.Authorizers, *bind.TransactOpts, error) {
	etherClient, err := b.CreateEthClient()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create etherClient")
	}

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
	gasLimitUnits, err := etherClient.EstimateGas(ctx, eth.CallMsg{
		To:   &contractAddress,
		From: from,
		Data: pack,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to estimate gas")
	}

	// Update gas limits + 10%
	gasLimitUnits = addPercents(gasLimitUnits, 10).Uint64()

	transactOpts := b.CreateSignedTransactionFromKeyStore(etherClient, gasLimitUnits)

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
		msg := "failed to execute AddAuthorizers transaction to ClientID = %s with amount = %s"
		return nil, errors.Wrapf(err, msg, b.ClientID(), address.String())
	}

	return tran, err
}

// RemoveEthereumAuthorizer Removes authorizer from Ethereum bridge. Only contract deployer can call this method
func (b *BridgeOwner) RemoveEthereumAuthorizer(ctx context.Context, address common.Address) (*types.Transaction, error) {
	instance, transactOpts, err := b.prepareAuthorizers(ctx, "removeAuthorizers", address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare bridge")
	}

	tran, err := instance.RemoveAuthorizers(transactOpts, address)
	if err != nil {
		msg := "failed to execute RemoveAuthorizers transaction to ClientID = %s with amount = %s"
		return nil, errors.Wrapf(err, msg, b.ClientID(), address.String())
	}

	return tran, err
}

func (b *BridgeOwner) AddEthereumAuthorizers(configDir string) {
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
