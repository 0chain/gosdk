package transaction

import (
	"context"

	"github.com/0chain/gosdk/zcnbridge/errors"
	"github.com/0chain/gosdk/zcncore"
)

var (
	// Ensure callback implements interface.
	_ zcncore.TransactionCallback = (*callback)(nil)
)

type (
	// callback implements zcncore.TransactionCallback interface.
	callback struct {
		// waitCh represents channel for making callback.OnTransactionComplete,
		// callback.OnVerifyComplete and callBack.OnAuthComplete operations async.
		waitCh chan interface{}
		err    error
	}
)

func NewStatus() zcncore.TransactionCallback {
	return &callback{
		waitCh: make(chan interface{}),
	}
}

// OnTransactionComplete implements zcncore.TransactionCallback interface.
func (cb *callback) OnTransactionComplete(zcnTxn *zcncore.Transaction, status int) {
	if status != zcncore.StatusSuccess {
		msg := "status is not success: " + TxnStatus(status).String() + "; err: " + zcnTxn.GetTransactionError()
		cb.err = errors.New("on_transaction_complete", msg)
	}

	cb.sendCall()
}

// OnVerifyComplete implements zcncore.TransactionCallback interface.
func (cb *callback) OnVerifyComplete(zcnTxn *zcncore.Transaction, status int) {
	if status != zcncore.StatusSuccess {
		msg := "status is not success: " + TxnStatus(status).String() + "; err: " + zcnTxn.GetVerifyError()
		cb.err = errors.New("on_transaction_verify", msg)
	}

	cb.sendCall()
}

// OnAuthComplete implements zcncore.TransactionCallback interface.
func (cb *callback) OnAuthComplete(_ *zcncore.Transaction, status int) {
	if status != zcncore.StatusSuccess {
		msg := "status is not success: " + TxnStatus(status).String()
		cb.err = errors.New("on_transaction_verify", msg)
	}

	cb.sendCall()
}

func (cb *callback) waitCompleteCall(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return errors.New("completing_transaction", "completing transaction context deadline exceeded")
	case <-cb.waitCh:
		return cb.err
	}
}

func (cb *callback) waitVerifyCall(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return errors.New("verifying_transaction", "verifying transaction context deadline exceeded")
	case <-cb.waitCh:
		return cb.err
	}
}

func (cb *callback) sendCall() {
	cb.waitCh <- true
}
