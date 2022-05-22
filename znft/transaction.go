package znft

import (
	"context"

	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

func (conf *Configuration) createTransactionWithGasPrice(ctx context.Context, pack []byte) (*bind.TransactOpts, error) {
	gasLimitUnits, err := conf.estimateGas(ctx, pack)
	if err != nil {
		return nil, err
	}

	transactOpts := conf.createSignedTransactionFromKeyStoreWithGasPrice(gasLimitUnits)

	return transactOpts, nil
}

func (conf *Configuration) createTransaction() (*bind.TransactOpts, error) {
	transactOpts := conf.createSignedTransactionFromKeyStore()

	return transactOpts, nil
}

func (conf *Configuration) estimateGas(ctx context.Context, pack []byte) (uint64, error) {
	etherClient, err := conf.CreateEthClient()
	if err != nil {
		return -1, errors.Wrap(err, "failed to create etherClient")
	}

	// To (contract)
	contractAddress := common.HexToAddress(conf.FactoryModuleERC721Address)

	// Gas limits in units
	fromAddress := common.HexToAddress(conf.WalletAddress)

	// Estimate gas
	gasLimitUnits, err := etherClient.EstimateGas(ctx, eth.CallMsg{
		To:   &contractAddress,
		From: fromAddress,
		Data: pack,
	})
	if err != nil {
		return 0, errors.Wrap(err, "failed to estimate gas")
	}

	return gasLimitUnits, nil
}
