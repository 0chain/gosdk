package magmasc

import (
	"github.com/0chain/gosdk/zmagmacore/errors"
)

const (
	errCodeBadRequest = "bad_request"
	errCodeInvalid    = "invalid_error"
)

var (
	// errDecodeData represents an error
	// that decode data was failed.
	errDecodeData = errors.New("decode_error", "decode error")

	// errInvalidSession represents an error
	// that a session was invalidated.
	errInvalidSession = errors.New(errCodeInvalid, "invalid session")

	// errInvalidConsumer represents an error
	// that consumer was invalidated.
	errInvalidConsumer = errors.New(errCodeInvalid, "invalid consumer")

	// errInvalidDataUsage represents an error
	// that a data usage was invalidated.
	errInvalidDataUsage = errors.New(errCodeInvalid, "invalid data usage")

	// errInvalidProvider represents an error
	// that provider was invalidated.
	errInvalidProvider = errors.New(errCodeInvalid, "invalid provider")

	// errInvalidTerms represents an error
	// that terms was invalidated.
	errInvalidTerms = errors.New(errCodeInvalid, "invalid terms")

	// errInvalidAccessPoint represents an error
	// that access point was invalidated.
	errInvalidAccessPoint = errors.New(errCodeInvalid, "invalid access point")

	// errInvalidUser represents an error
	// that user was invalidated.
	errInvalidUser = errors.New(errCodeInvalid, "invalid user")

	// errInvalidUserDataMarker represents an error
	// that user data marker was invalidated.
	errInvalidUserDataMarker = errors.New(errCodeInvalid, "invalid user data marker")
)
