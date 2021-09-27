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

	// errInvalidAcknowledgment represents an error
	// that an acknowledgment was invalidated.
	errInvalidAcknowledgment = errors.New(errCodeInvalid, "invalid acknowledgment")

	// errInvalidConsumer represents an error
	// that consumer was invalidated.
	errInvalidConsumer = errors.New(errCodeInvalid, "invalid consumer")

	// errInvalidDataUsage represents an error
	// that a data usage was invalidated.
	errInvalidDataUsage = errors.New(errCodeInvalid, "invalid data usage")

	// errInvalidProvider represents an error
	// that provider was invalidated.
	errInvalidProvider = errors.New(errCodeInvalid, "invalid provider")

	// errInvalidProviderTerms represents an error
	// that provider terms was invalidated.
	errInvalidProviderTerms = errors.New(errCodeInvalid, "invalid provider terms")
)
