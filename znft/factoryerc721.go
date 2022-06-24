package znft

import (
	"context"
	"math/big"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"

	factory "github.com/0chain/gosdk/znft/contracts/factorymoduleerc721/binding"
)

// Solidity functions

//	function createToken(
// 		address owner,
// 		string calldata name,
// 		string calldata symbol,
// 		string calldata uri,
// 		uint256 max,
// 		uint256,
// 		uint256,
// 		bytes calldata
//) external returns (address) {

type IFactoryERC721 interface {
	CreateToken(owner, name, symbol, uri string, max *big.Int, data []byte) error
}

type FactoryERC721 struct {
	session *factory.BindingSession
	ctx     context.Context
}

func (s *FactoryERC721) CreateToken(owner, name, symbol, uri string, max *big.Int, data []byte) error {
	ownerAddress := common.HexToAddress(owner)
	evmTr, err := s.session.CreateToken(ownerAddress, name, symbol, uri, max, nil, nil, data)
	if err != nil {
		err = errors.Wrapf(err, "failed to execute %s", "CreateToken")
		Logger.Error(err)
		return err
	}

	Logger.Info("Executed CreateToken, hash: ", evmTr.Hash().Hex())

	return nil
}
