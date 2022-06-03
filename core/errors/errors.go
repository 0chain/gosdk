package gosdkError

import "github.com/0chain/errors"

var (
	ErrTooManyRequests      = errors.New(TooManyRequests, "")
	ErrFileNotFound         = errors.New(FileNotFound, "")
	ErrInvalidReferencePath = errors.New(InvalidReferencePath, "")
	ErrMarshall             = errors.New(MarshallError, "")
	// Erasure Coding Errors
	ErrEC                   = errors.New(ECError, "")
	ErrECSplit              = errors.New(ECSplitError, "")
	ErrECVerify             = errors.New(ECVerifyError, "")
	ErrECReconstruct        = errors.New(ECReconstructError, "")
	ErrECInvalidInputLength = errors.New(ECInvalidInputLength, "")
	ErrECJoin               = errors.New(ECJoinError, "")
)
