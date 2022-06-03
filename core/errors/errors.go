package gosdkError

import "github.com/0chain/errors"

var (
	ErrTooManyRequests      = errors.New(TooManyRequests, "")
	ErrFileNotFound         = errors.New(FileNotFound, "")
	ErrInvalidReferencePath = errors.New(InvalidReferencePath, "")
)
