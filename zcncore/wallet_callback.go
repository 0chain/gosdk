package zcncore

import (
	"sync"

	"github.com/0chain/gosdk/core/common"
)

type walletCallback struct {
	sync.WaitGroup
	success bool

	balance common.Balance
	info    string
	err     error
}

func (cb *walletCallback) OnBalanceAvailable(status int, value int64, info string) {
	defer cb.Done()

	if status == StatusSuccess {
		cb.success = true
	} else {
		cb.success = false
	}
	cb.info = info
	cb.balance = common.Balance(value)
}
