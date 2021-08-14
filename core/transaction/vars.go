package transaction

import (
	"errors"

	"github.com/0chain/gosdk/core/conf"
)

var (
	// Config
	cfg *conf.Config
)

// SetConfig set config variables for transaction
func SetConfig(c *conf.Config) {
	cfg = c
}

var (
	// ErrNoAvailableSharder no any available sharder
	ErrNoAvailableSharder = errors.New("[txn] there is no any available sharder")
	// ErrNoTxnDetail No transaction detail was found on any of the sharders
	ErrNoTxnDetail = errors.New("[txn] no transaction detail was found on any of the sharders")

	// ErrTooLessConfirmation too less sharder to confirm transaction
	ErrTooLessConfirmation = errors.New("[txn] too less sharders to confirm it")

	// ErrConfigIsNotInitialized config is not initialized
	ErrConfigIsNotInitialized = errors.New("[txn] config is not initialized. please initialize it by transaction.SetConfig")
)
