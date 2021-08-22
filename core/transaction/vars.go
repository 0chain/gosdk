package transaction

import (
	"errors"
)

var (
	// ErrInvalidRequest invalid request
	ErrInvalidRequest = errors.New("[txn] invalid request")

	// ErrNoAvailableSharder no any available sharder
	ErrNoAvailableSharder = errors.New("[txn] there is no any available sharder")

	// ErrNoTxnDetail No transaction detail was found on any of the sharders
	ErrNoTxnDetail = errors.New("[txn] no transaction detail was found on any of the sharders")

	// ErrTooLessConfirmation too less sharder to confirm transaction
	ErrTooLessConfirmation = errors.New("[txn] too less sharders to confirm it")
)
