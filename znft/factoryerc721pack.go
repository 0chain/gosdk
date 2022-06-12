package znft

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	factory "github.com/0chain/gosdk/znft/contracts/factorymoduleerc721pack/binding"
)

// Solidity functions

// function createToken(
//  address owner,
//  string calldata name,
//  string calldata symbol,
//  string calldata uri,
//  uint256 max,
//  uint256 price,
//  uint256 batch,
//  bytes calldata
//) external returns (address) {

type IFactoryPack interface {
	Create(owner, name, symbol, uri string, max, price, batch *big.Int, calldata string) error
}

type FactoryPack struct {
	session *factory.BindingSession
	ctx     context.Context
}

func (s *FactoryPack) CreateToken(owner, name, symbol, uri string, max, price, batch *big.Int, data []byte) error {
	ownerAddress := common.HexToAddress(owner)
	evmTr, err := s.session.CreateToken(ownerAddress, name, symbol, uri, max, price, batch, data)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", "CreateToken")
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed CreateToken, hash: ", evmTr.Hash().Hex())

	return nil
}
