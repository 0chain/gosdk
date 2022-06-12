package znft

import (
	"context"
	"math/big"

	factory "github.com/0chain/gosdk/znft/contracts/factory/binding"
)

// Solidity functions

// function create(
//	address module,
//	string calldata name,
//	string calldata symbol,
//	string calldata uri,
//	uint256 max,
//	uint256 price,
//	uint256 batch,
//	bytes calldata data) external returns (address) {

// function register(address module, bool status)

type IFactoryRandom interface {
	Create(module, name, symbol, uri string, max, price, batch *big.Int, data string) string
	Register(address string, status bool)
}

type FactoryRandom struct {
	session *factory.BindingSession
	ctx     context.Context
}
