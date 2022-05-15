package znft

import (
	"context"

	binding "github.com/0chain/gosdk/zcnbridge/ethereum/bridge"
	eth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

func (conf *Configuration) createTransaction(ctx context.Context, method string, params ...interface{}) (*bind.TransactOpts, error) {
	etherClient, err := conf.CreateEthClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create etherClient")
	}

	// To (contract)
	contractAddress := common.HexToAddress(conf.FactoryModuleERC721Address)

	// Get ABI of the contract
	abi, err := binding.BridgeMetaData.GetAbi()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ABI")
	}

	// Pack the method argument
	pack, err := abi.Pack(method, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to pack arguments")
	}

	// Gas limits in units
	fromAddress := common.HexToAddress(conf.WalletAddress)

	gasLimitUnits, err := etherClient.EstimateGas(ctx, eth.CallMsg{
		To:   &contractAddress,
		From: fromAddress,
		Data: pack,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to estimate gas")
	}

	transactOpts := conf.CreateSignedTransactionFromKeyStore(ctx, gasLimitUnits)

	return transactOpts, nil
}
