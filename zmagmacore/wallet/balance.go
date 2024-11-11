// DEPRECATED: This package is deprecated and will be removed in a future release.
package wallet

import (
	"context"

	"github.com/0chain/gosdk/zcncore"
	"github.com/0chain/gosdk/zmagmacore/errors"
)

type (
	// getBalanceCallBack implements zcncore.GetBalanceCallback interface.
	getBalanceCallBack struct {
		balanceCh chan int64
		err       error
	}
)

// Balance responds balance of the wallet that used.
//
// NOTE: for using Balance you must set wallet info by running zcncore.SetWalletInfo.
func Balance(ctx context.Context) (int64, error) {
	cb := newGetBalanceCallBack()
	err := zcncore.GetBalance(cb)
	if err != nil {
		return 0, err
	}

	var b int64
	select {
	case <-ctx.Done():
		return 0, errors.New("get_balance", "context done is called")
	case b = <-cb.balanceCh:
		return b, cb.err
	}
}

// newGetBalanceCallBack creates initialized getBalanceCallBack.
func newGetBalanceCallBack() *getBalanceCallBack {
	return &getBalanceCallBack{
		balanceCh: make(chan int64),
	}
}

// OnBalanceAvailable implements zcncore.GetBalanceCallback interface.
func (b *getBalanceCallBack) OnBalanceAvailable(status int, value int64, _ string) {
	if status != zcncore.StatusSuccess {
		b.err = errors.New("get_balance", "failed respond balance")
	}

	b.balanceCh <- value
}
