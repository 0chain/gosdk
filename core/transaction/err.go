package transaction

import "errors"

var (
	// ErrNoAvailableSharder no any available sharder
	ErrNoAvailableSharder = errors.New("[sharder] there is no any available sharder")
	// ErrNoTxnDetail No transaction detail was found on any of the sharders
	ErrNoTxnDetail = errors.New("[sc] no transaction detail was found on any of the sharders")

	// ErrTooLessConfirmation too less sharder to confirm transaction
	ErrTooLessConfirmation = errors.New("[sc] too less sharders to confirm it")
)
