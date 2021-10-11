package magmasc

import (
	"github.com/0chain/gosdk/zmagmacore/errors"
)

const (
	ErrCodeBadRequest     = "bad_request"
	ErrCodeConsumerReg    = "consumer_reg"
	ErrCodeConsumerUpdate = "consumer_update"
	ErrCodeDataUsage      = "data_usage"
	ErrCodeDecode         = "decode_error"
	ErrCodeFetchData      = "fetch_data"
	ErrCodeInternal       = "internal_error"
	ErrCodeInvalid    = "invalid_error"
	ErrCodeProviderReg    = "provider_reg"
	ErrCodeProviderStake    = "provider_stake"
	ErrCodeProviderUnstake = "provider_unstake"
	ErrCodeProviderUpdate = "provider_update"
	ErrCodeSessionInit    = "session_init"
	ErrCodeSessionStart   = "session_start"
	ErrCodeSessionStop    = "session_stop"

	ErrCodeAccessPointReg    = "access_point_reg"
	ErrCodeAccessPointUpdate = "access_point_update"

	ErrCodeRewardPoolLock   = "reward_pool_lock"
	ErrCodeRewardPoolUnlock = "reward_pool_unlock"
	ErrCodeTokenPoolCreate  = "token_pool_create"
	ErrCodeTokenPoolSpend   = "token_pool_spend"

	ErrCodeUserUpdate = "user_update"

	ErrTextDecode     = "decode error"
	ErrTextUnexpected = "unexpected error"

	ErrCodeInvalidConfig = "invalid_config"
	ErrCodeInvalidFuncName = "invalid_func_name"
	ErrTextInvalidFuncName = "function with provided name is not supported"
)

var (
	// ErrInsufficientFunds represents an error that can occur while
	// check a balance value condition.
	ErrInsufficientFunds = errors.New(ErrCodeBadRequest, "insufficient funds")

	// ErrInternalUnexpected represents an error
	// that internal unexpected issue.
	ErrInternalUnexpected = errors.New(ErrCodeInternal, ErrTextUnexpected)

	// ErrInvalidAccessPoint represents an error
	// that access point was invalidated.
	ErrInvalidAccessPoint = errors.New(ErrCodeInvalid, "invalid access point")

	// ErrInvalidAccessPointID represents an error
	// that access point id was invalidated.
	ErrInvalidAccessPointID = errors.New(ErrCodeBadRequest, "invalid access_point_id")

	// ErrDecodeData represents an error
	// that decode data was failed.
	ErrDecodeData = errors.New("decode_error", "decode error")

	// ErrInvalidConsumer represents an error
	// that consumer was invalidated.
	ErrInvalidConsumer = errors.New(ErrCodeInvalid, "invalid consumer")

	// ErrInvalidConsumerExtID represents an error
	// that consumer external id was invalidated.
	ErrInvalidConsumerExtID = errors.New(ErrCodeBadRequest, "invalid consumer_ext_id")

	// ErrInvalidDataUsage represents an error
	// that a data usage was invalidated.
	ErrInvalidDataUsage = errors.New(ErrCodeInvalid, "invalid data usage")

	// ErrInvalidFuncName represents an error that can occur while
	// smart contract is calling with unsupported function name.
	ErrInvalidFuncName = errors.New(ErrCodeInvalidFuncName, ErrTextInvalidFuncName)

	// ErrInvalidProvider represents an error
	// that provider was invalidated.
	ErrInvalidProvider = errors.New(ErrCodeInvalid, "invalid provider")

	// ErrInvalidProviderExtID represents an error
	// that provider external id was invalidated.
	ErrInvalidProviderExtID = errors.New(ErrCodeBadRequest, "invalid provider_ext_id")

	// ErrInvalidSession represents an error
	// that a session was invalidated.
	ErrInvalidSession = errors.New(ErrCodeInvalid, "invalid session")

	// ErrInvalidTerms represents an error
	// that terms was invalidated.
	ErrInvalidTerms = errors.New(ErrCodeInvalid, "invalid terms")

	// ErrInvalidUser represents an error
	// that user was invalidated.
	ErrInvalidUser = errors.New(ErrCodeInvalid, "invalid user")

	// ErrInvalidDataMarker represents an error
	// that user data marker was invalidated.
	ErrInvalidDataMarker = errors.New(ErrCodeInvalid, "invalid data marker")

	// ErrNegativeValue represents an error that can occur while
	// a checked value is negative.
	ErrNegativeValue = errors.New(ErrCodeBadRequest, "negative value")

	// ErrNilPointerValue represents an error that can occur while
	// a checked value is a nil pointer.
	ErrNilPointerValue = errors.New(ErrCodeInternal, "nil pointer value")
)
