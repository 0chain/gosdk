package znft

import (
	"context"

	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

func (app *App) createTransactionWithGasPrice(ctx context.Context, address string, pack []byte) (*bind.TransactOpts, error) {
	gasLimitUnits, err := app.estimateGas(ctx, address, pack)
	if err != nil {
		return nil, err
	}

	transactOpts, err := app.createSignedTransactionFromKeyStoreWithGasPrice(gasLimitUnits)

	return transactOpts, err
}

func (app *App) createTransaction() (*bind.TransactOpts, error) {
	transactOpts, err := app.createSignedTransactionFromKeyStore()

	return transactOpts, err
}

func (app *App) estimateGas(ctx context.Context, address string, pack []byte) (uint64, error) {
	etherClient, err := CreateEthClient(app.cfg.EthereumNodeURL)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create etherClient")
	}

	// To (contract)
	contractAddress := common.HexToAddress(address)

	// Gas limits in units
	fromAddress := common.HexToAddress(app.cfg.WalletAddress)

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
