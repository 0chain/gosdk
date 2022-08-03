package znft

import (
	"context"
	"math/big"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"

	factory "github.com/0chain/gosdk/znft/contracts/factory/binding"
)

type IFactory interface {
	Count() (*big.Int, error)
	GetToken(*big.Int) (common.Address, error)
}

type Factory struct {
	session *factory.BindingSession
	ctx     context.Context
}

func (f *Factory) Count() (*big.Int, error) {
	count, err := f.session.Count()
	if err != nil {
		err = errors.Wrapf(err, "failed to read %s", "Count")
		Logger.Error(err)
		return nil, err
	}

	return count, nil
}

func (f *Factory) GetToken(index *big.Int) (addr common.Address, err error) {
	addr, err = f.session.TokenList(index)
	if err != nil {
		err = errors.Wrapf(err, "failed to read token list %s at index %v", "TokenList", index.Int64())
		Logger.Error(err)
		return addr, err
	}

	return addr, nil
}
